import { test, expect } from "@playwright/test";
import { DashboardPage } from "../../page-object/dashboard.page";
import { BasePage } from "../../page-object/base.page";

test.describe("Dashboard Tests", () => {
    test.beforeEach(async ({ page }) => {
        await BasePage.goto(page, "/dashboard#summary");
    });

    test("Verify modal title is visible", async ({ page }) => {
        await DashboardPage.verifyAddDashboardTitleVisible(page);
    });

    test("Verify Validators button is clickable", async ({ page }) => {
        await DashboardPage.clickValidators(page);
    });

    test("Verify Accounts button is disabled", async ({ page }) => {
        await DashboardPage.verifyAccountsButtonDisabled(page);
    });

    test("Verify Ethereum selection works", async ({ page }) => {
        await DashboardPage.clickValidators(page);
        await DashboardPage.clickContinue(page);
        await DashboardPage.clickEthereum(page);
        await DashboardPage.clickContinueNetwork(page);

        await expect(DashboardPage.overview(page)).toBeVisible({ timeout: 15000 });
    });
});
