package metrics

type DeploymentUnit struct {
	Deployed int64
	Called   int64
}

type SlicedMap struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

type DeploymentUnitFilter interface {
	isValidValue(DeploymentUnit) bool
	getField(DeploymentUnit) int64
}

type CalledUnitFilter struct{}

func (cu *CalledUnitFilter) isValidValue(d DeploymentUnit) bool { return d.Called > 0 }
func (cu *CalledUnitFilter) getField(d DeploymentUnit) int64    { return d.Called }

type DeployedUnitFilter struct{}

func (du *DeployedUnitFilter) isValidValue(d DeploymentUnit) bool { return d.Deployed > 0 }
func (du *DeployedUnitFilter) getField(d DeploymentUnit) int64    { return d.Deployed }
