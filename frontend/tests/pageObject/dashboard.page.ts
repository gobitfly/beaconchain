import type { Page, Locator } from "@playwright/test";
import { expect } from "@playwright/test"; 
import { BasePage } from "./base.page";

export class DashboardPage extends BasePage {
    override readonly page: Page;
    readonly overview: Locator;
    readonly onlineValidators: Locator;
    readonly heatmap: Locator;
    readonly manageValidatorsButton: Locator;
    readonly validatorsButton: Locator;
    readonly accountsButton: Locator;
    readonly continueButton: Locator;
    readonly addDashboardTitle: Locator;
    readonly ethereumOption: Locator;
    readonly gnosisOption: Locator;
    readonly backButton: Locator;
    readonly continueNetworkButton: Locator;

    constructor(page: Page) {
        super(page);
        this.page = page;
        this.overview = this.page.locator(".overview");
        this.onlineValidators = this.page.locator("text=Online Validators");
        this.heatmap = this.page.locator(".heatmap");
        this.manageValidatorsButton = this.page.locator("text=Manage Validators");
        this.validatorsButton = this.page.getByRole('button', { name: 'Validators' });
        this.accountsButton = this.page.getByRole('button', { name: 'Accounts Coming soon' });
        this.continueButton = this.page.getByRole('button', { name: 'Continue' });
        this.addDashboardTitle = this.page.getByText('Add a new dashboard');
        this.ethereumOption = this.page.getByRole('button', { name: "Ethereum"});
        this.gnosisOption = this.page.getByRole('button', { name: "Gnosis"});
        this.backButton = this.page.locator('text=Back');
        this.continueNetworkButton = this.page.getByRole('button', { name: 'Continue' });
    }

    async clickContinue() {
        await this.continueButton.click();
    }

    async verifyAddDashboardTitleVisible() {
        await expect(this.addDashboardTitle).toBeVisible();
    }

    async verifyAccountsButtonDisabled() {
        await expect(this.accountsButton).toHaveText("Accounts Coming soonNo");
    }

    async clickManageValidators() {
        await this.manageValidatorsButton.click();
    }

    async verifyOnlineValidatorsCount(expectedCount: string) {
        const count = await this.onlineValidators.locator(".count").textContent();
        expect(count).toBe(expectedCount);
    }

    async verifyHeatmapBlockColor(row: number, column: number, expectedColor: string) {
        const block = this.page.locator(`.heatmap-row:nth-child(${row}) .heatmap-block:nth-child(${column})`);
        const color = await block.evaluate((el) => getComputedStyle(el).backgroundColor);
        expect(color).toBe(expectedColor);
    }
    async clickValidators() {
        await this.validatorsButton.click();
    }

    async clickAccounts() {
        await this.accountsButton.click();
    }

    async clickEthereum() {
        await this.ethereumOption.click();
    }

    async clickBack() {
        await this.backButton.click();
    }

    async clickContinueNetwork() {
        await this.continueNetworkButton.click();
    }

    async verifyGnosisOptionDisabled() {
        await expect(this.gnosisOption).toHaveAttribute('class', /disabled|coming-soon/);
    }
}
