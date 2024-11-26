import type { Page } from "@playwright/test";

export class BasePage {
    readonly page: Page;

    constructor(page: Page){
        this.page = page;
    };

   async goto(endpoint: string): Promise <void> {
        await this.page.goto("https://v2-staging-mainnet.beaconcha.in" + endpoint);
        await this.page.waitForTimeout(2000);
    };
};
