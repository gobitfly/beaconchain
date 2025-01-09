import {
  expect, goto, test,
} from '../../utils/helpers'
import {
  addDashboardFlow, firstHeaderRow, headerContainer, megaMenu, overview, tabView,
} from '../../page-object/dashboard.page'

test.describe('Add Dashboard Tests Unauthorized User', () => {
  test.beforeEach(async ({ page }) => {
    await goto(page, '/', 'networkidle')
    await addDashboardFlow.continueButton(page).click()
  })

  test('Validate Page Load', async ({ page }) => {
    await expect(addDashboardFlow.dashboard(page)).toBeVisible({ timeout: 15000 })
  })

  test('Validate Presence of Dashboard Buttons (Validators)', async ({ page }) => {
    await headerContainer.validatorsButton(page).click()
    await expect(addDashboardFlow.dashboard(page)).toBeVisible({ timeout: 15000 })
  })

  test('Validate Presence of Dashboard Buttons (Notifications)', async ({ page }) => {
    await headerContainer.notificationsButton(page).click()
    await expect(headerContainer.notificationsDashboard(page)).toBeVisible({ timeout: 15000 })
  })

  test('Validate Presence of Dashboard Buttons (Add new dashboard)', async ({ page }) => {
    await headerContainer.addNewDashboardButton(page).click()
    await expect(addDashboardFlow.accountsButton(page)).toBeVisible({ timeout: 15000 })
  })

  test('Validate Presence of Dashboard Buttons (Manage Groups)', async ({ page }) => {
    await firstHeaderRow.manageGroupsButton(page).click()
    await expect(firstHeaderRow.managerGropsTitle(page)).toBeVisible({ timeout: 15000 })
  })

  test('Validate Presence of Dashboard Buttons (Manage Validators)', async ({ page }) => {
    await firstHeaderRow.manageValidatorsButton(page).click()
    await expect(firstHeaderRow.managerValidatorsTitle(page)).toBeVisible({ timeout: 15000 })
  })

  test('Validate Presence of Dashboard Buttons (Dashboard)', async ({ page }) => {
    await megaMenu.dashboardButton(page).click()
    await expect(addDashboardFlow.dashboard(page)).toBeVisible({ timeout: 15000 })
  })

  test('Validate Presence of Dashboard Buttons (Pricing)', async ({ page }) => {
    await megaMenu.pricingButton(page).click()
    await expect(megaMenu.pricingTitle(page)).toBeVisible({ timeout: 15000 })
  })

  test('Validate Presence of Dashboard Buttons (Notifications in mega menu)', async ({ page }) => {
    await megaMenu.notificationsButton(page).click()
    await expect(headerContainer.notificationsDashboard(page)).toBeVisible({ timeout: 15000 })
  })

  test('Dashboard: Validate Online Validators Section ', async ({ page }) => {
    await expect(overview.onlineValidatorBox(page)).toContainText('0.0000 ETH')
    await expect(overview.onlineValidatorBox(page)).toContainText('Balance ')
    await expect(overview.onlineValidatorBox(page)).toContainText('Online Validators')
    await expect(overview.onlineValidatorBox(page)).toContainText('0 | 0')
  })

  test('Dashboard: Validate Online Validators Section (Button)', async ({ page }) => {
    await overview.onlineValidatorBoxButton(page).click()
    await expect(overview.modalWindow(page)).toBeVisible()
  })

  test('Dashboard: Validate Online Validators Section (tooltip)', async ({ page }) => {
    await overview.onlineValidatorBoxTooltip(page).click()
    await expect(overview.tooltipWindow(page)).toContainText('Your total share of all validator balances: 0.0000 ETHTotal amount effectively participating in staking: 0.0000 ETHTotal amount you deposited: 0.0000 ETH')
  })

  test('Dashboard: Validate Efficiency Section', async ({ page }) => {
    await expect(overview.efficiencyBox(page)).toContainText('24h Efficiency')
    await expect(overview.efficiencyBox(page)).toContainText('0.00%')
  })

  test('Dashboard: Validate Efficiency Section (tooltip)', async ({ page }) => {
    await overview.efficiencyBoxTooltip(page).click()
    await expect(overview.tooltipWindow(page)).toContainText('24h: 0.00%7d: 0.00%30d: 0.00%All time: 0.00%')
  })

  test('Dashboard: Validate Rewards Section', async ({ page }) => {
    await expect(overview.rewardsBox(page)).toContainText('30d Rewards')
    await expect(overview.rewardsBox(page)).toContainText('0 ETH')
  })

  test('Dashboard: Validate Rewards Section (tooltip)', async ({ page }) => {
    await overview.rewardsBoxTooltip(page).click()
    await expect(overview.tooltipWindow(page)).toContainText('24h: 0.0000 ETH (CL) 0.0000 ETH (EL)7d: 0.0000 ETH (CL) 0.0000 ETH (EL)30d: 0.0000 ETH (CL) 0.0000 ETH (EL)All time: 0.0000 ETH (CL) 0.0000 ETH (EL)')
  })

  test('Dashboard: Validate APR Section', async ({ page }) => {
    await expect(overview.aprBox(page)).toContainText('30d APR')
    await expect(overview.aprBox(page)).toContainText('0.00%')
  })

  test('Dashboard: Validate APR Section (tooltip)', async ({ page }) => {
    await overview.aprBoxTooltip(page).click()
    await expect(overview.tooltipWindow(page)).toContainText('24h: 0.00% (CL) 0.00% (EL)7d: 0.00% (CL) 0.00% (EL)30d: 0.00% (CL) 0.00% (EL)All time: 0.00% (CL) 0.00% (EL)')
  })

  test('Dashboard: Validate Summary', async ({ page }) => {
    const data = [
      'Group',
      'Status',
      'Validators',
      'Live',
      'Efficiency',
      'Attestations',
      'Proposals',
      'Rewards',
    ]

    await tabView.summery(page).click()

    for (const headers of data) {
      await expect(tabView.headers(page)).toContainText(headers)
    }
  })

  test('Dashboard: Validate Reward', async ({ page }) => {
    const data = [
      'Epoch',
      'Age',
      'Duty',
      'Group',
      'Reward',
      'El Rewards',
      'Cl Rewards',
    ]

    await tabView.rewards(page).click()

    for (const headers of data) {
      await expect(tabView.headers(page)).toContainText(headers)
    }
  })

  test('Dashboard: Validate Blocks', async ({ page }) => {
    const data = [
      'Proposer',
      'Group',
      'Epoch',
      'Slot',
      'Block',
      'Age',
      'Status',
      'Rewards Recipient',
      'Proposer Rewards',
    ]

    await tabView.blocks(page).click()

    for (const headers of data) {
      await expect(tabView.headers(page)).toContainText(headers)
    }
  })

  test('Dashboard: Validate Heatmap', async ({ page }) => {
    await expect(tabView.heatmap(page)).toBeDisabled()
  })

  test('Dashboard: Validate Deposits', async ({ page }) => {
    const data = [
      'Public Key',
      'Index',
      'Group',
      'Group',
      'Age',
      'From',
      'Depositor',
      'Tx Hash',
      'Withdrawal Credential',
      'Amount',
      'Valid',
    ]

    await tabView.deposits(page).click()

    for (const headers of data) {
      await expect(tabView.headers(page).nth(0)).toContainText(headers)
    }
  })

  test('Dashboard: Validate CL Withdrawals', async ({ page }) => {
    const data = [
      'Index',
      'Group',
      'Epoch',
      'Slot',
      'Age',
      'Recipient',
      'Amount',
    ]

    await tabView.withdrawals(page).click()

    for (const headers of data) {
      await expect(tabView.headers(page)).toContainText(headers)
    }
  })
})
