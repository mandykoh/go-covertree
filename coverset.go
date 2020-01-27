package covertree

import "math"

type coverSet struct {
	layers    []coverSetLayer
	itemCount int
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
	cs.itemCount += len(layer)
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
		layers:    cs.layers,
		itemCount: 0,
	}

	var promotedChildren []itemWithChildren
	var promotedChildItems []interface{}
	var parentDistance = math.MaxFloat64

	for i := range cs.layers {
		layer := cs.layers[i].constrainedToDistance(distThreshold)
		childCoverSet.layers[i] = layer
		childCoverSet.itemCount += len(layer)

		if len(layer) > 0 && layer[0].withDistance.Distance < parentDistance {
			parentWithinThreshold = layer[0].withDistance.Item
			parentDistance = layer[0].withDistance.Distance
		}

		for _, csItem := range layer {
			for _, childItem := range csItem.takeChildrenAt(childLevel) {
				if childDist := distanceBetween(childItem, query); childDist <= distThreshold {
					promotedChild := itemWithChildren{withDistance: ItemWithDistance{childItem, childDist}, parent: csItem.withDistance.Item}
					promotedChildren = append(promotedChildren, promotedChild)
					promotedChildItems = append(promotedChildItems, childItem)
				}
			}
		}
	}

	if len(promotedChildItems) > 0 {
		children, err := loadChildren(promotedChildItems...)
		if err != nil {
			return childCoverSet, nil, err
		}

		for i := range promotedChildren {
			promotedChildren[i].children = children[i]
		}

		childCoverSet.addLayer(makeCoverSetLayer(promotedChildren))
	}

	return
}

func (cs coverSet) closest(maxItems int, maxDist float64) []ItemWithDistance {
	var results []ItemWithDistance
	var minIndices = make([]int, len(cs.layers))

	for len(results) < maxItems {

		var minItem ItemWithDistance
		var minLayerIndex = -1
		for layerIndex, layer := range cs.layers {
			minIndex := minIndices[layerIndex]
			if minIndex >= len(layer) {
				continue
			}

			item := layer[minIndex].withDistance
			if minLayerIndex == -1 || item.Distance < minItem.Distance {
				minLayerIndex = layerIndex
				minItem = item
			}
		}

		if minLayerIndex == -1 || minItem.Distance > maxDist {
			break
		}

		results = append(results, minItem)
		minIndices[minLayerIndex]++
	}

	return results
}
