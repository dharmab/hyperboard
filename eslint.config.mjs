import js from '@eslint/js';
import globals from 'globals';

export default [
  {
    ignores: [
      '**/node_modules/**',
      '**/vendor/**',
      '**/vendored/**',
      '**/*.min.js',
    ],
  },
  js.configs.recommended,
  {
    files: ['internal/web/static/**/*.js'],
    languageOptions: {
      sourceType: 'script',
      globals: {
        ...globals.browser,
        htmx: 'readonly',
      },
    },
  },
  {
    files: ['internal/web/static/upload.js'],
    languageOptions: {
      globals: {
        showError: 'readonly',
      },
    },
  },
];
