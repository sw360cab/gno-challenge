package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortSlicedMapTest(t *testing.T) {
	t.Parallel()

	type sMapOrderTest struct {
		unordered []SlicedMap
		ordered   []SlicedMap
	}

	tt := []sMapOrderTest{
		{
			unordered: []SlicedMap{
				{Key: "zzz", Value: 2},
				{Key: "yyy", Value: 1},
				{Key: "xxx", Value: 3},
			},
			ordered: []SlicedMap{
				{Key: "xxx", Value: 3},
				{Key: "zzz", Value: 2},
				{Key: "yyy", Value: 1},
			},
		},
		{ // same value
			unordered: []SlicedMap{
				{Key: "zzz", Value: 1},
				{Key: "xxx", Value: 1},
				{Key: "yyy", Value: 1},
			},
			ordered: []SlicedMap{
				{Key: "xxx", Value: 1},
				{Key: "yyy", Value: 1},
				{Key: "zzz", Value: 1},
			},
		},
	}

	for _, tc := range tt {
		assert.Equal(t, tc.ordered, SortKV(tc.unordered))
	}
}

func TestToSliceMapConverter(t *testing.T) {
	t.Parallel()

	type convertableMapsTest struct {
		inputMap             map[string]DeploymentUnit
		calledConvertedMap   []SlicedMap
		deployedConvertedMap []SlicedMap
	}

	tt := []convertableMapsTest{
		{
			inputMap: map[string]DeploymentUnit{
				"zzz": {
					Called:   1,
					Deployed: 1,
				},
				"yyy": {
					Called:   10,
					Deployed: 40,
				},
				"xxx": {
					Called:   3,
					Deployed: 4,
				},
			},
			calledConvertedMap: []SlicedMap{
				{Key: "zzz", Value: 1},
				{Key: "yyy", Value: 10},
				{Key: "xxx", Value: 3},
			},
			deployedConvertedMap: []SlicedMap{
				{Key: "zzz", Value: 1},
				{Key: "yyy", Value: 40},
				{Key: "xxx", Value: 4},
			},
		},
	}

	for _, tc := range tt {
		assert.ElementsMatch(t, tc.deployedConvertedMap, DeployedItemsSlicedMapConverter(tc.inputMap))
		assert.ElementsMatch(t, tc.calledConvertedMap, CalledItemsSlicedMapConverter(tc.inputMap))
	}
}
