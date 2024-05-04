package metrics

import (
	"sort"
)

// Sorts a slice of SlicedMaps elements in DESCENDING order.
// Items having equals Value field are sorted among them by Key field
func SortKV(unorderedKv []SlicedMap) []SlicedMap {
	keyValued := unorderedKv
	sort.SliceStable(keyValued, func(i, j int) bool {
		if keyValued[i].Value == keyValued[j].Value {
			// handle same value conflict using key in alphanum order
			return keyValued[i].Key < keyValued[j].Key
		}
		return keyValued[i].Value > keyValued[j].Value
	})
	return keyValued
}

// Converts an hash map into a slice of SlicedMap elements
func DefaultSlicedMapConverter(thisMap map[string]int64) []SlicedMap {
	var keyValued []SlicedMap = []SlicedMap{}
	for k, v := range thisMap {
		keyValued = append(keyValued, SlicedMap{
			Key:   k,
			Value: v,
		})
	}
	return keyValued
}

// Converts a map into a slice of SlicedMap elements filtering fields according to predicate
func slicedMapConverter(unitMap map[string]DeploymentUnit, filter DeploymentUnitFilter) []SlicedMap {
	var keyValued []SlicedMap = []SlicedMap{}
	for k, v := range unitMap {
		if filter.isValidValue(v) {
			keyValued = append(keyValued, SlicedMap{k, filter.getField(v)})
		}
	}
	return keyValued
}

// Converts an hash map of `deployed` items into a slice of SlicedMap elements
func DeployedItemsSlicedMapConverter(deployedMap map[string]DeploymentUnit) []SlicedMap {
	return slicedMapConverter(deployedMap, &DeployedUnitFilter{})
}

// Converts an hash map of `called` items into a slice of SlicedMap elements
func CalledItemsSlicedMapConverter(calledMap map[string]DeploymentUnit) []SlicedMap {
	return slicedMapConverter(calledMap, &CalledUnitFilter{})
}
