import type { Page, Locator } from "@playwright/test";
import { BasePage } from "./base.page";

export class LoginPage extends BasePage {
    override readonly page: Page;
    readonly email: Locator
    readonly password: Locator
    readonly loginBtn: Locator
    readonly errorEmail: Locator
    readonly errorPassword: Locator
    readonly toastMessage: Locator

    constructor(page: Page) {
        super(page); 
        this.page = page; 
        this.email = this.page.locator("#email")
        this.password = this.page.locator("#password")
        this.loginBtn = this.page.locator('[aria-label="Log in"]')
        this.errorEmail = this.page.locator('.p-error').nth(0)
        this.errorPassword = this.page.locator('.p-error').nth(1)
        this.toastMessage = this.page.locator(".p-toast-message-error")
    }

    async login(email: string, password: string): Promise <void>{
        await this.email.fill(email);
        await this.password.fill(password);
        await this.loginBtn.click();
    }
}
