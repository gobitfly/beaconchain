import {
  expect, goto, test,
} from '../../../utils/helpers'

test.describe('verify Search filters', {
  tag: '@landing-page',
}, () => {
  const validator = '391006'

  test.beforeEach(async ({ page }) => await goto(page, '/products#stacking-hub', 'networkidle'))

  test('verify filters work correctly', async ({ page }) => {
    const search = page.getByRole('combobox', { name: /Search by Address \/ Tx hash \/ Block \/ Token \/ ENS/i })
    await search.fill(validator)
    await expect(page.getByRole('button', { name: /^All$/ })).toHaveAttribute('aria-pressed', 'true')
    await page.getByRole('button', { name: /^Addresses$/ }).click()
    await expect(page.getByText('No results found.')).toBeVisible()
    await expect(page.getByRole('button', { name: /^All$/ })).toHaveAttribute('aria-pressed', 'false')
    await page.getByRole('button', { name: /^Tx Hashes$/ }).click()
    await expect(page.getByText('No results found.')).toBeVisible()
    await page.getByRole('button', { name: /^Validator Indices$/ }).click()
    const resultsRoot = page.locator('div[id^="reka-combobox-content-"]').filter({
      has: page.locator('div[role="group"][id^="reka-combobox-group-"]'),
    })
    await expect(resultsRoot).toBeVisible()
    const groupByLabel = (label: string) =>
      resultsRoot.locator('div[role="group"][id^="reka-combobox-group-"]').filter({
        has: page.getByText(label, { exact: true }),

      })
    const validators = groupByLabel('Validator Indices')
    await expect(validators).toBeVisible()
    await expect(validators.getByRole('option')).toHaveCount(2)
    await expect(page.getByRole('button', { name: /^Tx Hashes$/ })).toHaveAttribute('aria-pressed', 'true')
    await expect(page.getByRole('button', { name: /^Addresses$/ })).toHaveAttribute('aria-pressed', 'true')
    await expect(page.getByRole('button', { name: /^Validator Indices$/ })).toHaveAttribute('aria-pressed', 'true')
    await page.getByRole('button', { name: /^Blocks$/ }).click()
    await page.getByRole('button', { name: /^Validator Indices$/ }).click()

    const blocks = groupByLabel('Blocks')
    await blocks.scrollIntoViewIfNeeded()
    await expect(blocks).toBeVisible()
    await expect(blocks.getByRole('option')).toHaveCount(2)

    await page.getByRole('button', { name: /^All$/ }).click()

    await expect(page.getByRole('button', { name: /^Tx Hashes$/ })).toHaveAttribute('aria-pressed', 'false')
    await expect(page.getByRole('button', { name: /^Addresses$/ })).toHaveAttribute('aria-pressed', 'false')
    await expect(page.getByRole('button', { name: /^Validator Indices$/ })).toHaveAttribute('aria-pressed', 'false')
    await expect(page.getByRole('button', { name: /^Blocks$/ })).toHaveAttribute('aria-pressed', 'false')
    await expect(page.getByRole('button', { name: /^Tokens$/ })).toHaveAttribute('aria-pressed', 'false')
    await expect(page.getByRole('button', { name: /^All$/ })).toHaveAttribute('aria-pressed', 'true')
  })
})
