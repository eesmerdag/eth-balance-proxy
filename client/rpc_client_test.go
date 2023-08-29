package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetETHBalance(t *testing.T) {
	type fields struct {
		client HttpClient
	}
	tests := []struct {
		name    string
		fields  fields
		address string
		balance string
		err     bool
	}{
		{
			name: "error response",
			fields: fields{
				client: NewMockClientErrorCase(),
			},
			address: "fake-address",
			err:     true,
		},
		{
			name: "ok response",
			fields: fields{
				client: NewMockClientSuccessCaseForBalance(),
			},
			address: "fake-address",
			balance: "0.0001365000",
			err:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpcClient := &RpcClient{
				httpClient: tt.fields.client,
			}
			balance, err := rpcClient.GetAccountBalance(tt.address)
			if err != nil {
				assert.True(t, tt.err)
			} else {
				assert.Equal(t, balance, tt.balance)
			}
		})
	}
}

func TestEthSyncing(t *testing.T) {
	type fields struct {
		client HttpClient
	}
	tests := []struct {
		name   string
		fields fields
		err    bool
		resp   bool
	}{
		{
			name: "error from node response",
			fields: fields{
				client: NewMockClientErrorCase(),
			},
			err:  true,
			resp: false,
		},
		{
			name: "eth syncing",
			fields: fields{
				client: NewMockClientSuccessCaseForEthSyncing(),
			},
			err:  false,
			resp: true,
		},
		{
			name: "eth not syncing",
			fields: fields{
				client: NewMockClientSuccessCaseForEthNotSyncing(),
			},
			err:  false,
			resp: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpcClient := &RpcClient{
				httpClient: tt.fields.client,
			}
			resp, err := rpcClient.EthSyncing()
			if err != nil {
				assert.True(t, tt.err)
			} else {
				assert.False(t, tt.err)
				assert.Equal(t, resp, tt.resp)
			}
		})
	}
}
