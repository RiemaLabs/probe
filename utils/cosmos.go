package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

type hasProtoTx interface {
	GetProtoTx() *txtypes.Tx
}

func TxBytesToProto(rawTx []byte, txDecoder sdk.TxDecoder) (*txtypes.Tx, error) {
	sdkTx, err := txDecoder(rawTx)
	if err != nil {
		return nil, err
	}

	// Type assertion to get the underlying Tx
	if wrap, ok := sdkTx.(hasProtoTx); ok {
		return wrap.GetProtoTx(), nil
	}
	return nil, fmt.Errorf("failed to convert sdk.Tx to *txtypes.Tx")
}

func MakeNextKey(offset, limit, total uint64) []byte {
	if offset+limit >= total {
		return nil
	}

	next := offset + limit
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, next)
	return bytes.TrimLeft(b, "\x00")
}
