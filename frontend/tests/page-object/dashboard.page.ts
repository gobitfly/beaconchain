import type { Page } from '../utils/helpers'

export const addDashboardFlow = {
  accountsButton: (page: Page) => page.getByRole('button', { name: 'Accounts Coming soon' }),
  addDashboardTitle: (page: Page) => page.getByText('Add a new dashboard'),
  backButton: (page: Page) => page.locator('text=Back'),
  continueButton: (page: Page) => page.getByRole('button', { name: 'Continue' }),
  continueNetworkButton: (page: Page) => page.getByRole('button', { name: 'Continue' }),
  dashboard: (page: Page) => page.getByText('Online Validators'),
  ethereumOption: (page: Page) => page.getByRole('button', { name: 'Ethereum' }),
  gnosisOption: (page: Page) => page.getByRole('button', { name: 'Gnosis' }),
  manageValidatorsButton: (page: Page) => page.locator('text=Manage Validators'),
  onlineValidators: (page: Page) => page.locator('text=Online Validators'),
  validatorsButton: (page: Page) => page.getByRole('button', { name: 'Validators' }),
}
export const headerContainer = {
  addNewDashboardButton: (page: Page) => headerContainer.headerContainer(page)
    .getByRole('button', { name: 'Add new dashboard' }),
  headerContainer: (page: Page) => page.locator('.header-container'),
  notificationsButton: (page: Page) => headerContainer.headerContainer(page)
    .getByLabel('Notifications', { exact: true }),
  notificationsDashboard: (page: Page) => page.getByText('Email Notifications'),
  validatorsButton: (page: Page) => headerContainer.headerContainer(page)
    .getByLabel('Validators', { exact: true }),
}

export const firstHeaderRow = {
  firstHeaderRow: (page: Page) => page.locator('.header-row').nth(0),
  manageGroupsButton: (page: Page) => firstHeaderRow.firstHeaderRow(page)
    .getByLabel('Manage Groups'),
  managerGropsTitle: (page: Page) => page.getByText('Add or remove validator groups for your dashboard'),
  managerValidatorsTitle: (page: Page) => page.getByText('Add or remove validators from your dashboard'),
  manageValidatorsButton: (page: Page) => firstHeaderRow.firstHeaderRow(page)
    .getByLabel('Manage Validators'),
}

export const megaMenu = {
  dashboardButton: (page: Page) => megaMenu.megaMenu(page)
    .getByLabel('Dashboard'),
  megaMenu: (page: Page) => page.locator('.mega-menu'),
  notificationsButton: (page: Page) => megaMenu.megaMenu(page)
    .getByLabel('Notifications', { exact: true }),
  pricingButton: (page: Page) => megaMenu.megaMenu(page)
    .getByRole('link', { name: 'Pricing' }),
  pricingTitle: (page: Page) => page.getByText('Monitoring without limits on web and mobile.'),
}

export const overview = {
  aprBox: (page: Page) => page.locator('.overview .box').nth(3),
  aprBoxTooltip: (page: Page) => page.locator('.overview .box').nth(3).locator('.info'),
  efficiencyBox: (page: Page) => page.locator('.overview .box').nth(1),
  efficiencyBoxTooltip: (page: Page) => page.locator('.overview .box').nth(1).locator('.info'),
  modalWindow: (page: Page) => page.locator('[class*=overlay-mask]'),
  onlineValidatorBox: (page: Page) => page.locator('.overview .box').nth(0),
  onlineValidatorBoxButton: (page: Page) => page.locator('.overview .box').nth(0).getByRole('button', { name: 'Open Validator Overview' }),
  onlineValidatorBoxTooltip: (page: Page) => page.locator('.overview .box').nth(0).locator('.info'),
  onlineValidatorsTitle: (page: Page) => page.locator('.overview').locator('.main').getByText('Online Validators'),
  rewardsBox: (page: Page) => page.locator('.overview .box').nth(2),
  rewardsBoxTooltip: (page: Page) => page.locator('.overview .box').nth(2).locator('.info'),
  tooltipWindow: (page: Page) => page.locator('[class*= fit-content]'),
}

export const tabView = {
  blocks: (page: Page) => page.locator('.dashboard-tab-view').getByRole('tab', { name: 'Blocks' }),
  deposits: (page: Page) => page.locator('.dashboard-tab-view').getByRole('tab', { name: 'Deposits' }),
  headers: (page: Page) => page.locator('[class*= thead]'),
  heatmap: (page: Page) => page.locator('.dashboard-tab-view').getByRole('tab', { name: 'Heatmap' }),
  rewards: (page: Page) => page.locator('.dashboard-tab-view').getByRole('tab', { name: 'Rewards' }),
  summery: (page: Page) => page.locator('.dashboard-tab-view').getByRole('tab', { name: 'Summary' }),
  withdrawals: (page: Page) => page.locator('.dashboard-tab-view').getByRole('tab', { name: 'CL Withdrawals' }),
}
