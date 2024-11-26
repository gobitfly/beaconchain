import { test, expect } from "@playwright/test";
import { DashboardPage } from "../../pageObject/dashboard.page";

test.describe("Add Dashboard Modal Tests", () => {
    let dashboardPage: DashboardPage;

    test.beforeEach(async ({ page }) => {
        dashboardPage = new DashboardPage(page);
        await page.goto("https://v2-staging-mainnet.beaconcha.in/dashboard#summary");
        await page.waitForTimeout(2000);
    });

    test("Verify modal title is visible", async () => {
        await dashboardPage.verifyAddDashboardTitleVisible();
    });

    test("Verify Validators button is clickable", async () => {
        await dashboardPage.clickValidators();
    });

    test("Verify Accounts button is disabled", async () => {
        await dashboardPage.verifyAccountsButtonDisabled();
    });

    test("Verify Ethereum selection works", async () => {
        await dashboardPage.clickValidators();
        await dashboardPage.clickContinue();
        await dashboardPage.clickEthereum();
        await dashboardPage.clickContinueNetwork();
        await expect(dashboardPage.overview).toBeVisible({timeout: 15000})
    });
});
