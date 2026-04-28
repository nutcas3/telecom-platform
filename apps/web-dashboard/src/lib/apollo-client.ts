import { ApolloClient, InMemoryCache, HttpLink, ApolloLink } from '@apollo/client';
import { SetContextLink } from '@apollo/client/link/context';

// GraphQL API endpoint
const httpLink = new HttpLink({
  uri: process.env.NEXT_PUBLIC_API_SERVER_URL || 'http://localhost:8080/v1/graphql',
});

// Auth link to add authentication token to requests
const authLink = new SetContextLink((prevContext, operation) => {
  // Get the authentication token from local storage if it exists
  const token = localStorage.getItem('auth_token');
  
  return {
    ...prevContext,
    headers: {
      ...prevContext.headers,
      authorization: token ? `Bearer ${token}` : '',
    },
  };
});

// Create Apollo Client instance
export const apolloClient = new ApolloClient({
  link: ApolloLink.from([authLink, httpLink]),
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
    },
    query: {
      fetchPolicy: 'network-only',
    },
    mutate: {
      fetchPolicy: 'network-only',
    },
  },
});

// GraphQL queries and mutations can be defined here
// Example:
/*
import { gql } from '@apollo/client';

export const GET_SERVICES = gql`
  query GetServices {
    services {
      id
      name
      status
      health
    }
  }
`;
*/
