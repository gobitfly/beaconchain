import type { Page } from "@playwright/test";

export const BasePage = {
    goto: async (page: Page, endpoint: string) => {
        await page.goto("https://v2-staging-mainnet.beaconcha.in" + endpoint);
        await page.waitForTimeout(2000);
    },
};
