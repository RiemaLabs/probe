package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/RiemaLabs/probe/client"
	"github.com/RiemaLabs/probe/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	querier "github.com/RiemaLabs/probe/query"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func TestProbeIntegrationDevnet(t *testing.T) {
	os.Setenv("CHAIN_ID", "thunderbolt-devnet")
	os.Setenv("RPC_SERVER", "http://51.161.87.8:36657")

	logger.InitLogger("debug", false, os.Stdout)

	knownHeight := 600000

	cconfig := &client.ChainClientConfig{
		ChainID:               os.Getenv("CHAIN_ID"),
		RPCAddr:               os.Getenv("RPC_SERVER"),
		Debug:                 true,
		Timeout:               "30s",
		OutputFormat:          "json",
		Modules:               client.DefaultModuleBasics,
		CustomMsgTypeRegistry: client.DefaultCustomMsgTypeRegistry,
	}

	cl, err := client.NewChainClient(cconfig)
	require.NoError(t, err, "Failed to create chain client")
	assert.NotNil(t, cl, "Client should not be nil")

	t.Run("Test_ChainStatus", func(t *testing.T) {
		query := querier.Query{Client: cl, Options: &querier.QueryOptions{}}
		status, err := querier.StatusRPC(&query)

		require.NoError(t, err, "Failed to get chain status")
		assert.NotNil(t, status, "Status should not be nil")
		assert.NotEmpty(t, status.NodeInfo.Moniker, "Node moniker should not be empty")

		fmt.Printf("Chain Status Test Passed - Node: %s\n", status.NodeInfo.Moniker)
	})

	t.Run("Test_LatestBlock", func(t *testing.T) {
		query := querier.Query{Client: cl, Options: &querier.QueryOptions{}}
		block, err := querier.BlockRPC(&query)

		require.NoError(t, err, "Failed to get latest block")
		assert.NotNil(t, block, "Block should not be nil")
		assert.Greater(t, block.Block.Height, int64(0), "Block height should be positive")

		fmt.Printf("Latest Block Test Passed - Height: %d\n", block.Block.Height)
	})

	t.Run("Test_SpecificBlock", func(t *testing.T) {
		testHeight := int64(knownHeight)
		query := querier.Query{Client: cl, Options: &querier.QueryOptions{Height: testHeight}}
		block, err := querier.BlockRPC(&query)

		require.NoError(t, err, "Failed to get specific block")
		assert.NotNil(t, block, "Block should not be nil")
		assert.Equal(t, testHeight, block.Block.Height, "Block height should match requested height")

		fmt.Printf("Specific Block Test Passed - Height: %d\n", block.Block.Height)
	})

	t.Run("Test_BlockResults", func(t *testing.T) {
		testHeight := int64(knownHeight)
		query := querier.Query{Client: cl, Options: &querier.QueryOptions{Height: testHeight}}
		blockResults, err := querier.BlockResultsRPC(&query)

		require.NoError(t, err, "Failed to get block results")
		assert.NotNil(t, blockResults, "Block results should not be nil")
		assert.Equal(t, testHeight, blockResults.Height, "Block results height should match requested height")

		fmt.Printf("Block Results Test Passed - Height: %d\n", blockResults.Height)
	})

	t.Run("Test_TxsAtHeight", func(t *testing.T) {
		testHeight := int64(knownHeight)
		query := querier.Query{Client: cl, Options: &querier.QueryOptions{Height: testHeight}}

		txResponse, err := querier.TxsAtHeightRPC(&query, testHeight, cl.Codec)

		require.NoError(t, err, "Failed to get transactions")
		assert.NotNil(t, txResponse, "Transaction response should not be nil")

		if len(txResponse.Txs) > 0 {
			firstTxResponse := txResponse.TxResponses[0]
			assert.NotEmpty(t, firstTxResponse.TxHash, "Transaction hash should not be empty")
			assert.Equal(t, firstTxResponse.Height, testHeight, "Transaction height should match requested height")

			firstTx := txResponse.Txs[0]

			assert.NotEmpty(t, firstTx.Body.Messages, "Transaction messages should not be empty")

			assert.Equal(t, len(txResponse.Txs), len(txResponse.TxResponses), "Number of transactions should match number of transaction responses")
		} else {
			fmt.Println("Transactions Test Passed - No transactions found at this height")
		}
	})

	t.Run("Test_Txs", func(t *testing.T) {
		testHeight := int64(knownHeight)
		query := querier.Query{Client: cl, Options: &querier.QueryOptions{Height: testHeight}}
		txResponse, err := querier.TxsRPC(&query, testHeight, &txTypes.GetTxsEventRequest{OrderBy: txTypes.OrderBy_ORDER_BY_UNSPECIFIED, Page: 1, Limit: 100, Query: "tx.height=" + fmt.Sprintf("%d", testHeight)}, cl.Codec)
		require.NoError(t, err, "Failed to get transactions")
		assert.NotNil(t, txResponse, "Transaction response should not be nil")

		if len(txResponse.Txs) > 0 {
			firstTxResponse := txResponse.TxResponses[0]
			assert.NotEmpty(t, firstTxResponse.TxHash, "Transaction hash should not be empty")
			assert.Equal(t, firstTxResponse.Height, testHeight, "Transaction height should match requested height")

			firstTx := txResponse.Txs[0]

			assert.NotEmpty(t, firstTx.Body.Messages, "Transaction messages should not be empty")

			assert.Equal(t, len(txResponse.Txs), len(txResponse.TxResponses), "Number of transactions should match number of transaction responses")
		} else {
			fmt.Println("Transactions Response Test Passed - No transactions found at this height")
		}
	})

	t.Run("Test_TxsPagination", func(t *testing.T) {
		// 35 Txs in this block.
		testHeight := int64(444897)
		query := querier.Query{Client: cl, Options: &querier.QueryOptions{Height: testHeight}}
		txResponse0, err := querier.TxsRPC(&query, testHeight, &txTypes.GetTxsEventRequest{OrderBy: txTypes.OrderBy_ORDER_BY_UNSPECIFIED, Page: 1, Limit: 10, Query: "tx.height=" + fmt.Sprintf("%d", testHeight)}, cl.Codec)
		require.NoError(t, err, "Failed to get transactions")
		assert.NotNil(t, txResponse0, "Transaction response should not be nil")
		assert.Equal(t, 10, len(txResponse0.Txs), "First page should have 10 transactions")
		assert.Equal(t, uint64(35), txResponse0.Pagination.Total, "Total count should be 35")

		txResponse1, err := querier.TxsRPC(&query, testHeight, &txTypes.GetTxsEventRequest{OrderBy: txTypes.OrderBy_ORDER_BY_UNSPECIFIED, Page: 2, Limit: 10, Query: "tx.height=" + fmt.Sprintf("%d", testHeight)}, cl.Codec)
		require.NoError(t, err, "Failed to get transactions")
		assert.NotNil(t, txResponse1, "Transaction response should not be nil")
		assert.Equal(t, 10, len(txResponse1.Txs), "Second page should have 10 transactions")
		assert.Equal(t, uint64(35), txResponse1.Pagination.Total, "Total count should be 35")

		txResponse2, err := querier.TxsRPC(&query, testHeight, &txTypes.GetTxsEventRequest{OrderBy: txTypes.OrderBy_ORDER_BY_UNSPECIFIED, Page: 3, Limit: 10, Query: "tx.height=" + fmt.Sprintf("%d", testHeight)}, cl.Codec)
		require.NoError(t, err, "Failed to get transactions")
		assert.NotNil(t, txResponse2, "Transaction response should not be nil")
		assert.Equal(t, 10, len(txResponse2.Txs), "Third page should have 10 transactions")
		assert.Equal(t, uint64(35), txResponse2.Pagination.Total, "Total count should be 35")

		txResponse3, err := querier.TxsRPC(&query, testHeight, &txTypes.GetTxsEventRequest{OrderBy: txTypes.OrderBy_ORDER_BY_UNSPECIFIED, Page: 4, Limit: 10, Query: "tx.height=" + fmt.Sprintf("%d", testHeight)}, cl.Codec)
		require.NoError(t, err, "Failed to get transactions")
		assert.NotNil(t, txResponse3, "Transaction response should not be nil")
		assert.Equal(t, 5, len(txResponse3.Txs), "Fourth page should have 5 transactions")
		assert.Equal(t, uint64(35), txResponse3.Pagination.Total, "Total count should be 35")

		fmt.Printf("Transactions Pagination Test Passed - Found %d transactions\n", len(txResponse0.Txs))
	})

	t.Run("Test_TxsPagination", func(t *testing.T) {
		testHeight := int64(444897)
		query := querier.Query{Client: cl, Options: &querier.QueryOptions{Height: testHeight}}

		txResponse, err := querier.TxsAtHeightRPC(&query, testHeight, cl.Codec)

		require.NoError(t, err, "Failed to get transactions")
		assert.NotNil(t, txResponse, "Transaction response should not be nil")

		assert.Equal(t, uint64(35), txResponse.Total, "Total count should be 35")
		assert.Equal(t, int(35), len(txResponse.Txs), "Total txs count should be 35")
		assert.Equal(t, int(35), len(txResponse.TxResponses), "Total tx responses count should be 35")
	})

	t.Run("Test_TransactionsWasmExecute", func(t *testing.T) {
		testHeight := int64(401967)
		query := querier.Query{Client: cl, Options: &querier.QueryOptions{Height: testHeight}}
		queryString := "tx.height=" + fmt.Sprintf("%d", testHeight) + " AND message.action='/cosmwasm.wasm.v1.MsgExecuteContract'"

		txResponse, err := querier.TxsRPC(&query, testHeight, &txTypes.GetTxsEventRequest{OrderBy: txTypes.OrderBy_ORDER_BY_UNSPECIFIED, Page: 1, Limit: 100, Query: queryString}, cl.Codec)

		require.NoError(t, err, "Failed to get transactions")
		assert.NotNil(t, txResponse, "Transaction response should not be nil")

		// tx: 1ADCFD5F340236C1DFDC07B0E1BAB6ED4C3126D06904FFF56AE32A92DFBA868E
		assert.NotZero(t, txResponse.Pagination.Total, "Total count should be greater than 0")

		if len(txResponse.Txs) > 0 {
			firstTxResponse := txResponse.TxResponses[0]
			assert.NotEmpty(t, firstTxResponse.TxHash, "Transaction hash should not be empty")
			assert.Equal(t, firstTxResponse.Height, testHeight, "Transaction height should match requested height")

			firstTx := txResponse.Txs[0]

			assert.NotEmpty(t, firstTx.Body.Messages, "Transaction messages should not be empty")

			assert.Equal(t, len(txResponse.Txs), len(txResponse.TxResponses), "Number of transactions should match number of transaction responses")
		} else {
			fmt.Println("Transactions Response Test Passed - No transactions found at this height")
		}
	})

	t.Run("Test_WasmQuery", func(t *testing.T) {
		query := querier.WasmQuery{Client: cl, Options: &querier.WasmQueryOptions{ContractAddress: "bc1pnaze3k0mgk8gqtmw9es8hpuusrqaupa62dq3j744vnla874cvv0qakrmqh"}}
		balance, err := query.QueryCw20Balance("bc1pxu05753kc0jlc8dazxq9zhscdg48sq6hy5j5ayzxdftthwd633tsh8ph6z")
		require.NoError(t, err, "Failed to get balance")
		assert.NotNil(t, balance, "Balance should not be nil")
		fmt.Printf("Balance: %s\n", balance)
	})
}
