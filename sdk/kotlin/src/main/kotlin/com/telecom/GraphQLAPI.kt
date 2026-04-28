package com.telecom

/**
 * API for GraphQL queries
 */
class GraphQLAPI(private val client: HTTPClient) {
    
    /**
     * Execute a GraphQL query
     */
    suspend fun execute(query: String, variables: Map<String, Any>? = null): Map<String, Any> {
        val request = GraphQLRequest(query, variables)
        return client.post("/graphql", request)
    }
}
