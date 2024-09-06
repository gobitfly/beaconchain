import {
  expect, test,
} from '@nuxt/test-utils/playwright'

test('test public dashboard', async ({
  goto, page,
}) => {
  await goto('/', { waitUntil: 'hydration' })

  // make sure the onboarding dialog is displyed
  await expect(page.getByText('Add a new dashboard')).toBeVisible()

  // navigate to the second page of the wizard
  await page.getByText('Continue').click()
  await expect(page.getByText('What network are your validators on?')).toBeVisible()

  // navigate to the dashboard
  await page.getByText('Continue').click()
  await expect(page.getByText('Online Validators')).toBeVisible()

  // now open the manage validator menu
  await expect(page.getByText('Manage Validators')).toBeVisible()
  await page.getByText('Manage Validators').click()

  // and make sure the modal is displayed
  await expect(page.getByText('Add or remove validators from your dashboard')).toBeVisible()

  // search for validator number 5
  await page.getByPlaceholder('Index, Public key, Deposit or Withdrawal address or ENS, Graffiti, ...').press('5')

  // add the validator to the dashboard
  await page.getByPlaceholder('Index, Public key, Deposit or Withdrawal address or ENS, Graffiti, ...').press('Enter')

  // make sure the validator is added
  await expect(page.getByText('0x9699')).toBeVisible()

  // close the modal
  await page.getByText('Done').click()

  // make sure the default group is displayed
  await expect(page.getByRole('cell', { name: 'Default' })).toBeVisible()

  // save a screenshot at the end
  await page.screenshot({ path: 'test-results/screenshot.png' })
})
