import type { Page as PlaywrightPage } from '@playwright/test'
import {
  expect as playwrightExpect, test as playwrightTest,
} from '@playwright/test'

export const goto = async (
  page: Page,
  url: string,
  waitUntil?: 'load' | 'networkidle',
) => {
  await page.goto(url, { waitUntil })
}

export type Page = PlaywrightPage
export const test = playwrightTest
export const expect = playwrightExpect
