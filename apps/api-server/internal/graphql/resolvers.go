// Package graphql hosts the GraphQL resolver, split across multiple files:
//
//   resolvers_subscriber.go     - Subscriber queries
//   resolvers_usage.go          - Usage events & stats
//   resolvers_billing.go        - Invoices, rating plans, top-up
//   resolvers_mutations.go      - Subscriber CRUD mutations
//   resolvers_esim.go           - eSIM + payment-method mutations
//   resolvers_alerts.go         - Alerts, system stats, maintenance
//   resolvers_subscriptions.go  - Streaming subscriptions + charging
//
// This file holds only the root Resolver struct, constructor, helpers, and
// the HTTP handler wiring.
package graphql

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// ensure used imports remain referenced when helpers below are the only consumers
var _ = context.Background
var _ = base64.StdEncoding

// Resolver is the root GraphQL resolver.
type Resolver struct {
	db         *database.Database
	subscriber *services.SubscriberService
	charging   *services.ChargingService
	es2Service *services.ES2Service
}

// NewResolver creates a new GraphQL resolver with wired-up services.
func NewResolver(db *database.Database, cfg *config.Config) *Resolver {
	return &Resolver{
		db:         db,
		subscriber: services.NewSubscriberService(db, cfg),
		charging:   services.NewChargingService(db, cfg),
		es2Service: services.NewES2Service(&cfg.ES2),
	}
}

// parseID decodes a GraphQL "Subscriber:<n>" opaque ID to a uint.
func parseID(id string) (uint, error) {
	var n int
	_, err := fmt.Sscanf(id, "Subscriber:%d", &n)
	if err != nil {
		return 0, fmt.Errorf("invalid subscriber ID format")
	}
	return uint(n), nil
}

// parseCursorOffset reads the numeric offset from a decoded cursor string.
func parseCursorOffset(cursor string) int {
	var offset int
	fmt.Sscanf(cursor, "%d", &offset)
	return offset
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string { return new(s) }

// buildConnectionPageInfo builds a Relay-style PageInfo for a page of size count.
func buildConnectionPageInfo(offset, count int, total int64) *PageInfo {
	var startCursor, endCursor *string
	if count > 0 {
		startCursor = strPtr(base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%d", offset+1)))
		endCursor = strPtr(base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%d", offset+count)))
	}
	return &PageInfo{
		HasNextPage:     int64(offset+count) < total,
		HasPreviousPage: offset > 0,
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}
}

// buildUsageEventFilter builds a SQL WHERE clause from optional filters.
func buildUsageEventFilter(imsi *string, filter *UsageEventFilter) string {
	conditions := []string{}
	if imsi != nil {
		conditions = append(conditions, fmt.Sprintf("imsi = '%s'", *imsi))
	}
	if filter != nil {
		if filter.UsageType != nil {
			conditions = append(conditions, fmt.Sprintf("usage_type = '%s'", *filter.UsageType))
		}
		if filter.TimestampGte != nil {
			conditions = append(conditions, fmt.Sprintf("timestamp >= '%s'", filter.TimestampGte.Format(time.RFC3339)))
		}
		if filter.TimestampLte != nil {
			conditions = append(conditions, fmt.Sprintf("timestamp <= '%s'", filter.TimestampLte.Format(time.RFC3339)))
		}
		if filter.CostMin != nil {
			conditions = append(conditions, fmt.Sprintf("cost >= %f", *filter.CostMin))
		}
		if filter.CostMax != nil {
			conditions = append(conditions, fmt.Sprintf("cost <= %f", *filter.CostMax))
		}
	}
	if len(conditions) > 0 {
		return "WHERE " + strings.Join(conditions, " AND ")
	}
	return ""
}

// Query returns the QueryResolver (itself).
func (r *Resolver) Query() QueryResolver { return r }

// Mutation returns the MutationResolver (itself).
func (r *Resolver) Mutation() MutationResolver { return r }

// Subscription returns the SubscriptionResolver (itself).
func (r *Resolver) Subscription() SubscriptionResolver { return r }

// SetupGraphQLHandler wires the GraphQL HTTP endpoint and playground.
func SetupGraphQLHandler(router *gin.Engine, resolver *Resolver) {
	graphqlHandler := handler.NewDefaultServer(
		NewExecutableSchema(Config{Resolvers: resolver}),
	)
	router.POST("/graphql", gin.WrapH(graphqlHandler))
	if gin.Mode() == gin.DebugMode {
		router.GET("/graphql", gin.WrapH(playground.Handler("GraphQL Playground", "/graphql")))
	}
}
