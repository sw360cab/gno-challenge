package metrics

import (
	"slices"
	"testing"

	"github.com/hasura/go-graphql-client/pkg/jsonutil"
	"github.com/stretchr/testify/assert"
	"github.com/sw360cab/gno-devops/graphql"
	"go.uber.org/zap"
)

func filterTx(txs []graphql.Transaction, filterFn func(graphql.Transaction) bool) []graphql.Transaction {
	dest := []graphql.Transaction{}
	for _, tx := range txs {
		if filterFn(tx) {
			dest = append(dest, tx)
		}
	}
	return dest
}

func mapTx(txs []graphql.Transaction, mapFn func(graphql.Transaction) string) []string {
	dest := []string{}
	for _, e := range txs {
		dest = append(dest, mapFn(e))
	}
	return dest
}

func TestMultipleTransactionHandler(t *testing.T) {
	t.Parallel()

	dataValue := `{
	"transactions": [
		{
			"block_height": 301,
			"index": 0,
			"success": true,
			"messages": [
				{
					"route": "vm",
					"typeUrl": "add_package",
					"value": {
						"__typename": "MsgAddPackage",
						"creator": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf4",
						"package": {
							"name": "hello",
							"path": "gno.land/r/demo/a/b/c/d/e/f/g"
						}
					}
				}
			]
		},
		{
			"block_height": 305,
			"index": 0,
			"success": true,
			"messages": [
				{
					"route": "vm",
					"typeUrl": "exec",
					"value": {
						"__typename": "MsgCall",
						"caller": "g14vhcdsyf83ngsrrqc92kmw8q9xakqjm0v8448t",
						"pkg_path": "gno.land/r1/demo/minter",
						"func": "Mint"
					}
				}
			]
		},
		{
			"block_height": 309,
			"index": 0,
			"messages": [
				{
					"route": "vm",
					"typeUrl": "add_package",
					"value": {
						"__typename": "MsgAddPackage",
						"creator": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
						"package": {
							"name": "hello",
							"path": "gno.land/p/demo/hello1"
						}
					}
				}
			]
		}
	]}`

	data := graphql.SubscriptionGraphQLQuery{}
	assert.NoError(t, jsonutil.UnmarshalGraphQL([]byte(dataValue), &data))

	txMetrics := NewTransactionMetric(zap.L())

	go func() {
		// unlock semaphore
		<-txMetrics.LatestBlockCh
	}()

	txs := data.Transactions
	for _, tx := range txs {
		txMetrics.HandleTransactionMessage(tx)
	}

	assert.Equal(t, int64(len(txs)), txMetrics.CountTx)
	// successful transactions
	successTxs := filterTx(txs, func(t graphql.Transaction) bool { return t.Success })
	assert.Equal(t, int64(len(successTxs)), txMetrics.SuccessTx)
	// senders
	senders := mapTx(txs, func(t graphql.Transaction) string {
		msg := t.Messages[0]
		switch msg.Value.TypeName {
		case string(graphql.MsgAddPackage):
			return msg.Value.MsgAddPackage.Creator
		case string(graphql.MsgCall):
			return msg.Value.MsgCall.Caller
		default:
			return ""
		}
	})
	assert.Equal(t, len(slices.Compact(senders)), len(txMetrics.Senders))

	msgTypes := mapTx(txs, func(t graphql.Transaction) string { return t.Messages[0].Value.TypeName })
	slices.Sort(msgTypes)
	assert.Equal(t, len(slices.Compact(msgTypes)), len(txMetrics.MessageTypes))

	realmsMapper := func(t graphql.Transaction) string {
		msg := t.Messages[0]
		switch msg.Value.TypeName {
		case string(graphql.MsgAddPackage):
			return msg.Value.MsgAddPackage.Package.Path
		case string(graphql.MsgCall):
			return msg.Value.MsgCall.PkgPath
		default:
			return ""
		}
	}
	realmsType := filterTx(txs, func(t graphql.Transaction) bool {
		return realmRegExp.MatchString(realmsMapper(t))
	})
	realms := mapTx(realmsType, realmsMapper)
	assert.Equal(t, len(slices.Compact(realms)), len(txMetrics.Realms))

	deployedRealms := filterTx(txs, func(t graphql.Transaction) bool {
		return realmRegExp.MatchString(realmsMapper(t)) && t.Messages[0].Value.MsgAddPackage.Package.Path != ""
	})

	assert.Equal(t, len(deployedRealms), len(func() (dest []string) {
		for k, v := range txMetrics.Realms {
			if v.Deployed > 0 {
				dest = append(dest, k)
			}
		}
		return dest
	}()))

}
