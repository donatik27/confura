package rpc

import (
	"context"
	"net/http"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/conflux-chain/conflux-infura/node"
	"github.com/conflux-chain/conflux-infura/util/rate"
	"github.com/conflux-chain/conflux-infura/util/rpc/handlers"
	"github.com/conflux-chain/conflux-infura/util/rpc/middlewares"
	"github.com/openweb3/go-rpc-provider"
)

const (
	ctxKeyClientProvider = handlers.CtxKey("Infura-RPC-Client-Provider")
	ctxKeyClient         = handlers.CtxKey("Infura-RPC-Client")
)

// go-rpc-provider only supports static middlewares for RPC server.
func init() {
	// middlewares executed in order

	// rate limit
	rpc.HookHandleBatch(middlewares.RateLimitBatch)
	rpc.HookHandleCallMsg(middlewares.RateLimit)

	// metrics
	rpc.HookHandleBatch(middlewares.MetricsBatch)
	rpc.HookHandleCallMsg(middlewares.Metrics)

	// log
	rpc.HookHandleBatch(middlewares.LogBatch)
	rpc.HookHandleCallMsg(middlewares.Log)

	// cfx/eth client
	rpc.HookHandleCallMsg(clientMiddleware)
}

// Inject values into context for static RPC call middlewares, e.g. rate limit
func httpMiddleware(registry *rate.Registry, clientProvider interface{}) handlers.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ctx = context.WithValue(ctx, handlers.CtxKeyRealIP, handlers.GetIPAddress(r))
			ctx = context.WithValue(ctx, handlers.CtxKeyRateRegistry, registry)
			ctx = context.WithValue(ctx, ctxKeyClientProvider, clientProvider)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func clientMiddleware(next rpc.HandleCallMsgFunc) rpc.HandleCallMsgFunc {
	return func(ctx context.Context, msg *rpc.JsonRpcMessage) *rpc.JsonRpcMessage {
		var client interface{}
		var err error

		if cfxProvider, ok := ctx.Value(ctxKeyClientProvider).(*node.CfxClientProvider); ok {
			switch msg.Method {
			case "cfx_getLogs":
				client, err = cfxProvider.GetClientByIPGroup(ctx, node.GroupCfxLogs)
			default:
				client, err = cfxProvider.GetClientByIP(ctx)
			}
		} else if ethProvider, ok := ctx.Value(ctxKeyClientProvider).(*node.EthClientProvider); ok {
			switch msg.Method {
			case "eth_getLogs":
				client, err = ethProvider.GetClientByIPGroup(ctx, node.GroupEthLogs)
			default:
				client, err = ethProvider.GetClientByIP(ctx)
			}
		} else {
			return next(ctx, msg)
		}

		// no fullnode available to request RPC
		if err != nil {
			return msg.ErrorResponse(err)
		}

		ctx = context.WithValue(ctx, ctxKeyClient, client)

		return next(ctx, msg)
	}
}

func GetCfxClientFromContext(ctx context.Context) sdk.ClientOperator {
	return ctx.Value(ctxKeyClient).(sdk.ClientOperator)
}

func GetEthClientFromContext(ctx context.Context) *node.Web3goClient {
	return ctx.Value(ctxKeyClient).(*node.Web3goClient)
}