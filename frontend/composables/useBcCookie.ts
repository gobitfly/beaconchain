import {
  type CookieOptions,
  useCookie,
} from '#app/composables/cookie'

type CookieName =
  | 'bc-account-dashboard-key'
  | 'bc-age-format'
  | 'bc-cookies-preference'
  | 'bc-user-dashboards'
  | 'bc-validator-dashboard-key'

// for now without the other `type overload` there is no way to use
// `readonly` feature of `useCookieNuxt` (we might adapt this if needed)
type OptionsUseCookie<T> = CookieOptions<T> & {
  readonly?: false,
}

/**
 * A wrapper around the Nuxt `useCookie` composable to manage application-specific cookies.
 * This allows us to have autocompletion for the cookie names.
 *
 */
export const useBcCookie = <T = string | undefined>(
  name: CookieName,
  options?: OptionsUseCookie<T>,
) => {
  // https://developer.chrome.com/blog/cookie-max-age-expires
  const maxDaysForCookies = 400
  // eslint-disable-next-line no-restricted-syntax
  return useCookie(name, {
    maxAge: getSeconds({ days: maxDaysForCookies }),
    ...options,
  })
}
