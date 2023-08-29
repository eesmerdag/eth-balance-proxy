package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/savier89/ethunitconv"
	"io/ioutil"
	"net/http"
)

type RpcRequest struct {
	Version string      `json:"jsonrpc"`
	Id      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type RpcResponse struct {
	Version string          `json:"jsonrpc"`
	Id      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *RpcError       `json:"error,omitempty"`
}

type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RpcClientI interface {
	GetAccountBalance(address string) (string, error)
	EthSyncing() (bool, error)
}

type RpcClient struct {
	httpClient HttpClient
	endpoint   string
}

func NewRpcClient(httpClient *http.Client, endpoint string) *RpcClient {
	return &RpcClient{
		httpClient: httpClient,
		endpoint:   endpoint,
	}
}

func (c *RpcClient) GetAccountBalance(address string) (string, error) {
	resp, err := c.rpcCall("eth_getBalance", []string{
		address,
		"latest",
	})
	if err != nil {
		return "", err
	}

	var result string
	err = json.Unmarshal(resp.Result, &result)
	if err != nil {
		return "", fmt.Errorf("failed to get result: %w", err)
	}

	ether := ethunitconv.FromWei(result, "Ether")
	return ether, nil
}

func (c *RpcClient) EthSyncing() (bool, error) {
	resp, err := c.rpcCall("eth_syncing", []string{})
	if err != nil {
		return false, err
	}

	var result bool
	err = json.Unmarshal(resp.Result, &result)
	if err != nil {
		return false, fmt.Errorf("failed to get result: %w", err)
	}

	return result, nil
}

func (c *RpcClient) rpcCall(method string, params interface{}) (RpcResponse, error) {
	body, err := json.Marshal(RpcRequest{
		Version: "2.0",
		Method:  method,
		Params:  params,
		Id:      1,
	})

	if err != nil {
		return RpcResponse{}, errors.Wrap(err, "failed to marshal json")
	}

	req, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return RpcResponse{}, err
	}

	req.Header.Set("Accepts", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return RpcResponse{}, fmt.Errorf("failed to make request: %w", err)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return RpcResponse{}, fmt.Errorf("failed to get response body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return RpcResponse{}, fmt.Errorf("request failed, status code - %d, response - %s", resp.StatusCode, string(b))
	}

	var rpcResponse RpcResponse
	err = json.Unmarshal(b, &rpcResponse)
	if err != nil {
		return RpcResponse{}, fmt.Errorf("failed to parse response body: %w", err)
	}

	if rpcResponse.Error != nil {
		return rpcResponse, fmt.Errorf("rpc call failed, code - %d, message - %s", rpcResponse.Error.Code, rpcResponse.Error.Message)
	}

	return rpcResponse, nil
}
