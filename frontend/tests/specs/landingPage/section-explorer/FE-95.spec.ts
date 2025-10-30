import {
  expect, goto, test,
} from '../../../utils/helpers'

test.describe('Verify section Explorer', {
  tag: '@landing-page',
}, () => {
  test.beforeEach(async ({ page }) => await goto(page, '/products#stacking-hub', 'networkidle'))

  test('verify Landing Page is opened', async ({ page }) => {
    await expect(page.getByText('Seamless access to blockchain data on')).toBeVisible()
    await expect(page.getByLabel('Seamless access to blockchain').getByText('ETH')).toBeVisible()
  })

  test('verify user see go to Explorer btn and can click on it', async ({ page }) => {
    await expect(page.getByRole('link', { name: 'Go to Explorer' })).toBeVisible()
    await page.getByRole('link', { name: 'Go to Explorer' }).click()
    await expect(page).toHaveURL('/')
  })

  test('Verify all text for Explorer section is visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Comprehensive staking insights on' })).toBeVisible()
    await expect(page.getByLabel('Comprehensive staking').getByText('ETH')).toBeVisible()
    await expect(page.getByText('The fastest way to look up blocks, transactions, wallets and smart contracts')).toBeVisible()
    await expect(page.getByText('Transaction tracking')).toBeVisible()
    await expect(page.getByText('Track transactions in real time')).toBeVisible()
    await expect(page.getByText('Wallet & Token Insights')).toBeVisible()
    await expect(page.getByText('View wallet balances & token holdings')).toBeVisible()
    await expect(page.getByText('Smart Contract Analysis')).toBeVisible()
    await expect(page.getByText('Analyze smart contract interactions')).toBeVisible()
    await expect(page.getByText('Gas Fee Monitoring')).toBeVisible()
    await expect(page.getByText('Monitor network status & fees')).toBeVisible()
  })
})
