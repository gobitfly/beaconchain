import type { Page } from "@playwright/test";

export const DashboardPage = {
    dashboard: (page: Page) => page.getByText('Online Validators'),
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
};
