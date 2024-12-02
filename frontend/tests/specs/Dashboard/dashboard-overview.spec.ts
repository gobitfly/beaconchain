import {
  expect, goto, test,
} from '../../utils/helpers'
import { DashboardPage } from '../../page-object/dashboard.page'

test.describe('Dashboard Tests Unauthorized User', () => {
  test.beforeEach(async ({ page }) => await goto(page, '/dashboard#summary', 'networkidle'))

  test('Verify modal title is visible', async ({ page }) => {
    const title = DashboardPage.addDashboardTitle(page)
    await expect(title).toBeVisible()
  })

  test('Verify Validators button is clickable', async ({ page }) => {
    const validatorsButton = DashboardPage.validatorsButton(page)
    await validatorsButton.click()
  })

  test('Verify Accounts button is disabled', async ({ page }) => {
    const accountsButton = DashboardPage.accountsButton(page)
    await expect(accountsButton).toHaveText('Accounts Coming soonNo')
  })

  test('Verify Ethereum selection works', async ({ page }) => {
    await DashboardPage.validatorsButton(page).click()
    await DashboardPage.continueButton(page).click()
    await DashboardPage.ethereumOption(page).click()
    await DashboardPage.continueNetworkButton(page).click()

    const dashboard = DashboardPage.dashboard(page)
    await expect(dashboard).toBeVisible({ timeout: 15000 })
  })
})
