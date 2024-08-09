// @ts-check
import perfectionist from 'eslint-plugin-perfectionist'

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
    'vue/max-len': [
      'error',
      {
        code: 120,
        ignoreHTMLAttributeValues: true,
        ignoreStrings: true,
      },
    ],
  },
},
)
  .prepend({
    ignores: [
      'types/api',
      'public',
      'assets/css/prime_origin.scss',
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
