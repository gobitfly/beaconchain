
import { LoginPage } from "../../page-object/login.page";
import { test, expect, goto } from "../../utils/helpers";
import { DashboardPage } from "../../page-object/dashboard.page";

test.describe("Login", () => {
    test.beforeEach(async ({page}) => {
        console.log('Base URL:', process.env.NUXT_PUBLIC_DOMAIN);
        await goto(page, "/login", { waitUntil: 'hydration' });
        await page.waitForLoadState('networkidle');
    });

    test.skip("Successful login with valid credentials", async ({ page }) => {
        await page.waitForLoadState('networkidle');
        await LoginPage.email(page).fill("");
        await LoginPage.password(page).fill("");
        await page.waitForLoadState('networkidle');
        await LoginPage.loginBtn(page).click();

        await expect(DashboardPage.dashboard(page)).toBeVisible({ timeout: 15000 });
    });
    test("The login button is active after filling in all fields", async ({ page }) => {
        await LoginPage.email(page).fill("tanya.testDummydata@bitfly.at");
        await LoginPage.password(page).fill("test1999");

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
        await LoginPage.email(page).fill("tanya.testtest@");
        await LoginPage.password(page).fill("test1999223");

        await expect(LoginPage.errorEmail(page)).toContainText("Please provide a valid email address.");
        await expect(LoginPage.loginBtn(page)).toBeDisabled();
    });

    test("Login with an incorrect password", async ({ page }) => {
        await LoginPage.email(page).fill("tanyaetst@bitfly.at");
        await LoginPage.password(page).fill("WrongPsw");
        await LoginPage.loginBtn(page).click();

        await expect(LoginPage.errorPassword(page)).toContainText("Please enter your password.");
        await expect(LoginPage.toastMessage(page)).toContainText("Cannot log in: your email or your password is unknown.");
    });
    test("Restrictions on the minimum password length", async ({ page }) => {
        await LoginPage.email(page).fill("tanya.testDummydata@bitfly.at");
        await LoginPage.password(page).fill("123");

        await expect(LoginPage.errorPassword(page)).toContainText("Please provide at least 5 characters.");
        await expect(LoginPage.loginBtn(page)).toBeDisabled();
    });

    test("Login with unregistered email address", async ({ page }) => {
        await LoginPage.email(page).fill("tanya.testDummydata@bitfly.at");
        await LoginPage.password(page).fill("Stasevych1999");
        await LoginPage.loginBtn(page).click();

        await expect(LoginPage.errorPassword(page)).toContainText("Please enter your password.");
        await expect(LoginPage.toastMessage(page)).toContainText("Cannot log in: your email or your password is unknown.");
    });
});
