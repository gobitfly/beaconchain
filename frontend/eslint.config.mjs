// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt(
  {
    ignores: [
      "types/api",
      "public",
    ],
  },
  {
      rules: {
        "@typescript-eslint/no-explicit-any": "off", // should be removed eventually
      },
  },
)
