import {
  expect, goto, test,
} from '../../../utils/helpers'

test.describe('Dashboard Tests Unauthorized User', {
  tag: '@landing-page',
}, () => {
  test.beforeEach(async ({ page }) => await goto(page, '/products#stacking-hub', 'networkidle'))

  test('verify Landing Page is opened', async ({ page }) => {
    await expect(page.getByText('Seamless access to blockchain data on')).toBeVisible()
  })

  test('verify three fragment Links exist (API, Stakinghub, Explorer)', async ({ page }) => {
    await expect(page.getByRole('navigation').getByRole('link', { name: 'API' })).toBeVisible()
    await expect(page.getByRole('navigation').getByRole('link', { name: 'StakingHub' })).toBeVisible()
    await expect(page.getByRole('navigation').getByRole('link', { name: 'Explorer' })).toBeVisible()
    await page.getByRole('link', {
      exact: true, name: 'Beaconchain Homepage',
    }).click()
    await expect(page).toHaveURL('/')
  })
})
