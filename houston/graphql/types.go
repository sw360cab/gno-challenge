package graphql

import "time"

/********************/
/* GraphQL modeling */
/********************/
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
		MsgRun struct {
			Caller string
		} `graphql:"... on MsgRun"`
		BankMsgSend struct {
			FromAddress string `graphql:"from_address"`
		} `graphql:"... on BankMsgSend"`
	}
}

type Transaction struct {
	Messages    []Message
	BlockHeight int64 `graphql:"block_height"`
	Index       int64
	Hash        string
	Success     bool
}

type Block struct {
	ChainId string    `json:"-" graphql:"-"` // for future implementations
	Height  int       `json:"height,omitempty"`
	Time    time.Time `json:"time,omitempty"`
}

type SubscriptionResponse struct {
	Transactions Transaction
}

/*************************/
/* GraphQL query results */
/*************************/
type SubscriptionGraphQLQuery struct {
	Transactions []Transaction `graphql:"transactions(filter: { message : { route: vm } })"` // ignore bank
}

type ExistingBlocksGraphQLQuery struct {
	Transactions []Transaction `graphql:"transactions(filter: { message : { route: vm }, to_block_height: $toBlock })"`
}

type LeftoversBlocksGraphQLQuery struct {
	Transactions []Transaction `graphql:"transactions(filter: { message : { route: vm }, from_block_height: $toBlock, to_block_height: $toBlock, to_index: $toIndex })"`
}

type LastBlockBeforeTimeQuery struct {
	Blocks []Block `graphql:"blocks(filter: { from_height: $fromHeight, to_height: $toHeight, to_time: $toTime })"`
}

type BlocksBeforeTimeQuery struct {
	Blocks []Block `graphql:"blocks(filter: { to_time: $toTime })"`
}

/*******************/
/* GraphQL filters */
/*******************/
type ExistingTransactionFilter struct {
	ToBlock int64
}
type LeftoversTransactionFilter struct {
	ToBlock int64
	ToIndex int64
}

type LastBlockBeforeTimeFilter struct {
	FromHeight int64
	ToHeight   int64
	ToTime     time.Time
}

// Sample query in pure GraphQL
//
// {
//   transactions(filter: {message: {route: vm}}) {
//     success
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
