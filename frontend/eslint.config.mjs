import eslintPluginNewlineDestructuring from 'eslint-plugin-newline-destructuring'
// @ts-check
import perfectionist from 'eslint-plugin-perfectionist'
import stylistic from '@stylistic/eslint-plugin'
import eslintPluginJsonc from 'eslint-plugin-jsonc'

import tailwindcss from 'eslint-plugin-tailwindcss'

import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt({
  plugins: { '@stylistic': stylistic },
  rules: {
    '@stylistic/array-bracket-newline': [
      'error',
      { minItems: 2 },
    ],
    '@stylistic/array-bracket-spacing': [
      'error',
      'always',
    ],
    '@stylistic/array-element-newline': [
      'error',
      { minItems: 2 },
    ],
    '@stylistic/member-delimiter-style': [
      'error',
      {
        multiline: {
          delimiter: 'comma',
        },
        singleline: {
          delimiter: 'comma',
        },
      },
    ],
    '@stylistic/object-curly-newline': [
      'error',
      {
        ExportDeclaration: {
          minProperties: 2,
          multiline: true,
        },
        ImportDeclaration: {
          minProperties: 2,
          multiline: true,
        },
        ObjectExpression: {
          consistent: true,
          minProperties: 2,
          multiline: true,
        },
        ObjectPattern: {
          consistent: true,
          minProperties: 2,
          multiline: true,
        },
      },
    ],
    '@stylistic/object-curly-spacing': [
      'error',
      'always',
    ],
    '@stylistic/object-property-newline': [
      'error',
      { allowAllPropertiesOnSameLine: true },
    ],
    '@stylistic/quotes': [
      'error',
      'single',
    ],
    'no-console': 'warn',
    'vue/max-len': [
      'error',
      {
        code: 120,
        ignoreHTMLAttributeValues: true,
        ignoreStrings: true,
      },
    ],
    'vue/v-bind-style': [
      'error',
      'shorthand',
      {
        sameNameShorthand: 'always',
      },
    ],
  },
},
)
  .prepend({
    ignores: [
      'assets/css/prime_origin.scss',
      'public',
      'package-lock.json',
      'tsconfig.json',
      'types/api',
    ],
  })
  .override('nuxt/typescript/rules', {
    // TODO: remove this rule
    rules: { '@typescript-eslint/no-explicit-any': 'off' },
  })
  .append(
    // @ts-ignore -- it seams like the plugin is currently not compatible but seems to work
    // (plz update, try to remove this or open an issue when stumbling across this line)
    perfectionist.configs['recommended-natural'],
    // @ts-check
    {
      rules: {
      // disable the rules as there are conflicts
        'perfectionist/sort-imports': 'off',
        'perfectionist/sort-vue-attributes': 'off',
      },
    },
  )
  .append(
    {
      plugins: { 'newline-destructuring': eslintPluginNewlineDestructuring },
      rules: {
        'newline-destructuring/newline': [
          'error',
          { items: 1 },
        ],
      },
    },
  )
  .append({
    files: [ 'layers/**/*.vue' ],
    plugins: { tailwindcss },
    rules: {
      'tailwindcss/classnames-order': [ 'error' ],
      'tailwindcss/enforces-negative-arbitrary-values': [ 'error' ],
      'tailwindcss/enforces-shorthand': [ 'error' ],
      'tailwindcss/no-contradicting-classname': [ 'error' ],
      'tailwindcss/no-unnecessary-arbitrary-value': [ 'error' ],
      //  the goal should be that we enable these one day
      // 'tailwindcss/no-arbitrary-value': [ 'error' ],
      // 'tailwindcss/no-custom-classname': [ 'error' ],
    },
    settings: {
      tailwindcss: {
        config: new URL('./layers/base/app/assets/css/main.css', import.meta.url).pathname,
      },
    },
  })
  .append(
    ...eslintPluginJsonc.configs['flat/recommended-with-json'],
    {
      files: [ '**/*.json' ],
      rules: {
        '@stylistic/no-multiple-empty-lines': [
          'error',
          {
            max: 0,
          },
        ],
        '@stylistic/no-trailing-spaces': 'error',
        'jsonc/indent': [
          'error',
          4,
        ],
        'jsonc/sort-array-values': [
          'warn',
          {
            order: {
              natural: true,
              type: 'asc',
            },
            pathPattern: 'conventionalCommits.scopes',
          },
        ],
        'jsonc/sort-keys': [
          'error',
          'asc',
          {
            natural: true,
          },
        ],
      },
    },
    {
      files: [ 'locales/**/*.json' ],
      rules: {
        'jsonc/key-name-casing': [
          'error',
          {
            camelCase: false,
            ignores: [
              'mGNO',
              'xDAI',
            ],
            SCREAMING_SNAKE_CASE: true,
            snake_case: true,
          },
        ],
      },
    },
  )
  .append({
    files: [
      'components/**/*.vue',
      'composables/**/*.ts',
    ],
    rules: {
      'no-restricted-syntax': [
        'error',
        {
          message: 'Use `useBcCookie()` instead.',
          selector:
          'CallExpression[callee.name="useCookie"]',
        },
      ],
    },
  },
  )
  .append({
    files: [ 'components/**/*.vue' ],
    ignores: [ '**/BcIcon.vue' ],
    rules: {
      'vue/no-restricted-syntax': [
        'error',
        {
          message: 'Use BcIcon instead.',
          // match BcIconSomethingThatComesAfter and LazyBcIconAnything but not BcIcon
          selector: 'VElement[name=/^(lazy)?bcicon(?!$)[a-z]+$/]',
        },
      ],
    },
  },
  )
