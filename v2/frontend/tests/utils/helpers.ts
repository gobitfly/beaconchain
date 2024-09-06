import type { Page as playwrightPage } from 'playwright-core'
import {
  expect as playwrightExpext, test as playwrightTest,
} from '@nuxt/test-utils/playwright'

export const goto = async (page: Page, url: string, waitUntil?: | 'load' | 'networkidle') => {
  await page.goto(url, { waitUntil })
}
export type Page = playwrightPage
export const test = playwrightTest
export const expect = playwrightExpext
