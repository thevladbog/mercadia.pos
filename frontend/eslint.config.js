import js from '@eslint/js';
import eslintConfigPrettier from 'eslint-config-prettier';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import { defineConfig, globalIgnores } from 'eslint/config';
import globals from 'globals';
import tseslint from 'typescript-eslint';

export default defineConfig([
  globalIgnores(['**/node_modules/**', '**/dist/**', '**/src/generated/**']),
  {
    files: ['**/*.{js,mjs,cjs,ts,tsx}'],
    extends: [js.configs.recommended, ...tseslint.configs.recommended, eslintConfigPrettier],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': [
        'warn',
        {
          allowConstantExport: true,
          allowExportNames: [
            'ACCENT_PRESETS',
            'ACCENT_PRESET_OPTIONS',
            'Dialog',
            'DialogClose',
            'DialogTrigger',
            'Tabs',
            'accentPresetLabel',
            'applyTheme',
            'badgeVariants',
            'buttonVariants',
            'cardVariants',
            'clearTheme',
            'computeDenominationTotal',
            'deriveAccentTokens',
            'kindLabel',
            'resolveAccentHex',
            'statusLabel',
            'useAuth',
            'useTheme',
          ],
        },
      ],
    },
  },
  {
    files: ['**/*.{js,mjs,cjs}'],
    languageOptions: {
      globals: globals.node,
    },
  },
  {
    files: ['eslint.config.js'],
    languageOptions: {
      globals: globals.node,
    },
  },
]);
