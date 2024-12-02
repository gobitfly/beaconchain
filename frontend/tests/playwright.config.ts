import path from 'path'
import { fileURLToPath } from 'node:url'
import dotenv from 'dotenv'
import {
  defineConfig, devices,
} from '@playwright/test'
import type { ConfigOptions } from '@nuxt/test-utils/playwright'

dotenv.config({ path: path.resolve(process.cwd(), '.env') })

export default defineConfig<ConfigOptions>({
  outputDir: './results',
  projects: [ {
    name: 'chromium',
    use: {
      ...devices['Desktop Chrome'], channel: 'chromium',
    },
  } ],
  testMatch: '**/*.spec.ts',
  timeout: 300000,
  use: {
    baseURL: process.env.NUXT_PUBLIC_DOMAIN,
    ignoreHTTPSErrors: true,
    nuxt: {
      rootDir: fileURLToPath(new URL('.', import.meta.url)),
    },
  },
})
