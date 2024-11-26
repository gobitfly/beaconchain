import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";

export const DashboardPage = {
    overview: (page: Page) => page.locator(".overview"),
    onlineValidators: (page: Page) => page.locator("text=Online Validators"),
    heatmap: (page: Page) => page.locator(".heatmap"),
    manageValidatorsButton: (page: Page) => page.locator("text=Manage Validators"),
    validatorsButton: (page: Page) => page.getByRole("button", { name: "Validators" }),
    accountsButton: (page: Page) => page.getByRole("button", { name: "Accounts Coming soon" }),
    continueButton: (page: Page) => page.getByRole("button", { name: "Continue" }),
    addDashboardTitle: (page: Page) => page.getByText("Add a new dashboard"),
    ethereumOption: (page: Page) => page.getByRole("button", { name: "Ethereum" }),
    gnosisOption: (page: Page) => page.getByRole("button", { name: "Gnosis" }),
    backButton: (page: Page) => page.locator("text=Back"),
    continueNetworkButton: (page: Page) => page.getByRole("button", { name: "Continue" }),

    clickContinue: async (page: Page) => {
        await DashboardPage.continueButton(page).click();
    },

    verifyAddDashboardTitleVisible: async (page: Page) => {
        await expect(DashboardPage.addDashboardTitle(page)).toBeVisible();
    },

    verifyAccountsButtonDisabled: async (page: Page) => {
        await expect(DashboardPage.accountsButton(page)).toHaveText("Accounts Coming soonNo");
    },

    clickManageValidators: async (page: Page) => {
        await DashboardPage.manageValidatorsButton(page).click();
    },

    verifyOnlineValidatorsCount: async (page: Page, expectedCount: string) => {
        const count = await DashboardPage.onlineValidators(page).locator(".count").textContent();
        expect(count).toBe(expectedCount);
    },

    verifyHeatmapBlockColor: async (page: Page, row: number, column: number, expectedColor: string) => {
        const block = page.locator(`.heatmap-row:nth-child(${row}) .heatmap-block:nth-child(${column})`);
        const color = await block.evaluate((el) => getComputedStyle(el).backgroundColor);
        expect(color).toBe(expectedColor);
    },

    clickValidators: async (page: Page) => {
        await DashboardPage.validatorsButton(page).click();
    },

    clickAccounts: async (page: Page) => {
        await DashboardPage.accountsButton(page).click();
    },

    clickEthereum: async (page: Page) => {
        await DashboardPage.ethereumOption(page).click();
    },

    clickBack: async (page: Page) => {
        await DashboardPage.backButton(page).click();
    },

    clickContinueNetwork: async (page: Page) => {
        await DashboardPage.continueNetworkButton(page).click();
    },

    verifyGnosisOptionDisabled: async (page: Page) => {
        await expect(DashboardPage.gnosisOption(page)).toHaveAttribute("class", /disabled|coming-soon/);
    },
};
