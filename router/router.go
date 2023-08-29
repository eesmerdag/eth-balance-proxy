package router

import (
	"encoding/json"
	"errors"
	"eth-balance-proxy/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path"})
)

type Balance struct {
	Balance string `json:"balance"`
	Unit    string `json:"unit"`
}

type ErrorResp struct {
	Message string
	Code    int
}

type Router struct {
	router    *mux.Router
	rpcClient client.RpcClientI
}

func NewRouter(rpcClientI client.RpcClientI) (*Router, error) {
	router := mux.NewRouter()

	r := &Router{
		router:    router,
		rpcClient: rpcClientI,
	}

	syncing, err := r.rpcClient.EthSyncing()
	if err != nil {
		return nil, err
	} else if syncing {
		return nil, errors.New("RPC node syncing eth at the moment")
	}

	router.HandleFunc("/eth/balance/{address}", r.balance).Methods(http.MethodGet)
	router.HandleFunc("/live", r.liveness).Methods(http.MethodGet)
	router.HandleFunc("/ready", r.readiness).Methods(http.MethodGet)
	router.Handle("/metrics", promhttp.Handler())
	router.Use(panicRecovery)

	router.Use(prometheusMiddleware)
	return r, nil
}

func panicRecovery(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				errorResponse(w, "unexpected internal error", http.StatusInternalServerError)
				return
			}
		}()

		h.ServeHTTP(w, r)
	})
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}

func (rt Router) balance(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["address"]

	if !common.IsHexAddress(addr) {
		errorResponse(w, "address is not valid", http.StatusBadRequest)
		return
	}

	res, err := rt.rpcClient.GetAccountBalance(addr)
	if err != nil {
		errorResponse(w, "error getting balance from node", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Balance{Balance: res, Unit: "ether"})
}

func (rt Router) liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (rt Router) readiness(w http.ResponseWriter, r *http.Request) {
	syncing, err := rt.rpcClient.EthSyncing()
	if err != nil {
		errorResponse(w, "error getting eth_syncing info", http.StatusInternalServerError)
		return
	}

	if syncing {
		errorResponse(w, "RPC node syncing eth at the moment", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func errorResponse(w http.ResponseWriter, message string, code int) {
	errObj := ErrorResp{
		Message: message,
		Code:    code,
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errObj)
}
