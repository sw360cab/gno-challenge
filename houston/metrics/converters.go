package metrics

import (
	"sort"
)

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

// func SortKVFN(unorderedKv []SlicedMap, mapCoverter mapCoverter) []SlicedMap {
// 	keyValued := unorderedKv
// 	sort.SliceStable(keyValued, func(i, j int) bool {
// 		if keyValued[i].Value == keyValued[j].Value {
// 			// handle same value conflict using key in alphanum order
// 			return keyValued[i].Key < keyValued[j].Key
// 		}
// 		return keyValued[i].Value > keyValued[j].Value
// 	})
// 	return keyValued
// }

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

// Converts an hash map of `deployed` items into a slice of SlicedMap elements
func DeployedItemsSlicedMapConverter(deployedMap map[string]DeploymentUnit) []SlicedMap {
	var keyValued []SlicedMap = []SlicedMap{}
	for k, v := range deployedMap {
		if v.Deployed > 0 {
			keyValued = append(keyValued, SlicedMap{k, int64(v.Deployed)})
		}
	}
	return keyValued
}

// Converts an hash map of `called` items into a slice of SlicedMap elements
func CalledItemsSlicedMapConverter(calledMap map[string]DeploymentUnit) []SlicedMap {
	var keyValued []SlicedMap = []SlicedMap{}
	for k, v := range calledMap {
		if v.Called > 0 {
			keyValued = append(keyValued, SlicedMap{k, int64(v.Called)})
		}
	}
	return keyValued
}
