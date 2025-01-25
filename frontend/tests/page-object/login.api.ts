import { expect } from '../utils/helpers'
import { config } from '../auth/user-auth'

export async function loginViaAPI(request: any, page: any, payload: any) {
  const response = await request.post('/api/i/login', {
    data: payload,
    headers: config.headers,
  })

  expect(response.ok()).toBeTruthy()
  const browserCookie = config.cookies
  await page.context().addCookies(browserCookie)
  console.log('browserCookie:', browserCookie)
}
