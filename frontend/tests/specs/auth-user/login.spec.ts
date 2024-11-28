// import { test, expect } from "@nuxt/test-utils/playwright";
// import { BasePage } from "../../page-object/base.page";
import { LoginPage } from "../../page-object/login.page";
import { DashboardPage } from "../../page-object/dashboard.page";

import { test, expect, goto } from "../../utils/helpers";
test.describe("Login", () => {
    test.beforeEach(async ({page}) => {
        console.log('Base URL:', process.env.NUXT_PUBLIC_DOMAIN);
        await goto(page, "/login");
    });
    test("Successful login with valid credentials", async ({ page }) => {
        await LoginPage.email(page).fill("tanya.stasevych@bitfly.at");
        await page.waitForTimeout(5000)
        await LoginPage.password(page).fill("Stasevych1999");
        await page.waitForTimeout(5000)
        await LoginPage.loginBtn(page).click();

        await expect(DashboardPage.dashboard(page)).toBeVisible({ timeout: 15000 });
    });

    test("The login button is active after filling in all fields", async ({ page }) => {
        await LoginPage.email(page).fill("tanya.stasevych@bitfly.at");
        await LoginPage.password(page).fill("Stasevych1999");

        await expect(LoginPage.loginBtn(page)).toBeEnabled();
    });

    test("Login with space", async ({ page }) => {
        await LoginPage.email(page).fill(" ");
        await LoginPage.password(page).fill(" ");

        await expect(LoginPage.errorEmail(page)).toContainText("Please provide a valid email address.");
        await expect(LoginPage.errorPassword(page)).toContainText("Please provide at least 5 characters.");
        await expect(LoginPage.loginBtn(page)).toBeDisabled();
    });

    test("Login with an incorrect email format", async ({ page }) => {
        await LoginPage.email(page).fill("tanya.stasevych@");
        await LoginPage.password(page).fill("Stasevych1999");

        await expect(LoginPage.errorEmail(page)).toContainText("Please provide a valid email address.");
        await expect(LoginPage.loginBtn(page)).toBeDisabled();
    });

    test("Login with an incorrect password", async ({ page }) => {
        await LoginPage.email(page).fill("tanya.stasevych@bitfly.at");
        await LoginPage.password(page).fill("WrongPsw");
        await LoginPage.loginBtn(page).click();

        await expect(LoginPage.errorPassword(page)).toContainText("Please enter your password.");
        await expect(LoginPage.toastMessage(page)).toContainText("Cannot log in: your email or your password is unknown.");
    });

    test("Restrictions on the minimum password length", async ({ page }) => {
        await LoginPage.email(page).fill("tanya.stasevych@bitfly.at");
        await LoginPage.password(page).fill("123");

        await expect(LoginPage.errorPassword(page)).toContainText("Please provide at least 5 characters.");
        await expect(LoginPage.loginBtn(page)).toBeDisabled();
    });

    test("Login with unregistered email address", async ({ page }) => {
        await LoginPage.email(page).fill("tanya.tes@bitfly.at");
        await LoginPage.password(page).fill("Stasevych1999");
        await LoginPage.loginBtn(page).click();

        await expect(LoginPage.errorPassword(page)).toContainText("Please enter your password.");
        await expect(LoginPage.toastMessage(page)).toContainText("Cannot log in: your email or your password is unknown.");
    });
});
