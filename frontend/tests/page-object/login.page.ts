import type { Page } from "@playwright/test";

export const LoginPage = {
    email: (page: Page) => page.locator("#email"),
    password: (page: Page) => page.locator("#password"),
    loginBtn: (page: Page) => page.locator('[aria-label="Log in"]'),
    errorEmail: (page: Page) => page.locator(".p-error").nth(0),
    errorPassword: (page: Page) => page.locator(".p-error").nth(1),
    toastMessage: (page: Page) => page.locator(".p-toast-message-error"),
};
