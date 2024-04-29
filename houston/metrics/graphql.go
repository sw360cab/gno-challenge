package metrics

// The purpose of this package is to model GraphQL Transactions
// according to "github.com/gnolang/tx-indexer/serve/graph/model"
// and to the schema types in `https://github.com/gnolang/tx-indexer/tree/main/serve/graph/schema`.
// The objects can be employed with the GQL client to perform queries or subscriptions

import (
	"context"
	"fmt"

	"github.com/hasura/go-graphql-client"
	"github.com/hasura/go-graphql-client/pkg/jsonutil"
)

const GRAPHQL_URL_ENV = "GRAPHQL_URL"
const DEFAULT_GRAPHQL_URL string = "http://127.0.0.1:8546/graphql/query"

// {
//   transactions(filter: {message: {route: vm}}) {
//     messages {
//       route
//       typeUrl
//       value {
//         __typename
//         ... on MsgAddPackage {
//           creator
//           package {
//             name
//             path
//           }
//         }
//         ... on MsgCall {
//           caller
//           pkg_path
//           func
//         }
//       }
//     }
//   }
// }

type MsgType string

const msgAddPackage MsgType = "MsgAddPackage"
const msgCall MsgType = "MsgCall"

type Message struct {
	Route   string
	TypeUrl string
	Value   struct {
		TypeName      string `graphql:"__typename"`
		MsgAddPackage struct {
			Creator string
			Package struct {
				Name string
				Path string
			}
		} `graphql:"... on MsgAddPackage"`
		MsgCall struct {
			Caller  string
			PkgPath string `graphql:"pkg_path"`
			Func    string
		} `graphql:"... on MsgCall"`
	}
}

type Transaction struct {
	Messages []Message
	Index    int
	Hash     string
}

type GraphQLQuery struct {
	Transactions []Transaction `graphql:"transactions(filter: {message: {route: vm}})"`
}

type SubscriptionResponse struct {
	Transactions Transaction
}

// GraphQL Client
type TransactionMessageHandler interface {
	HandleTransactionMessage(transaction Transaction) error
}

type GraphQLClient struct {
	Endpoint                    string
	SubscriptionResponseHandler TransactionMessageHandler
}

func (gqlClient *GraphQLClient) CreateGQLStaticQuery() error {
	var query GraphQLQuery
	client := graphql.NewClient(gqlClient.Endpoint, nil)
	err := client.Query(context.Background(), &query, nil)
	if err != nil {
		return err
	}
	fmt.Println(query.Transactions[0])
	return nil
}

func (gqlClient *GraphQLClient) CreateGQLSubscription() error {
	var subscriptionRequest GraphQLQuery
	client := graphql.NewSubscriptionClient(gqlClient.Endpoint)
	defer client.Close()

	subscriptionId, err := client.Subscribe(&subscriptionRequest, nil, func(dataValue []byte, errValue error) error {
		if errValue != nil {
			// if returns error, it will failback to `onError` event
			fmt.Println("Problem from subscription callback:" + errValue.Error())
			return nil
		}
		data := SubscriptionResponse{}
		err := jsonutil.UnmarshalGraphQL(dataValue, &data)
		if err != nil {
			fmt.Println("Problem receiving obj:" + err.Error())
		}

		gqlClient.SubscriptionResponseHandler.HandleTransactionMessage(data.Transactions)
		return nil
	})

	if err != nil {
		fmt.Println("Subscription:" + err.Error())
		return err
	}

	fmt.Printf("Subscription id %s:\n", subscriptionId)
	client.Run()
	return nil
}
