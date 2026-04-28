import { HTTPClient } from './http-client';

export class GraphQLAPI {
  constructor(private client: HTTPClient) {}

  async execute(query: string, variables?: Record<string, any>): Promise<any> {
    const data = { query };
    if (variables) {
      (data as any).variables = variables;
    }

    return this.client.request<any>({
      method: 'POST',
      endpoint: '/graphql',
      data,
    });
  }
}
