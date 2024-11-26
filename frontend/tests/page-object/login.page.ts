import type { Page } from "@playwright/test";

export const LoginPage = {
    email: (page: Page) => page.locator("#email"),
    password: (page: Page) => page.locator("#password"),
    loginBtn: (page: Page) => page.locator('[aria-label="Log in"]'),
    errorEmail: (page: Page) => page.locator(".p-error").nth(0),
    errorPassword: (page: Page) => page.locator(".p-error").nth(1),
    toastMessage: (page: Page) => page.locator(".p-toast-message-error"),

    login: async (page: Page, userEmail: string, userPassword: string): Promise<void> => {
        await LoginPage.email(page).fill(userEmail);
        await LoginPage.password(page).fill(userPassword);
        await LoginPage.loginBtn(page).click();
    },
};
