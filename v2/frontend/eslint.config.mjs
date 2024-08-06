// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt()
.prepend(
  {
    ignores: [
      "types/api",
      "public",
    ],
  }
)
.override('nuxt/typescript/rules', {
  rules: {
    '@typescript-eslint/no-explicit-any': 'off', // TODO: remove this rule
  }
})
