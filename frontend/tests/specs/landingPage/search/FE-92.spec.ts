import {
  expect, goto, test,
} from '../../../utils/helpers'

test.describe('verify Search functionality', {
  tag: '@landing-page',
}, () => {
  const validator = '391006'
  const tokens = '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48'
  const address = '0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97'
  const txHash = '0x6ca8e9d835da57932f26ae154da7964651b5fa9aea632a54300cc7727ff1a311'

  test.beforeEach(async ({ page }) => await goto(page, '/products#stacking-hub', 'networkidle'))

  test('verify search filed is visible', async ({ page }) => {
    await expect(page.getByRole('combobox', { name: 'Search by Address / Tx hash / Block / Token / ENS' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Explore Transactions & Validator Staking on Ethereum' })).toBeVisible()
  })

  test('verify typing invalid value wull result to "No results found"', async ({ page }) => {
    const search = page.getByRole('combobox', { name: /Search by Address \/ Tx hash \/ Block \/ Token \/ ENS/i })
    await search.fill('@')
    await expect(page.getByText('No results found.')).toBeVisible()
  })

  test('verify search by address', async ({ page }) => {
    const search = page.getByRole('combobox', { name: /Search by Address \/ Tx hash \/ Block \/ Token \/ ENS/i })
    await search.fill(address)
    await expect(page.getByRole('button', { name: /^Addresses$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^All$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Tx Hashes$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Validator Indices$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Blocks$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Tokens$/ })).toBeVisible()
    const listbox = page.getByRole('listbox')
    await expect(listbox).toBeVisible()
    const options = listbox.getByRole('option')
    await expect(options).toHaveCount(2)
    await expect(listbox.getByRole('group', { name: 'Tx Hashes' })).toHaveCount(0)
    await expect(listbox.getByRole('group', { name: 'Validator Indices' })).toHaveCount(0)
    await expect(listbox.getByRole('group', { name: 'Blocks' })).toHaveCount(0)
    await expect(listbox.getByRole('group', { name: 'Tokens' })).toHaveCount(0)

    await options.nth(0).click()
    await expect(page).toHaveURL(
      /beaconcha\.in\/address\/0x4838b106fce9647bdf1e7877bf73ce8b0bad5f97/i,
    )
    await page.goBack()
    await search.fill(address)
    await listbox.waitFor()
    await options.nth(1).click()
    await expect(page).toHaveURL(
      /hoodi-staging\.beaconcha\.in\/address\/0x4838b106fce9647bdf1e7877bf73ce8b0bad5f97/i,
    )
  })

  test('verify search by Tx Hash', async ({ page }) => {
    const search = page.getByRole('combobox', { name: /Search by Address \/ Tx hash \/ Block \/ Token \/ ENS/i })
    await search.fill(txHash)
    await expect(page.getByRole('button', { name: /^Addresses$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^All$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Tx Hashes$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Validator Indices$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Blocks$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Tokens$/ })).toBeVisible()
    const listbox = page.getByRole('listbox')
    await expect(listbox).toBeVisible()
    const options = listbox.getByRole('option')
    await expect(options).toHaveCount(1)
    await expect(listbox.getByRole('group', { name: 'Addresses' })).toHaveCount(0)
    await expect(listbox.getByRole('group', { name: 'Validator Indices' })).toHaveCount(0)
    await expect(listbox.getByRole('group', { name: 'Blocks' })).toHaveCount(0)
    await expect(listbox.getByRole('group', { name: 'Tokens' })).toHaveCount(0)

    await options.nth(0).click()
    await expect(page).toHaveURL(
      /beaconcha\.in\/tx\/0x6ca8e9d835da57932f26ae154da7964651b5fa9aea632a54300cc7727ff1a311/i,
    )
  })

  test('verify search by Validator/Epoch/Slot/Block', async ({ page }) => {
    const search = page.getByRole('combobox', { name: /Search by Address \/ Tx hash \/ Block \/ Token \/ ENS/i })
    await search.fill(validator)
    await expect(page.getByRole('button', { name: /^Addresses$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^All$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Tx Hashes$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Validator Indices$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Blocks$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Tokens$/ })).toBeVisible()
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

    const slots = groupByLabel('Slots')
    await expect(slots).toBeVisible()
    await expect(slots.getByRole('option')).toHaveCount(2)

    const blocks = groupByLabel('Blocks')
    await blocks.scrollIntoViewIfNeeded()
    await expect(blocks).toBeVisible()
    await expect(blocks.getByRole('option')).toHaveCount(2)

    const epoch = groupByLabel('Epochs')
    await expect(epoch).toBeVisible()
    await expect(epoch.getByRole('option')).toHaveCount(1)

    async function reopen() {
      await search.fill(validator)
      await expect(resultsRoot).toBeVisible()
    }
    // --- Validator ---
    await reopen()
    await groupByLabel('Validator Indices').getByRole('option').nth(0).click()
    await expect(page).toHaveURL(/beaconcha\.in\/validator\/391006/i)
    await page.goBack()

    await reopen()
    await groupByLabel('Validator Indices').getByRole('option').nth(1).click()
    await expect(page).toHaveURL(/https?:\/\/hoodi-staging\.beaconcha\.in\/validator\/391006/i)

    // --- Slots ---
    await page.goBack()
    await reopen()
    await groupByLabel('Slots').getByRole('option').nth(0).click()
    await expect(page).toHaveURL(/beaconcha\.in\/slot\/391006/i)

    await page.goBack()
    await reopen()
    await groupByLabel('Slots').getByRole('option').nth(1).click()
    await expect(page).toHaveURL(/https?:\/\/hoodi-staging\.beaconcha\.in\/slot\/391006/i)

    // --- Blocks ---
    await page.goBack()
    await reopen()
    await groupByLabel('Blocks').getByRole('option').nth(0).click()
    await expect(page).toHaveURL(/beaconcha\.in\/block\/391006/i)

    await page.goBack()
    await reopen()
    await groupByLabel('Blocks').getByRole('option').nth(1).click()
    await expect(page).toHaveURL(/https?:\/\/hoodi-staging\.beaconcha\.in\/block\/391006/i)

    // --- Epochs ---
    await page.goBack()
    await reopen()
    await groupByLabel('Epochs').getByRole('option').nth(0).click()
    await expect(page).toHaveURL(/beaconcha\.in\/epoch\/391006/i)
  })

  test('verify search by Tokens', async ({ page }) => {
    const search = page.getByRole('combobox', { name: /Search by Address \/ Tx hash \/ Block \/ Token \/ ENS/i })
    await search.fill(tokens)
    await expect(page.getByRole('button', { name: /^Addresses$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^All$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Tx Hashes$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Validator Indices$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Blocks$/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /^Tokens$/ })).toBeVisible()
    const resultsRoot = page.locator('div[id^="reka-combobox-content-"]').filter({
      has: page.locator('div[role="group"][id^="reka-combobox-group-"]'),
    })
    await expect(resultsRoot).toBeVisible()
    const groupByLabel = (label: string) =>
      resultsRoot.locator('div[role="group"][id^="reka-combobox-group-"]').filter({
        has: page.getByText(label, { exact: true }),

      })
    const token = groupByLabel('Tokens')
    await expect(token).toBeVisible()
    await expect(token.getByRole('option')).toHaveCount(1)

    await groupByLabel('Tokens').getByRole('option').nth(0).click()
    await expect(page).toHaveURL(/beaconcha\.in\/token\/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48/i)
  })
})
