package covertree

import "math"

type coverSet struct {
	layers           []coverSetLayer
	totalItemCount   int
	visibleItemCount int
}

func coverSetWithItems(items []interface{}, parent interface{}, query interface{}, distanceFunc DistanceFunc, loadChildren func(...interface{}) ([]LevelsWithItems, error)) (coverSet, error) {
	var cs coverSet

	if len(items) > 0 {
		children, err := loadChildren(items...)
		if err != nil {
			return cs, err
		}

		itemsForLayer := make([]itemWithChildren, len(items))
		for i, item := range items {
			distance := distanceFunc(item, query)
			itemsForLayer[i] = itemWithChildren{withDistance: ItemWithDistance{item, distance}, parent: parent, children: children[i]}
		}

		cs.addLayer(makeCoverSetLayer(itemsForLayer))
	}

	return cs, nil
}

func (cs *coverSet) addLayer(layer coverSetLayer) {
	cs.layers = append(cs.layers, layer)
	cs.totalItemCount += len(layer)
	cs.visibleItemCount += len(layer)
}

func (cs coverSet) atBottom() bool {
	for _, layer := range cs.layers {
		for _, csItem := range layer {
			if csItem.hasChildren() {
				return false
			}
		}
	}
	return true
}

func (cs coverSet) child(query interface{}, distThreshold float64, childLevel int, distanceBetween DistanceFunc, loadChildren func(...interface{}) ([]LevelsWithItems, error)) (childCoverSet coverSet, parentWithinThreshold interface{}, err error) {
	childCoverSet = coverSet{
		layers:           cs.layers,
		totalItemCount:   cs.totalItemCount,
		visibleItemCount: 0,
	}

	var promotedChildren []itemWithChildren
	var minParentDistance = math.MaxFloat64

	for i := range cs.layers {
		layer := cs.layers[i].constrainedToDistance(distThreshold)
		childCoverSet.layers[i] = layer
		childCoverSet.visibleItemCount += len(layer)

		if len(layer) > 0 && layer[0].withDistance.Distance < minParentDistance {
			parentWithinThreshold = layer[0].withDistance.Item
			minParentDistance = layer[0].withDistance.Distance
		}

		for _, csItem := range layer {
			for _, childItem := range csItem.takeChildrenAt(childLevel) {
				if childDist := distanceBetween(childItem, query); childDist <= distThreshold {
					promotedChild := itemWithChildren{withDistance: ItemWithDistance{childItem, childDist}, parent: csItem.withDistance.Item}
					promotedChildren = append(promotedChildren, promotedChild)
				}
			}
		}
	}

	if len(promotedChildren) > 0 {
		children := make([]interface{}, len(promotedChildren))
		for i := range promotedChildren {
			children[i] = promotedChildren[i].withDistance.Item
		}

		grandchildren, err := loadChildren(children...)
		if err != nil {
			return childCoverSet, nil, err
		}

		for i := range promotedChildren {
			promotedChildren[i].children = grandchildren[i]
		}

		childCoverSet.addLayer(makeCoverSetLayer(promotedChildren))
	}

	return
}

func (cs coverSet) bound(maxItems int, maxDist float64) float64 {
	var count = 0
	var minIndices = make([]int, len(cs.layers))
	var boundDistance = maxDist

	for count < maxItems {

		var minLayerIndex = -1
		for layerIndex, layer := range cs.layers {
			minIndex := minIndices[layerIndex]
			if minIndex >= len(layer) {
				continue
			}

			itemDistance := layer[minIndex].withDistance.Distance
			if minLayerIndex == -1 || itemDistance < boundDistance {
				minLayerIndex = layerIndex
				boundDistance = itemDistance
			}
		}

		if minLayerIndex == -1 || boundDistance > maxDist {
			break
		}

		count++
		minIndices[minLayerIndex]++
	}

	if count == maxItems {
		return boundDistance
	}

	return maxDist
}

func (cs coverSet) closest(maxItems int, maxDist float64) []ItemWithDistance {
	var results []ItemWithDistance
	var minIndices = make([]int, len(cs.layers))

	for len(results) < maxItems {

		var minItem *ItemWithDistance
		var minLayerIndex = -1
		for layerIndex, layer := range cs.layers {
			minIndex := minIndices[layerIndex]
			if minIndex >= len(layer) {
				continue
			}

			item := &layer[minIndex].withDistance
			if minLayerIndex == -1 || item.Distance < minItem.Distance {
				minLayerIndex = layerIndex
				minItem = item
			}
		}

		if minLayerIndex == -1 || minItem.Distance > maxDist {
			break
		}

		results = append(results, *minItem)
		minIndices[minLayerIndex]++
	}

	return results
}
