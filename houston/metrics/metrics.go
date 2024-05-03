package metrics

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"go.uber.org/zap"
)

const LatestBlockHeightRedisKey string = "LATEST_BLOCK_HEIGHT"

var (
	realmRegExp   = regexp.MustCompile(`gno\.land\/r\/.*`)
	packageRegExp = regexp.MustCompile(`gno\.land\/p\/.*`)
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

type TransactionMetric struct {
	CountTx       int64
	SuccessTx     int64
	MessageTypes  map[string]int64
	Senders       map[string]int64
	Realms        map[string]DeploymentUnit
	Packages      map[string]DeploymentUnit
	ctx           context.Context
	LatestBlockCh chan LeftoversTransactionFilter
	mu            sync.Mutex
	logger        *zap.Logger
}

var _ TransactionMetricsCollector = &TransactionMetric{}

func NewTransactionMetric(logger *zap.Logger) *TransactionMetric {
	return &TransactionMetric{
		MessageTypes:  make(map[string]int64),
		Senders:       make(map[string]int64),
		Realms:        make(map[string]DeploymentUnit),
		Packages:      make(map[string]DeploymentUnit),
		LatestBlockCh: make(chan LeftoversTransactionFilter),
		ctx:           context.Background(),
		logger:        logger,
	}
}

func (tm *TransactionMetric) HandleTransactionMessage(transaction Transaction) error {
	if len(transaction.Messages) == 0 { // this should never happen
		return fmt.Errorf("No message found in transaction")
	}
	msg := transaction.Messages[0]

	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.logger.Debug("Transaction received", zap.String("transaction", fmt.Sprintf("%v", transaction)))
	// handle first transaction received
	// check pre-existing blocks not handled yet
	if tm.CountTx == 0 {
		tm.LatestBlockCh <- LeftoversTransactionFilter{
			ToBlock: transaction.BlockHeight,
			ToIndex: transaction.Index,
		}
	}
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

	if _, ok := actionMap[packagePath]; !ok {
		actionMap[packagePath] = DeploymentUnit{}
	}
	currentActionUnit := actionMap[packagePath]
	if deployment {
		actionMap[packagePath] = DeploymentUnit{
			Deployed: currentActionUnit.Deployed + 1,
			Called:   currentActionUnit.Called,
		}
	} else {
		actionMap[packagePath] = DeploymentUnit{
			Deployed: currentActionUnit.Deployed,
			Called:   currentActionUnit.Called + 1,
		}
	}
	return nil
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
	if tm.CountTx == 0 {
		return 0
	}
	return float64((tm.SuccessTx * 100) / tm.CountTx)
}

func (tm *TransactionMetric) GetMessageTypes() []SlicedMap {
	// tm.mu.Lock()
	// defer tm.mu.Unlock()
	return SortKV(DefaultSlicedMapConverter(tm.MessageTypes))
}

func (tm *TransactionMetric) GetTopTransactionSenders() []SlicedMap {
	return SortKV(DefaultSlicedMapConverter(tm.Senders))
}

func (tm *TransactionMetric) GetMostActiveRealmsDeployed() []SlicedMap {
	return SortKV(DeployedItemsSlicedMapConverter(tm.Realms))
}

func (tm *TransactionMetric) GetMostActiveRealmsCalled() []SlicedMap {
	return SortKV(CalledItemsSlicedMapConverter(tm.Realms))
}

func (tm *TransactionMetric) GetMostActivePackagesDeployed() []SlicedMap {
	return SortKV(DeployedItemsSlicedMapConverter(tm.Packages))
}

func (tm *TransactionMetric) GetMostActivePackagesCalled() []SlicedMap {
	return SortKV(CalledItemsSlicedMapConverter(tm.Packages))
}
