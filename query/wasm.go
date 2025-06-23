package query

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/RiemaLabs/probe/client"
	"github.com/gogo/protobuf/proto"
)

// SmartContractStateRequest represents the protobuf message for smart contract state queries
type SmartContractStateRequest struct {
	Address   string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	QueryData []byte `protobuf:"bytes,2,opt,name=query_data,json=queryData,proto3" json:"query_data,omitempty"`
}

func (m *SmartContractStateRequest) Reset()         { *m = SmartContractStateRequest{} }
func (m *SmartContractStateRequest) String() string { return proto.CompactTextString(m) }
func (*SmartContractStateRequest) ProtoMessage()    {}

// SmartContractStateResponse represents the protobuf message for smart contract state query responses
type SmartContractStateResponse struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *SmartContractStateResponse) Reset()         { *m = SmartContractStateResponse{} }
func (m *SmartContractStateResponse) String() string { return proto.CompactTextString(m) }
func (*SmartContractStateResponse) ProtoMessage()    {}

type WasmQueryOptions struct {
	ContractAddress string
}

type WasmQuery struct {
	Client  *client.ChainClient
	Options *WasmQueryOptions
}

func (q *WasmQuery) GetQueryContext() (context.Context, context.CancelFunc) {
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return ctx, cancel
}

func (q *WasmQuery) QueryContractState(msg []byte) ([]byte, error) {
	ctx, cancel := q.GetQueryContext()
	defer cancel()

	// Create the protobuf request
	queryRequest := &SmartContractStateRequest{
		Address:   q.Options.ContractAddress,
		QueryData: msg,
	}

	// Serialize using protobuf
	queryMsgBytes, err := proto.Marshal(queryRequest)
	if err != nil {
		return nil, err
	}

	// Use the correct ABCI query path for smart contract queries
	res, err := q.Client.RPCClient.ABCIQuery(ctx, "/cosmwasm.wasm.v1.Query/SmartContractState", queryMsgBytes)
	if err != nil {
		return nil, err
	}
	if res.Response.Code != 0 {
		return nil, fmt.Errorf("failed to query contract state: %s", res.Response.Log)
	}

	// Unmarshal the protobuf response
	var response SmartContractStateResponse
	if err := proto.Unmarshal(res.Response.Value, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contract state response: %w", err)
	}

	return response.Data, nil
}

func (q *WasmQuery) QueryCw20Balance(address string) (*big.Int, error) {
	queryMsg := []byte(fmt.Sprintf(`{"balance":{"address":"%s"}}`, address))
	res, err := q.QueryContractState(queryMsg)
	if err != nil {
		return nil, err
	}

	// Parse the response as JSON to extract the balance
	var response struct {
		Balance string `json:"balance"`
	}

	if err := json.Unmarshal(res, &response); err != nil {
		return nil, fmt.Errorf("failed to parse contract response: %w", err)
	}

	// Convert the balance string to big.Int
	balance := new(big.Int)
	balance.SetString(response.Balance, 10)

	return balance, nil
}
