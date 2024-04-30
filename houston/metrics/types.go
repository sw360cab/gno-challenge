package metrics

type DeploymentUnit struct {
	Deployed int64
	Called   int64
}

type SlicedMap struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

type mapCoverter func(map[string]int64) []SlicedMap
