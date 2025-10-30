import path from 'path'
import dotenv from 'dotenv'
import {
  defineConfig, devices,
} from '@playwright/test'

const envName = process.env.ENV || 'local'
const envFile = `.env.${envName}`
dotenv.config({ path: path.resolve(process.cwd(), 'tests', envFile) })

console.log(`✅ Running Playwright with environment: ${envName}`)
console.log(`📄 Loaded env file: ${envFile}`)
console.log(`🌍 Base URL: ${process.env.URL}`)

export default defineConfig({
  outputDir: './results',
  projects: [ {
    name: 'chromium',
    use: {
      ...devices['Desktop Chrome'],
      channel: 'chromium',
    },
  } ],
  retries: 3,
  testMatch: [ '**/*.spec.ts' ],
  timeout: 20000,
  use: {
    baseURL: process.env.URL,
    ignoreHTTPSErrors: true,
  },
})
