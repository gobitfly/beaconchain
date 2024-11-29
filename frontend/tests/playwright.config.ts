
import path from 'path';
import dotenv from 'dotenv'
import { fileURLToPath } from 'node:url'
import { defineConfig, devices } from '@playwright/test';
import type { ConfigOptions } from '@nuxt/test-utils/playwright'

dotenv.config({ path: path.resolve(process.cwd(), '.env') })
  
export default defineConfig<ConfigOptions>({
  outputDir: './results',
  testMatch: '**/*.spec.ts',
  timeout: 300000,
  use: {
    nuxt: {
      rootDir: fileURLToPath(new URL('.', import.meta.url))
    },
    baseURL: process.env.NUXT_PUBLIC_DOMAIN,
    ignoreHTTPSErrors: true,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'], channel: 'chromium' },
    },
  ],
});
