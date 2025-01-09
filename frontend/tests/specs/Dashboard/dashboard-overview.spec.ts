import {
  expect, goto, test,
} from '../../utils/helpers'
import { addDashboardFlow } from '../../page-object/dashboard.page'

test.describe('Add Dashboard Tests Unauthorized User', () => {
  test.beforeEach(async ({ page }) => {
    await goto(page, '/dashboard#summary', 'networkidle')
  })

  test('Verify modal title is visible', async ({ page }) => {
    const title = addDashboardFlow.addDashboardTitle(page)
    await expect(title).toBeVisible()
  })

  test('Verify Validators button is clickable', async ({ page }) => {
    const validatorsButton = addDashboardFlow.validatorsButton(page)
    await validatorsButton.click()
  })

  test('Verify Accounts button is disabled', async ({ page }) => {
    const accountsButton = addDashboardFlow.accountsButton(page)
    await expect(accountsButton).toHaveText('Accounts Coming soonNo')
  })

  test('Verify Ethereum selection works', async ({ page }) => {
    await addDashboardFlow.validatorsButton(page).click()
    await addDashboardFlow.continueButton(page).click()
    await addDashboardFlow.ethereumOption(page).click()
    await addDashboardFlow.continueNetworkButton(page).click()

    const dashboard = addDashboardFlow.dashboard(page)
    await expect(dashboard).toBeVisible({ timeout: 15000 })
  })
})
