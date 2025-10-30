import {
  expect, goto, test,
} from '../../../utils/helpers'
import { LoginPage } from '../../../page-object/login.page'

const email = process.env.USER_EMAIL!
const password = process.env.USER_PASSWORD!
test.describe(
  'Dashboard Tests Unauthorized User',
  {
    tag: '@landing-page',
  },
  () => {
    test.beforeEach(
      async ({ page }) =>
        await goto(page, '/products#stacking-hub', 'networkidle'),
    )

    test('verify Landing Page is opened', async ({ page }) => {
      await expect(
        page.getByText('Seamless access to blockchain data on'),
      ).toBeVisible()
    })

    test('verify Login Button is visible', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Log in' })).toBeVisible()
    })

    test('click on Login Button and verify user redirects to Sign In page', async ({
      page,
    }) => {
      const loginBtn = page.getByRole('button', { name: 'Log in' })
      await loginBtn.focus()
      await page.keyboard.press('Enter')
      await expect(page.getByRole('heading', { name: 'Sign in to beaconcha.in' })).toBeVisible()
      await LoginPage.email(page).fill(email)
      await LoginPage.password(page).fill(password)
      await page.waitForLoadState('networkidle')
      await LoginPage.loginBtn(page).click()
      await expect(page).toHaveURL('/products#stacking-hub')
      await expect(page.getByText('Seamless access to blockchain data on')).toBeVisible()
      await expect(page.getByRole('button', { name: 'Open user menu' })).toBeVisible()
      const userMenuBtn = page.getByRole('button', { name: 'Open user menu' })
      await userMenuBtn.focus()
      await page.keyboard.press('Enter')
      await expect(page).toHaveURL('/user/settings')
      await expect(page.getByText('Account Settings')).toBeVisible()
    })
  },
)
