import { LoginPage } from '../../page-object/login.page'
import {
  expect, goto, test,
} from '../../utils/helpers'
import { addDashboardFlow } from '../../page-object/dashboard.page'
import { config } from '../../auth/user-auth'

test.describe('Login', () => {
  test.beforeEach(async ({ page }) => {
    await goto(page, '/login', 'networkidle')
  })

  test('Successful login with valid credentials', async ({ page }) => {
    const userEmail = config.userAuth.guppy.email
    const userPassword = config.userAuth.guppy.password
    await page.waitForLoadState('networkidle')
    await LoginPage.email(page).fill(userEmail)
    await LoginPage.password(page).fill(userPassword)
    await page.waitForLoadState('networkidle')
    await LoginPage.loginBtn(page).click()

    await expect(addDashboardFlow.dashboard(page)).toBeVisible({ timeout: 15000 })
  })
  test('The login button is active after filling in all fields', async ({ page }) => {
    await LoginPage.email(page).fill('testDummydata@bitfly.at')
    await LoginPage.password(page).fill('test19999')

    await expect(LoginPage.loginBtn(page)).toBeEnabled()
  })

  test('Login with space', async ({ page }) => {
    await LoginPage.email(page).fill(' ')
    await LoginPage.password(page).fill(' ')

    await expect(LoginPage.errorEmail(page)).toBeVisible()
    await expect(LoginPage.errorPasswordMaxLength(page)).toBeVisible()
    await expect(LoginPage.loginBtn(page)).toBeDisabled()
  })

  test('Login with an incorrect email format', async ({ page }) => {
    await LoginPage.email(page).fill('testtest@')
    await LoginPage.password(page).fill('test199')

    await expect(LoginPage.errorEmail(page)).toBeVisible()
    await expect(LoginPage.loginBtn(page)).toBeDisabled()
  })

  test('Login with an incorrect password', async ({ page }) => {
    await LoginPage.email(page).fill('tst@bitfly.at')
    await LoginPage.password(page).fill('WrongPsw')
    await LoginPage.loginBtn(page).click()

    await expect(LoginPage.errorInvalidPassword(page)).toBeVisible()
    await expect(LoginPage.toastMessageCannotLogin(page)).toBeVisible()
  })
  test('Restrictions on the minimum password length', async ({ page }) => {
    await LoginPage.email(page).fill('testDummydata@bitfly.at')
    await LoginPage.password(page).fill('123')

    await expect(LoginPage.errorPasswordMaxLength(page)).toBeVisible()
    await expect(LoginPage.loginBtn(page)).toBeDisabled()
  })

  test('Login with unregistered email address', async ({ page }) => {
    await LoginPage.email(page).fill('testDummydata@bitfly.at')
    await LoginPage.password(page).fill('SomePassword')
    await LoginPage.loginBtn(page).click()

    await expect(LoginPage.errorInvalidPassword(page)).toBeVisible()
    await expect(LoginPage.toastMessageCannotLogin(page)).toBeVisible()
  })

  test('Login with case-variant email address', async ({ page }) => {
    await LoginPage.email(page).fill('')
    await LoginPage.password(page).fill('')
    await LoginPage.loginBtn(page).click()

    await expect(LoginPage.errorInvalidPassword(page)).toBeVisible()
    await expect(LoginPage.toastMessageCannotLogin(page)).toBeVisible()
    await page.waitForLoadState('networkidle')
    await LoginPage.loginBtn(page).click()

    await expect(addDashboardFlow.dashboard(page)).toBeVisible({ timeout: 15000 })
  })
})
