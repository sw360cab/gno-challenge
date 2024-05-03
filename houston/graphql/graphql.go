package graphql

// The purpose of this package is to model GraphQL Transactions
// according to "https://github.com/gnolang/tx-indexer/tree/main/serve/graph/model"
// and to the schema types in `https://github.com/gnolang/tx-indexer/tree/main/serve/graph/schema`.
// The objects can be employed with the GQL client to perform queries or subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/hasura/go-graphql-client/pkg/jsonutil"
)

const (
	GRAPHQL_URL_ENV     string = "GRAPHQL_URL"
	DEFAULT_GRAPHQL_URL string = "http://127.0.0.1:8546/graphql/query"
)

type MsgValue string

const MsgAddPackage MsgValue = "MsgAddPackage"
const MsgCall MsgValue = "MsgCall"
const MsgRun MsgValue = "MsgRun"
const BankMsgSend MsgValue = "BankMsgSend"

type TransactionMessageHandler interface {
	HandleTransactionMessage(transaction Transaction) error
}

// // SumIntsOrFloats sums the values of map m. It supports both floats and integers
// // as map values.
// func SumIntsOrFloats[K comparable, V int64 | float64](m map[K]V) V {
// 	var s V
// 	for _, v := range m {
// 		s += v
// 	}
// 	return s
// }

type GraphQLClient struct {
	Endpoint                    string
	SubscriptionResponseHandler TransactionMessageHandler
}

// Performs a query toward GraphQL server given a filter and a query reference
// where the returned object will be
func (gqlClient *GraphQLClient) createGQLStaticQuery(filterMap map[string]interface{}, query interface{}) error {
	client := graphql.NewClient(gqlClient.Endpoint, nil)
	err := client.Query(context.Background(), query, filterMap)
	if err != nil {
		return err
	}
	return nil
}

// Handles the request of pre-existing blocks to the GraphQL server given a reference block
// and a bootstrap time, which defines a limit for the blocks to be queried.
func (gqlClient *GraphQLClient) QueryPreExistingBlocks(leftoverBlock LeftoversTransactionFilter, bootstapTime time.Time) ([]Transaction, error) {
	var transactions []Transaction = []Transaction{}
	var queryExistingBlocks *ExistingBlocksGraphQLQuery
	var queryLeftoverBlocks *LeftoversBlocksGraphQLQuery
	var queryLastBlockBeforeTime *LastBlockBeforeTimeQuery

	if leftoverBlock.ToBlock == 0 && leftoverBlock.ToIndex == 0 { // no blocks left to Query
		return transactions, nil
	}

	if leftoverBlock.ToBlock > 0 {
		queryExistingBlocks = &ExistingBlocksGraphQLQuery{}
		err := gqlClient.createGQLStaticQuery(map[string]interface{}{
			"toBlock": leftoverBlock.ToBlock - 1,
		}, queryExistingBlocks)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, queryExistingBlocks.Transactions...)
	}

	// NOTE: this is a fine tuning since GraphQL subscription is not returning
	// transaction in order. This means that a transaction from a block may
	// finish before another with a lower index
	queryLastBlockBeforeTime = &LastBlockBeforeTimeQuery{}
	err := gqlClient.createGQLStaticQuery(map[string]interface{}{
		"fromHeight": leftoverBlock.ToBlock,
		"toHeight":   leftoverBlock.ToBlock + 1, // exclusive interval
		"toTime":     bootstapTime,
	}, queryLastBlockBeforeTime)
	if err != nil {
		return nil, err
	}

	// if last index is 0 it must not query because it will query twice
	// an incoming block handled by the subscription itself
	if leftoverBlock.ToIndex > 0 && len(queryLastBlockBeforeTime.Blocks) > 1 {
		lastValidBlock := queryLastBlockBeforeTime.Blocks[len(queryLastBlockBeforeTime.Blocks)-1]
		queryLeftoverBlocks = &LeftoversBlocksGraphQLQuery{}
		err := gqlClient.createGQLStaticQuery(map[string]interface{}{
			"toBlock": lastValidBlock,
			"toIndex": leftoverBlock.ToIndex,
		}, queryLeftoverBlocks)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, queryLeftoverBlocks.Transactions...)
	}
	return transactions, nil
}

// Prepares and sends a subscription request to the GraphQL server.
// It binds the callback in charge of handling incoming data.
func (gqlClient *GraphQLClient) CreateGQLSubscription() error {
	var subscriptionRequest SubscriptionGraphQLQuery
	client := graphql.NewSubscriptionClient(gqlClient.Endpoint).WithRetryTimeout(5 * time.Minute)
	defer client.Close()

	// Performs the subscription with a given callback in charge of handling received data
	subscriptionId, err := client.Subscribe(&subscriptionRequest, nil, func(dataValue []byte, errValue error) error {
		// If the call return error, onError event will be triggered
		// The function returns subscription ID and error. You can use subscription ID to unsubscribe the subscription
		if errValue != nil {
			return fmt.Errorf("Problem from subscription callback: %w", errValue)
		}
		data := SubscriptionResponse{}
		err := jsonutil.UnmarshalGraphQL(dataValue, &data)
		if err != nil {
			return fmt.Errorf("Problem receiving obj : %w", err)
		}

		// call the data handler
		gqlClient.SubscriptionResponseHandler.HandleTransactionMessage(data.Transactions)
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("Connecting to GraphQL server with Subscription id: %s ...\n", subscriptionId)
	err = client.Run()
	if err != nil {
		return err
	}
	return nil
}
