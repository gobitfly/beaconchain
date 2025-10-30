import {
  expect, goto, test,
} from '../../../utils/helpers'

test.describe('Verify section Explorer', {
  tag: '@landing-page',
}, () => {
  test.beforeEach(async ({ page }) => await goto(page, '/products#stacking-hub', 'networkidle'))

  test('verify Landing Page is opened', async ({ page }) => {
    await expect(page.getByText('Explore Transactions & Validator Staking on')).toBeVisible()
    await expect(page.getByLabel('Explore Transactions &').getByText('ETH')).toBeVisible()
  })

  test('check footer contains all nececcary sections', async ({ page }) => {
    await expect(page.getByRole('link', { name: 'Beaconchain Homepage By' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Services' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Resources' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Links' })).toBeVisible()
    await expect(page.getByRole('contentinfo').getByRole('link', {
      exact: true, name: 'API',
    })).toBeVisible()
    await expect(page.getByRole('contentinfo').getByRole('link', { name: 'StakingHub' })).toBeVisible()
    await expect(page.getByRole('contentinfo').getByRole('link', {
      exact: true, name: 'Explorer',
    })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Advertise' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'beaconcha.in Premium' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Swag Shop' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'API Pricing' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Contact Sales' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Site Status' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Go to our Discord server' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Go to our X profile' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Check out our GitHub' })).toBeVisible()
    await expect(page.getByText('Follow us on social media to find out the latest updates on our progress.')).toBeVisible()
  })

  test('check user can click on each link provided in Services section', async ({ page }) => {
    await page.getByRole('contentinfo').getByRole('link', {
      exact: true, name: 'API',
    }).click()
    await expect(page).toHaveURL(/https:\/\/docs\.beaconcha\.in\/login\?redirect=(?:%2[Ff]|\/)(?:&[^#]*)?$/)
    await page.goBack()

    await page.getByRole('contentinfo').getByRole('link', {
      exact: true, name: 'StakingHub',
    }).click()
    await expect(page).toHaveURL(/\/dashboard$/i)
    await page.goBack()

    await page.getByRole('contentinfo').getByRole('link', {
      exact: true, name: 'Explorer',
    }).click()
    await expect(page).toHaveURL(/\/$/i)
    await page.goBack()
  })

  test('check user can click on each link provided in Resources section', async ({ page }) => {
    await page.getByRole('link', { name: 'Advertise' }).click()
    await expect(page).toHaveURL(/\/advertisewithus$/i)
    await page.goBack()

    await page.getByRole('link', { name: 'beaconcha.in Premium' }).click()
    await expect(page).toHaveURL(/\/premium$/i)
    await page.goBack()

    await page.getByRole('link', { name: 'Swag Shop' }).click()
    await expect(page).toHaveURL(/https:\/\/shop\.beaconcha\.in\/#!\/?/i)
    await page.goBack()

    await page.getByRole('link', { name: 'API Pricing' }).click()
    await expect(page).toHaveURL(/\/products\/pricing$/i)
    await page.goBack()

    await page.getByRole('link', { name: 'Contact Sales' }).click()
    await expect(page).toHaveURL(/\/contact-sales$/i)
    await page.goBack()

    await page.getByRole('link', { name: 'Site Status' }).click()
    await expect(page).toHaveURL(/https:\/\/status\.beaconcha\.in\/?/i)
    await page.goBack()
  })

  test('check user can click on each link provided in Links section', async ({ page }) => {
    await page.getByRole('link', { name: 'Go to our Discord server' }).click()
    await expect(page).toHaveURL(/https:\/\/discord\.com\/invite\/nVGbBnvvnA/i)
    await page.goBack()

    await page.getByRole('link', { name: 'Go to our X profile' }).click()
    await expect(page).toHaveURL(/https:\/\/x\.com\/beaconcha_in\/?/i)
    await page.goBack()

    await page.getByRole('link', { name: 'Check out our GitHub' }).click()
    await expect(page).toHaveURL(/https:\/\/github\.com\/gobitfly/i)
    await page.goBack()
  })

  test('check Imprint, Terms, Privacy section', async ({ page }) => {
    await page.getByRole('link', { name: 'Imprint' }).click()
    await expect(page).toHaveURL(/https:\/\/beaconcha\.in\/imprint$/i)
    await expect(page.getByRole('heading', { name: 'Imprint' })).toBeVisible()
    await page.goBack()

    await page.getByRole('link', { name: 'Terms' }).click()
    await expect(page).toHaveURL(/https:\/\/storage\.googleapis\.com\/legal\.beaconcha\.in\/tos\.pdf$/i)
    await page.goBack()

    await page.getByRole('link', { name: 'Privacy' }).click()
    await expect(page).toHaveURL(/https:\/\/storage\.googleapis\.com\/legal\.beaconcha\.in\/privacy\.pdf$/i)
    await page.goBack()

    await expect(page.getByText('© 2025 beaconcha.in. All rights reserved')).toBeVisible()
  })
})
