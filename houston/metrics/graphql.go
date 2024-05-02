package metrics

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

const msgAddPackage MsgValue = "MsgAddPackage"
const msgCall MsgValue = "MsgCall"
const msgRun MsgValue = "MsgRun"
const bankMsgSend MsgValue = "BankMsgSend"

func (gqlClient *GraphQLClient) createGQLStaticQuery(existingBlock map[string]interface{}, query interface{}) error {
	client := graphql.NewClient(gqlClient.Endpoint, nil)
	err := client.Query(context.Background(), query, existingBlock)
	if err != nil {
		return err
	}
	return nil
}

func (gqlClient *GraphQLClient) CreateGQLStaticQuery(leftoverBlock LeftoversTransactionFilter, bootstapTime time.Time) ([]Transaction, error) {
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

func (gqlClient *GraphQLClient) CreateGQLSubscription() error {
	var subscriptionRequest SubscriptionGraphQLQuery
	client := graphql.NewSubscriptionClient(gqlClient.Endpoint).WithRetryTimeout(5 * time.Minute)
	defer client.Close()

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
