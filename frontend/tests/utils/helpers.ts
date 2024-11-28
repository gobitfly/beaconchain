
import type { Page } from 'playwright-core';

import { test as baseTest, expect as baseExpect } from "@nuxt/test-utils/playwright";

class TestHelper {
  static async goto(page: Page, url: string) {
    await page.goto(url);
  }
}

export const test = baseTest;
export const expect = baseExpect;
export const goto = TestHelper.goto;