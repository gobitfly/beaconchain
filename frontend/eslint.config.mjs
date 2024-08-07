// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt({
  rules: {
    'vue/max-len': [
      'error',
      {
        code: 120,
        ignoreStrings: true,
        ignoreHTMLAttributeValues: true,
      },
    ],
  },
})
  .prepend({
    ignores: ['types/api', 'public', 'assets/css/prime_origin.scss'],
  })
  .override('nuxt/typescript/rules', {
    rules: {
      '@typescript-eslint/no-explicit-any': 'off', // TODO: remove this rule
    },
  })
