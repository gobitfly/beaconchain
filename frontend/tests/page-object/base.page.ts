import type { Page } from "@playwright/test";

export const BasePage = {
    goto: async (page: Page, endpoint: string) => {
        await page.goto(endpoint);
        await page.waitForTimeout(2000);
    },
};
