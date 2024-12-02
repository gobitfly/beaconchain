import type { Page } from '../utils/helpers'

export const BasePage = {
  goto: async (page: Page, endpoint: string) => {
    await page.goto(endpoint)
    await page.waitForTimeout(2000)
  },
}
