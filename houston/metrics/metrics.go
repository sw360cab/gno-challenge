package metrics

import (
	"fmt"
	"regexp"
	"sort"
	"sync"
)

type TransactionMetricsCollector interface {
	// number of transactions
	GetTransactionCount() int64
	// percentage of success of transaction
	GetTransactionSuccessRate() float64
	// number of different message types
	GetMessageTypes() []SlicedMap
	// most active transaction senders
	GetTopTransactionSenders() []SlicedMap
	// most active realms deployed
	GetMostActiveRealmsDeployed() []SlicedMap
	// most active realms called
	GetMostActiveRealmsCalled() []SlicedMap
	// most active packages deployed
	GetMostActivePackagesDeployed() []SlicedMap
	// most active packages called
	GetMostActivePackagesCalled() []SlicedMap
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
	CountTx      int64
	SuccessTx    int64
	MessageTypes map[string]int64
	Senders      map[string]int64
	Realms       map[string]DeploymentUnit
	Packages     map[string]DeploymentUnit
	mu           sync.Mutex
}

var _ TransactionMetricsCollector = &TransactionMetric{}

func NewTransactionMetric() *TransactionMetric {
	return &TransactionMetric{
		CountTx:      0,
		MessageTypes: make(map[string]int64),
		Senders:      make(map[string]int64),
		Realms:       make(map[string]DeploymentUnit),
		Packages:     make(map[string]DeploymentUnit),
	}
}

func (tm *TransactionMetric) HandleTransactionMessage(transaction Transaction) error {
	if len(transaction.Messages) == 0 { // should never happen
		return fmt.Errorf("No message found in transaction")
	}
	msg := transaction.Messages[0]

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// update total
	tm.CountTx++
	// update success
	if transaction.Success {
		tm.SuccessTx++
	}
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
	case string(msgRun):
		// "typeUrl": "run"
		creator = msg.Value.MsgRun.Caller
	case string(bankMsgSend):
		// "typeUrl": "send" -> "route": "bank"
		creator = msg.Value.BankMsgSend.FromAddress
	}
	// update sender
	tm.Senders[creator] = tm.Senders[creator] + 1

	switch {
	case realmRegExp.MatchString(packagePath):
		actionMap = tm.Realms
	case packageRegExp.MatchString(packagePath):
		actionMap = tm.Packages
		// future cases
	default: // unhandled message
		return nil
	}

	currentActionUnit := actionMap[packagePath]
	if deployment {
		currentActionUnit.Deployed = currentActionUnit.Deployed + 1
	} else {
		currentActionUnit.Called = currentActionUnit.Called + 1
	}
	return nil
}

type SlicedMap struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

type mapCoverter func(map[string]int64) []SlicedMap

func defaultMapConverter(thisMap map[string]int64) []SlicedMap {
	var keyValued []SlicedMap
	for k, v := range thisMap {
		keyValued = append(keyValued, SlicedMap{k, v})
	}
	return keyValued
}

func (tm *TransactionMetric) sortKV(unorderedKv []SlicedMap) []SlicedMap {
	keyValued := unorderedKv
	sort.SliceStable(keyValued, func(i, j int) bool {
		return keyValued[i].Value > keyValued[j].Value
	})
	return keyValued
}

// Aggregation Methods
func (tm *TransactionMetric) GetTransactionCount() int64 {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.CountTx
}

func (tm *TransactionMetric) GetTransactionSuccessRate() float64 {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return float64((tm.SuccessTx * 100) / tm.CountTx)
}

// func (tm *TransactionMetric) GetTransactionOutcomeTypes() []SlicedMap {
// 	tm.mu.Lock()
// 	defer tm.mu.Unlock()
// 	kv := []SlicedMap{}
// 	kv = append(kv,
// 		SlicedMap{
// 			Key:   "success",
// 			Value: tm.SuccessTx,
// 		},
// 		SlicedMap{
// 			Key:   "failure",
// 			Value: tm.CountTx - tm.SuccessTx,
// 		})
// 	return kv
// }

func (tm *TransactionMetric) GetMessageTypes() []SlicedMap {
	// tm.mu.Lock()
	// defer tm.mu.Unlock()
	return tm.sortKV(defaultMapConverter(tm.MessageTypes))
}

func (tm *TransactionMetric) GetTopTransactionSenders() []SlicedMap {
	return tm.sortKV(defaultMapConverter(tm.Senders))
}

func (tm *TransactionMetric) GetMostActiveRealmsDeployed() []SlicedMap {
	var keyValued []SlicedMap
	for k, v := range tm.Realms {
		keyValued = append(keyValued, SlicedMap{k, int64(v.Deployed)})
	}
	return tm.sortKV(keyValued)
}

func (tm *TransactionMetric) GetMostActiveRealmsCalled() []SlicedMap {
	var keyValued []SlicedMap
	for k, v := range tm.Realms {
		keyValued = append(keyValued, SlicedMap{k, int64(v.Called)})
	}
	return tm.sortKV(keyValued)
}

func (tm *TransactionMetric) GetMostActivePackagesDeployed() []SlicedMap {
	var keyValued []SlicedMap
	for k, v := range tm.Packages {
		keyValued = append(keyValued, SlicedMap{k, int64(v.Deployed)})
	}
	return tm.sortKV(keyValued)
}

func (tm *TransactionMetric) GetMostActivePackagesCalled() []SlicedMap {
	var keyValued []SlicedMap
	for k, v := range tm.Packages {
		keyValued = append(keyValued, SlicedMap{k, int64(v.Called)})
	}
	return tm.sortKV(keyValued)
}
