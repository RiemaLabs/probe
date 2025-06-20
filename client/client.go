// APACHE NOTICE
// Sourced with modifications from https://github.com/strangelove-ventures/lens
package client

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/RiemaLabs/probe/logger"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	libclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
)

type ChainClient struct {
	Config    *ChainClientConfig
	RPCClient rpcclient.Client
	Codec     Codec
}

func NewChainClient(ccc *ChainClientConfig) (*ChainClient, error) {

	codec, err := MakeCodec(ccc.Modules, ccc.CustomMsgTypeRegistry)

	if err != nil {
		return nil, err
	}

	cc := &ChainClient{
		Config: ccc,
		Codec:  codec,
	}

	if err := cc.Init(); err != nil {
		return nil, err
	}

	return cc, nil
}

func (cc *ChainClient) Init() error {

	timeout, _ := time.ParseDuration(cc.Config.Timeout)
	rpcClient, err := NewRPCClient(cc.Config.RPCAddr, timeout, cc.Config.Debug)
	if err != nil {
		return err
	}

	cc.RPCClient = rpcClient

	return nil
}

type loggingRoundTripper struct {
	rt http.RoundTripper
}

func (lrt *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	logger.Debug("Request URL", "url", req.URL.String())
	logger.Debug("Request Body", "body", string(bodyBytes))
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return lrt.rt.RoundTrip(req)
}

func NewRPCClient(addr string, timeout time.Duration, debug bool) (*rpchttp.HTTP, error) {
	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		return nil, err
	}
	httpClient.Timeout = timeout
	if debug {
		httpClient.Transport = &loggingRoundTripper{httpClient.Transport}
	}
	rpcClient, err := rpchttp.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}

	err = rpcClient.Start()
	if err != nil {
		return nil, err
	}
	return rpcClient, nil
}
