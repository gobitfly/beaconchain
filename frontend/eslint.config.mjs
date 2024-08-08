// @ts-check
import perfectionist from 'eslint-plugin-perfectionist'

import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt({
  rules: {
    'vue/max-len': [
      'error',
      {
        code: 120,
        ignoreHTMLAttributeValues: true,
        ignoreStrings: true,
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
