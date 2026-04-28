package telecom

import (
	"context"
)

// GraphQLAPI handles GraphQL API calls
type GraphQLAPI struct {
	client *HTTPClient
}

// NewGraphQLAPI creates a new GraphQLAPI
func NewGraphQLAPI(client *HTTPClient) *GraphQLAPI {
	return &GraphQLAPI{client: client}
}

// Execute executes a GraphQL query
func (g *GraphQLAPI) Execute(ctx context.Context, query string, variables map[string]interface{}) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"query": query,
	}
	if variables != nil {
		data["variables"] = variables
	}

	var result map[string]interface{}
	err := g.client.Post(ctx, "/graphql", data, &result)
	return result, err
}
