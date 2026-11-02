package graphql

import (
	"context"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/playingwithclouds/veil/internal/api/graphql/generated"
	"github.com/playingwithclouds/veil/internal/tokens"
)

// NewHandler returns the HTTP handler for the GraphQL endpoint.
//
// Transports are added by hand so the websocket transport uses a permissive
// upgrader: the default one only allows same-origin upgrades, which rejects
// the web app's cross-origin (:3000 → :8080) connection.
func NewHandler(resolver *Resolver) http.Handler {
	server := gqlhandler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	server.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			// Local app without accounts: any origin may connect.
			CheckOrigin:     func(r *http.Request) bool { return true },
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	})
	server.AddTransport(transport.Options{})
	server.AddTransport(transport.GET{})
	server.AddTransport(transport.POST{})
	server.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	server.Use(extension.Introspection{})
	server.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})
	server.AroundOperations(checkOperationScope)
	return server
}

// checkOperationScope refuses an operation the request's token doesn't
// allow: queries and subscriptions need read, mutations write. Requests that
// passed no token check (loopback, tokens not required) are trusted.
func checkOperationScope(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	operation := graphql.GetOperationContext(ctx).Operation
	scope := tokens.ScopeRead
	if operation != nil && operation.Operation == ast.Mutation {
		scope = tokens.ScopeWrite
	}
	if tokens.FromContext(ctx).Allows(scope) {
		return next(ctx)
	}
	return graphql.OneShot(graphql.ErrorResponse(ctx, "this token lacks the %s scope", scope))
}
