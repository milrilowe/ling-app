//  @ts-check

import { tanstackConfig } from '@tanstack/eslint-config'

export default [
  ...tanstackConfig,
  {
    ignores: ['eslint.config.js', 'prettier.config.js', '**/*.test.ts', '**/*.test.tsx', 'src/test/**'],
  },
  {
    rules: {
      // Disable import ordering rules
      'sort-imports': 'off',
      'import/order': 'off',
      'perfectionist/sort-imports': 'off',
      'perfectionist/sort-named-imports': 'off',
      // Allow inline type specifiers (common in shadcn/ui)
      'import/consistent-type-specifier-style': 'off',
      // Allow [] array syntax
      '@typescript-eslint/array-type': 'off',
      // These are too strict for practical use
      '@typescript-eslint/no-unnecessary-condition': 'off',
      '@typescript-eslint/no-unnecessary-type-assertion': 'off',
      '@typescript-eslint/consistent-type-imports': 'off',
      'no-shadow': 'off',
    },
  },
]
