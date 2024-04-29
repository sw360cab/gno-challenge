package metrics

import (
	"fmt"
	"regexp"
	"sort"
)

type TransactionMetricsCollector interface {
	// number of transactions
	GetTransactionCount() (int64, error)
	// number of different message types
	GetMessageTypes() ([]KV, error)
	// most active transaction senders
	GetTopTransactionSenders() ([]KV, error)
	// most active realms deployed
	GetMostActiveRealmsDeployed() ([]KV, error)
	// most active realms called
	GetMostActiveRealmsCalled() ([]KV, error)
	// most active packages deployed
	GetMostActivePackagesDeployed() ([]KV, error)
	// most active packages called
	GetMostActivePackagesCalled() ([]KV, error)
}

var (
	realmRegExp   = regexp.MustCompile(`gno\.land\/r\/.*`)
	packageRegExp = regexp.MustCompile(`gno\.land\/p\/.*`)
)

type DeploymentUnit struct {
	Deployed int64
	Called   int64
}

type TransactionMetric struct {
	Count        int64
	MessageTypes map[string]int64
	Senders      map[string]int64
	Realms       map[string]DeploymentUnit
	Packages     map[string]DeploymentUnit
}

var _ TransactionMetricsCollector = TransactionMetric{}

func NewTransactionMetric() *TransactionMetric {
	return &TransactionMetric{
		Count:        0,
		MessageTypes: make(map[string]int64),
		Senders:      make(map[string]int64),
		Realms:       make(map[string]DeploymentUnit),
		Packages:     make(map[string]DeploymentUnit),
	}
}

func (tm TransactionMetric) HandleTransactionMessage(transaction Transaction) error {
	if len(transaction.Messages) == 0 { // should never happen
		return fmt.Errorf("No message found in transaction")
	}
	msg := transaction.Messages[0]

	// update total
	tm.Count++
	// update message types
	tm.MessageTypes[msg.Value.TypeName] = tm.MessageTypes[msg.Value.TypeName] + 1

	var creator, packagePath string
	var deployment bool
	var actionMap map[string]DeploymentUnit
	switch msg.Value.TypeName {
	case string(msgAddPackage):
		creator = msg.Value.MsgAddPackage.Creator
		packagePath = msg.Value.MsgAddPackage.Package.Path
		// "typeUrl": "add_package" -> Deployed
		deployment = true
	case string(msgCall):
		// "typeUrl": "exec" -> Called
		creator = msg.Value.MsgCall.Caller
		packagePath = msg.Value.MsgCall.PkgPath
	}

	// update sender
	tm.Senders[creator] = tm.Senders[creator] + 1

	switch {
	case realmRegExp.MatchString(packagePath):
		actionMap = tm.Realms
	case packageRegExp.MatchString(packagePath):
		actionMap = tm.Packages
		// future cases
	}

	currentActionUnit := actionMap[packagePath]
	if deployment {
		currentActionUnit.Deployed = currentActionUnit.Deployed + 1
	} else {
		currentActionUnit.Called = currentActionUnit.Called + 1
	}
	return nil
}

type KV struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

type mapCoverter func(map[string]int64) []KV

func defaultMapConverter(thisMap map[string]int64) []KV {
	var keyValued []KV
	for k, v := range thisMap {
		keyValued = append(keyValued, KV{k, v})
	}
	return keyValued
}

func (tm TransactionMetric) sortKV(unorderedKv []KV) []KV {
	keyValued := unorderedKv
	sort.SliceStable(keyValued, func(i, j int) bool {
		return keyValued[i].Value < keyValued[j].Value
	})
	return keyValued
}

// Aggregated Methods

func (tm TransactionMetric) GetTransactionCount() (int64, error) {
	return tm.Count, nil
}

func (tm TransactionMetric) GetMessageTypes() ([]KV, error) {
	return tm.sortKV(defaultMapConverter(tm.MessageTypes)), nil
}

func (tm TransactionMetric) GetTopTransactionSenders() ([]KV, error) {
	return tm.sortKV(defaultMapConverter(tm.Senders)), nil
}

func (tm TransactionMetric) GetMostActiveRealmsDeployed() ([]KV, error) {
	var keyValued []KV
	for k, v := range tm.Realms {
		keyValued = append(keyValued, KV{k, int64(v.Deployed)})
	}
	return tm.sortKV(keyValued), nil
}

func (tm TransactionMetric) GetMostActiveRealmsCalled() ([]KV, error) {
	var keyValued []KV
	for k, v := range tm.Realms {
		keyValued = append(keyValued, KV{k, int64(v.Called)})
	}
	return tm.sortKV(keyValued), nil
}

func (tm TransactionMetric) GetMostActivePackagesDeployed() ([]KV, error) {
	var keyValued []KV
	for k, v := range tm.Packages {
		keyValued = append(keyValued, KV{k, int64(v.Deployed)})
	}
	return tm.sortKV(keyValued), nil
}

func (tm TransactionMetric) GetMostActivePackagesCalled() ([]KV, error) {
	var keyValued []KV
	for k, v := range tm.Packages {
		keyValued = append(keyValued, KV{k, int64(v.Called)})
	}
	return tm.sortKV(keyValued), nil
}
