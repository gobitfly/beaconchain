import {
  expect, goto, test,
} from '../../../utils/helpers'

test.describe('Section API', {
  tag: '@landing-page',
}, () => {
  test.beforeEach(async ({ page }) => await goto(page, '/products#stacking-hub', 'networkidle'))

  test('Verify all text for API section is visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Unified API to access data on' })).toBeVisible()
    await expect(page.getByLabel('Unified API to access data on').getByText('ETH')).toBeVisible()
    await expect(page.getByText('Power your apps with unified')).toBeVisible()
    await expect(page.getByRole('link', { name: 'Go to API' })).toBeVisible()
    await expect(page.getByText('Unlock the power of multichain with our comprehensive docs. Clear guides to power your blockchain apps.')).toBeVisible()
    await expect(page.getByRole('heading', { name: 'API Docs' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Explorer API' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'StakingHub API' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Resource Management API' })).toBeVisible()
    await expect(page.getByText('Real-time transactions and blocks')).toBeVisible()
    await expect(page.getByText('Historical gas prices and fees')).toBeVisible()
    await expect(page.getByText('Address and contract analytics')).toBeVisible()
    await expect(page.getByText('Validator performance metrics')).toBeVisible()
    await expect(page.getByText('Real-time staking rewards data')).toBeVisible()
    await expect(page.getByText('Validator status and lifecycle events')).toBeVisible()
    await expect(page.getByText('Add/remove dashboards')).toBeVisible()
    await expect(page.getByText('Configure validator notifications')).toBeVisible()
    await expect(page.getByRole('heading', { name: 'API Pricing Plan' })).toBeVisible()
    await expect(page.getByText('Start for free or choose a plan that suits you best')).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Staking & Multi-EVM Data Made' })).toBeVisible()
    await expect(page.getByText('Access validator stats, rewards, and onchain data with one powerful API.')).toBeVisible()
    await expect(page.getByRole('link', { name: 'Get your API Key' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Compare Plans' })).toBeVisible()
  })

  test('Verify all links navigates to the correct page', async ({ page }) => {
    await page.getByRole('link', { name: 'Go to API' }).click()
    await expect(page).toHaveURL('products/pricing')
    await page.goBack()
    await page.getByRole('link').filter({ hasText: /^$/ }).click()
    await expect(page).toHaveURL(/https:\/\/docs\.beaconcha\.in\/login\?redirect=(?:%2[Ff]|\/)(?:&[^#]*)?$/)
    await page.goBack()
    await page.getByRole('link', { name: 'Compare Plans' }).click()
    await expect(page).toHaveURL('/products/pricing')
    await page.goBack()
    await page.getByRole('link', { name: 'Get your API Key' }).click()
    await expect(page).toHaveURL('/products/api/key-management')
    await page.goBack()
  })
})
