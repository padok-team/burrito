import hooksPlugin from 'eslint-plugin-react-hooks';
import tseslint from 'typescript-eslint';
import eslint from '@eslint/js';

export default [
  {
    plugins: {
      'react-hooks': hooksPlugin,
    },
    rules: {
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
    },
  },
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {
    ignores: ['**/dist'],
  },
];
