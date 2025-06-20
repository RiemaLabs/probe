package query

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/RiemaLabs/probe/client"
	"github.com/RiemaLabs/probe/utils"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"
)

// TxsAtHeightRPC Get Transactions for the given block height.
// Other query options can be specified with the GetTxsEventRequest.
//
// This version only uses the 26657 RPC endpoint (CometBFT).
func TxsAtHeightRPC(q *Query, height int64, codec client.Codec) (*txTypes.GetTxsEventResponse, error) {
	if q.Options.Pagination == nil {
		pagination := &query.PageRequest{Limit: 100}
		q.Options.Pagination = pagination
	}
	orderBy := txTypes.OrderBy_ORDER_BY_UNSPECIFIED

	req := &txTypes.GetTxsEventRequest{OrderBy: orderBy, Page: q.Options.Pagination.Offset, Limit: q.Options.Pagination.Limit, Query: "tx.height=" + fmt.Sprintf("%d", height)}
	return TxsRPC(q, height, req, codec)
}

// TxRPC Get Transactions for the given block height.
// Other query options can be specified with the GetTxsEventRequest.
//
// This version only uses the 26657 RPC endpoint (CometBFT).
func TxsRPC(q *Query, height int64, req *txTypes.GetTxsEventRequest, codec client.Codec) (*txTypes.GetTxsEventResponse, error) {
	ctx, cancel := q.GetQueryContext()
	defer cancel()

	header, err := q.Client.RPCClient.Header(ctx, &height)
	if err != nil {
		return nil, err
	}

	timestamp := header.Header.Time

	orderBy := ""
	if req.OrderBy == txTypes.OrderBy_ORDER_BY_ASC {
		orderBy = "asc"
	} else if req.OrderBy == txTypes.OrderBy_ORDER_BY_DESC {
		orderBy = "desc"
	}

	page := int(req.Page)
	perPage := int(req.Limit)

	txs, err := q.Client.RPCClient.TxSearch(ctx, req.Query, false, &page, &perPage, orderBy)
	if err != nil {
		return nil, err
	}

	nextKey := utils.MakeNextKey(req.Page, req.Limit, uint64(txs.TotalCount))

	return BuildGetTxsEventResponse(timestamp, txs, codec.TxConfig.TxDecoder(), nextKey)
}

func BuildGetTxsEventResponse(
	timestamp time.Time,
	results *coretypes.ResultTxSearch,
	decoder sdk.TxDecoder,
	nextKey []byte,
) (*txTypes.GetTxsEventResponse, error) {

	txs := make([]*txTypes.Tx, 0, len(results.Txs))
	txResponses := make([]*sdk.TxResponse, 0, len(results.Txs))
	for _, r := range results.Txs {
		tx, err := utils.TxBytesToProto(r.Tx, decoder)
		if err != nil {
			return nil, err
		}

		txs = append(txs, tx)
		anyTx, err := codectypes.NewAnyWithValue(tx)
		if err != nil {
			return nil, err
		}

		raw := r.TxResult.Log
		msgLogs, err := sdk.ParseABCILogs(raw)
		if err != nil {
			msgLogs = []sdk.ABCIMessageLog{{
				MsgIndex: 0,
				Log:      r.TxResult.Log,
				Events:   sdk.StringifyEvents(r.TxResult.Events),
			}}
		}

		dataHex := hex.EncodeToString(r.TxResult.Data)

		txResp := &sdk.TxResponse{
			Height:    r.Height,
			TxHash:    r.Hash.String(),
			Codespace: r.TxResult.Codespace,
			Code:      r.TxResult.Code,
			Data:      dataHex,
			RawLog:    r.TxResult.Log,
			Logs:      msgLogs,
			Info:      r.TxResult.Info,
			GasWanted: r.TxResult.GasWanted,
			GasUsed:   r.TxResult.GasUsed,
			Tx:        anyTx,
			Timestamp: timestamp.Format(time.RFC3339),
		}
		txResponses = append(txResponses, txResp)
	}

	return &txTypes.GetTxsEventResponse{
		Txs:         txs,
		TxResponses: txResponses,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   uint64(results.TotalCount),
		},
		Total: uint64(results.TotalCount),
	}, nil
}
