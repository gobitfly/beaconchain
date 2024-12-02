import type { Page } from '../utils/helpers'

export const DashboardPage = {
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
