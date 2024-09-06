import { defineConfig } from '@playwright/test'
import type { ConfigOptions } from '@nuxt/test-utils/playwright'

export default defineConfig<ConfigOptions>({
  timeout: 30000,
  use: {
    nuxt: {
      host: 'http://localhost:3000',
    },
  },
})
