import {
  expect, goto, test,
} from '../../../utils/helpers'

test.describe('Section StakingHub', {
  tag: '@landing-page',
}, () => {
  test.beforeEach(async ({ page }) => await goto(page, '/products#stacking-hub', 'networkidle'))

  test('Verify Go to StakingHub link', async ({ page }) => {
    await page.getByRole('link', { name: 'Go to StakingHub' }).click()
    await expect(page).toHaveURL(/\/dashboard$/i)
    await page.goBack()
  })

  test('Verify the Set up notifications link', async ({ page }) => {
    await page.getByRole('link', { name: 'Set up notifications' }).click()
    await expect(page).toHaveURL(/\/notifications(?:#.*)?$/i)
    await page.goBack()
  })

  test('Verify App Store link', async ({
    context,
    page,
  }) => {
    const link = page.getByRole('link', { name: 'Download the Beaconchain Dashboard App on the App Store' })
    await expect(link).toHaveAttribute(
      'href',
      /apps\.apple\.com\/([a-z]{2}(?:-[A-Z]{2})?\/)?app\/beaconchain-dashboard\/id1541822121/i,
    )

    const [ newPage ] = await Promise.all([
      context.waitForEvent('page'),
      link.click(),
    ])

    await newPage.waitForLoadState('domcontentloaded')

    await expect(newPage).toHaveURL(
      /apps\.apple\.com\/.*\/app\/beaconchain-dashboard\/id1541822121/i,
    )

    await newPage.close()
  })

  test('Verify Google Play link', async ({
    context,
    page,
  }) => {
    const link = page.getByRole('link', { name: 'Download the Beaconchain Dashboard App on Play Store' })
    await expect(link).toHaveAttribute('href', /play\.google\.com\/store\/apps\/details\?id=in\.beaconcha\.mobile/i)

    const [ newPage ] = await Promise.all([
      context.waitForEvent('page'),
      link.click(),
    ])

    await newPage.waitForLoadState('domcontentloaded')

    await expect(newPage).toHaveURL(/play\.google\.com\/store\/apps\/details.*id=in\.beaconcha\.mobile/i)

    await newPage.close()
  })

  test('Verify dashboard creation link.', async ({ page }) => {
    await page.getByRole('link', { name: 'Create your Dashboard' }).click()
    await expect(page).toHaveURL('/dashboard')
    await page.goBack()
  })

  test('Verify all text for StakingHub section is visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Comprehensive staking insights on' })).toBeVisible()
    await expect(page.getByText('Stay ahead of your validators—get alerts for validator duties, and monitor their performance in real time.')).toBeVisible()
    await expect(page.getByRole('link', { name: 'Download the Beaconchain Dashboard App on the App Store' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Download the Beaconchain Dashboard App on Play Store' })).toBeVisible()
    await expect(page.getByText('Keeping the network secure is hard')).toBeVisible()
    await expect(page.getByText('Let us help you with that.Easily check on your validators with our Beaconchain Dashboard app. Wherever you are, whenever you want.')).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Validator Dashboards' })).toBeVisible()
    await expect(page.getByText('Keeping track of your validators performance is tricky')).toBeVisible()
    await expect(page.getByText('Let us make it simple. Add your validators to a personal dashboard and get clear insights into rewards, attestations, and missed duties — all in one place.')).toBeVisible()
  })
})
