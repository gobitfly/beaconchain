import { defineConfig, devices } from '@playwright/test';
import type { ConfigOptions } from '@nuxt/test-utils/playwright'
import path from 'path';
// import { fileURLToPath } from 'url';
// import { dirname, resolve } from 'path';

// const __filename = fileURLToPath(import.meta.url);
// const __dirname = dirname(__filename);
// const env = process.env.ENV || 'staging';

// dotenv.config({
//           path: resolve(__dirname, `.env.${env}`), 
//           override: true, 
//       });

import dotenv from 'dotenv'

dotenv.config({ path: path.resolve(process.cwd(), '.env') })
console.log("!!!!!!!!!!!!!!!!!", process.env)
  
export default defineConfig<ConfigOptions>({
  webServer: {
    command: 'npm run dev',
    url: process.env.NUXT_PUBLIC_DOMAIN,
    reuseExistingServer: !process.env.CI,
    stdout: 'ignore',
    stderr: 'pipe',
  },
  outputDir: './results',
  testMatch: '**/*.spec.ts',
  timeout: 30000,
  use: {
    baseURL: process.env.NUXT_PUBLIC_DOMAIN,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'], channel: 'chromium' },
    },
  ],
});
