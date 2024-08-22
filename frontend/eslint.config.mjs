import tailwind from 'eslint-plugin-tailwindcss'
// @ts-check
import perfectionist from 'eslint-plugin-perfectionist'
import eslintPluginJsonc from 'eslint-plugin-jsonc'

import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt({
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
    'vue/max-attributes-per-line': 'off',
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
    ...tailwind.configs['flat/recommended'].map(config => ({
      ...config,
      files: [ 'components/base/**/*.vue' ],
    })),
  )
  .append(
    ...eslintPluginJsonc.configs['flat/recommended-with-json'],
    {
      rules: {
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
