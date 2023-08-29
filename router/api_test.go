package router

import (
	"eth-balance-proxy/client"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestBalanceEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		req          *http.Request
		address      string
		rpcNodeError bool
		rpcNodeOk    bool
		ok           bool
		code         int
	}{
		{
			name: "invalid address",
			req: &http.Request{
				URL: &url.URL{},
			},
			address: "XXX",
			code:    http.StatusBadRequest,
		},
		{
			name: "rpc node error",
			req: &http.Request{
				URL: &url.URL{},
			},
			address:      "0x00000000219ab540356cBB839Cbe05303d7705Fa",
			rpcNodeError: true,
			code:         http.StatusInternalServerError,
		},
		{
			name: "ok",
			req: &http.Request{
				URL: &url.URL{},
			},
			address:   "0x00000000219ab540356cBB839Cbe05303d7705Fa",
			rpcNodeOk: true,
			code:      http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRPCClient := client.NewMockRpcClientI(mockCtrl)
			mockRPCClient.EXPECT().EthSyncing().Return(false, nil)

			tt.req = mux.SetURLVars(tt.req, map[string]string{
				"address": tt.address,
			})

			if tt.rpcNodeError {
				mockRPCClient.EXPECT().GetAccountBalance(tt.address).Return("", errors.New("invalid request or something like that"))
			} else if tt.rpcNodeOk {
				mockRPCClient.EXPECT().GetAccountBalance(tt.address).Return("100000", nil)
			}

			router, _ := NewRouter(mockRPCClient)
			res := http.HandlerFunc(router.balance)
			rr := httptest.NewRecorder()
			res.ServeHTTP(rr, tt.req)
			assert.True(t, rr.Code == tt.code)
		})
	}
}

func TestReadinessEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		req          *http.Request
		rpcNodeError bool
		nodeSyncing  bool
		ok           bool
		code         int
	}{
		{
			name: "rpc node error",
			req: &http.Request{
				URL: &url.URL{},
			},
			rpcNodeError: true,
			code:         http.StatusInternalServerError,
		},
		{
			name: "node returns true",
			req: &http.Request{
				URL: &url.URL{},
			},
			nodeSyncing: true,
			code:        http.StatusServiceUnavailable,
		},
		{
			name: "node returns false",
			req: &http.Request{
				URL: &url.URL{},
			},
			nodeSyncing: false,
			code:        http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRPCClient := client.NewMockRpcClientI(mockCtrl)
			mockRPCClient.EXPECT().EthSyncing().Return(false, nil)

			if tt.rpcNodeError {
				mockRPCClient.EXPECT().EthSyncing().Return(false, errors.New("node does not answer"))
			} else if tt.nodeSyncing {
				mockRPCClient.EXPECT().EthSyncing().Return(true, nil)
			} else if !tt.nodeSyncing {
				mockRPCClient.EXPECT().EthSyncing().Return(false, nil)
			}

			router, _ := NewRouter(mockRPCClient)
			res := http.HandlerFunc(router.readiness)
			rr := httptest.NewRecorder()
			res.ServeHTTP(rr, tt.req)
			assert.True(t, rr.Code == tt.code)
		})
	}
}
