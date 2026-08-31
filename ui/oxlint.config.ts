import { defineConfig } from 'oxlint';

export default defineConfig({
  plugins: ['react', 'typescript'],
  env: {
    builtin: true
  },
  settings: {
    react: {
      version: '19.2.8'
    }
  },
  ignorePatterns: ['**/dist'],
  rules: {
    // eslint rules
    'no-case-declarations': 'error',
    'no-empty': 'error',
    'no-fallthrough': 'error',
    'no-prototype-builtins': 'error',
    'no-redeclare': 'error',
    'no-regex-spaces': 'error',
    'no-unexpected-multiline': 'error',
    'no-unused-expressions': 'error',
    'no-unused-vars': 'error',
    'no-array-constructor': 'error',
    // eslint-plugin-react-hooks rules
    'react/rules-of-hooks': 'error',
    'react/exhaustive-deps': 'warn',
    // typescript-eslint rules
    'typescript/ban-ts-comment': 'error',
    'typescript/no-empty-object-type': 'error',
    'typescript/no-explicit-any': 'error',
    'typescript/no-namespace': 'error',
    'typescript/no-require-imports': 'error',
    'typescript/no-unnecessary-type-constraint': 'error',
    'typescript/no-unsafe-function-type': 'error',
    // Disable this rule, fix it later
    'typescript/no-floating-promises': 'off'
  },
  options: {
    typeAware: true,
    reportUnusedDisableDirectives: 'error',
    maxWarnings: 0
  }
});
