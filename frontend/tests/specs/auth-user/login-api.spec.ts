import { test } from '../../utils/helpers'
import { config } from '../../auth/user-auth'
import { loginViaAPI } from '~/tests/page-object/login.api'

test('should login via API and visit dashboard successfully', async ({
  page, request,
}) => {
  const guppyPassword = config.userAuth.guppy
  await loginViaAPI(request, page, guppyPassword)
  await page.goto('/dashboard/#summary')
  const context = page.context()
  const cookies = await context.cookies()
  console.log('Cookies:', cookies)
})
