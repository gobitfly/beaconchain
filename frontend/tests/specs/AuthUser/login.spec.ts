import {test, expect }from "@playwright/test";
import { LoginPage } from "../../pageObject/login.page";
import { DashboardPage } from "~/tests/pageObject/dashboard.page";

let loginPage: LoginPage
test.describe('Login', ()=>{

    test.beforeEach(async({page})=>{
        loginPage = new LoginPage(page) 
        await loginPage.goto("/login")
    })

    test("Successful login with valid credentials", async ({page})=>{
        const dashboardPage = new DashboardPage(page)
        await page.waitForTimeout(5000)
        await loginPage.email.fill("tanya.stasevych@bitfly.at");
        await page.waitForTimeout(5000)
        await loginPage.password.fill("Stasevych1999");
    
        await loginPage.loginBtn.waitFor({ state: 'attached' })
        await page.waitForTimeout(5000)

        await loginPage.loginBtn.click();

        await expect(dashboardPage.overview).toBeVisible({timeout: 15000})
    })

    test("The login button is active after filling in all fields", async ()=>{
        await loginPage.email.fill("tanya.stasevych@bitfly.at");
        await loginPage.password.fill("Stasevych1999");

        await expect(loginPage.loginBtn).toBeEnabled();
    })

    test("Login with space", async ()=>{
        await loginPage.email.fill(" ");
        await loginPage.password.fill(" ");

        await expect(loginPage.errorEmail).toContainText("Please provide a valid email address.");
        await expect(loginPage.errorPassword).toContainText("Please provide at least 5 characters.");
        await expect(loginPage.loginBtn).toBeDisabled();
    })

    test("Login with an incorrect email format", async ()=>{
        await loginPage.email.fill("tanya.stasevych@");
        await loginPage.password.fill("Stasevych1999");

        await expect(loginPage.errorEmail).toContainText("Please provide a valid email address.");
        await expect(loginPage.loginBtn).toBeDisabled();
    })

    test("login with an incorrect password", async ()=>{
        await loginPage.email.fill("tanya.stasevych@bitfly.at");
        await loginPage.password.fill("WrongPsw");
        await loginPage.loginBtn.click();

        await expect(loginPage.errorPassword).toContainText("Please enter your password.");
        await expect(loginPage.toastMessage).toContainText("Cannot log in: your email or your password is unknown.");
    })

    test("Restrictions on the minimum password length", async ()=>{
        await loginPage.email.fill("tanya.stasevych@bitfly.at");
        await loginPage.password.fill("123");

        await expect(loginPage.errorPassword).toContainText("Please provide at least 5 characters.");
        await expect(loginPage.loginBtn).toBeDisabled();
    })

    test("Login with unregistered email address", async ()=>{
        await loginPage.email.fill("tanya.tes@bitfly.at");
        await loginPage.password.fill("Stasevych1999");
        await loginPage.loginBtn.click();

        await expect(loginPage.errorPassword).toContainText("Please enter your password.");
        await expect(loginPage.toastMessage).toContainText("Cannot log in: your email or your password is unknown.");
    })
})
