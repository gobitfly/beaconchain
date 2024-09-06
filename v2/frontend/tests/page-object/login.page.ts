import type { Page } from '../utils/helpers'

export const LoginPage = {
  email: (page: Page) => page.getByLabel('Email address'),
  errorEmail: (page: Page) => page.locator('text=Please provide a valid email address.'),
  errorInvalidPassword: (page: Page) => page.locator('text=Please enter your password.'),
  errorPasswordMaxLength: (page: Page) => page.locator('text=Please provide at least 5 characters.'),
  loginBtn: (page: Page) => page.getByRole('button', { name: 'Log in' }),
  password: (page: Page) => page.getByLabel('Password'),
  toastMessageCannotLogin: (page: Page) => page.locator('text=Cannot log in: your email or your password is unknown.'),
}
