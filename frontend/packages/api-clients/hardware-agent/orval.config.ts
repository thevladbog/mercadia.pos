import { defineConfig } from 'orval';

export default defineConfig({
  'hardware-agent': {
    input: {
      target: '../../../../contracts/openapi/hardware-agent.openapi.json',
    },
    output: {
      mode: 'tags-split',
      target: 'src/generated',
      schemas: 'src/generated/models',
      client: 'react-query',
      httpClient: 'fetch',
      clean: true,
      override: {
        mutator: {
          path: './src/mutator.ts',
          name: 'customFetch',
        },
      },
    },
  },
});
