import { defineConfig } from 'orval';

export default defineConfig({
  central: {
    input: {
      target: '../../../../contracts/openapi/central.openapi.json',
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
