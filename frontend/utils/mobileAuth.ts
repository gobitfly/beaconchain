import type { LocationQuery } from 'vue-router'

// Provides mobile auth related GET uri params to the the new target route path.
// Reasoning is that a mobile login request must be handled by mobile, if a user clicks on register or forgot password
// during mobile login, they still must be forwarded to the app in the end.
// This must be implemented on authentication pages like login, register, forgot password
export function provideMobileAuthParams (queryParams: LocationQuery, path: string) {
  if (queryParams.redirect_uri && queryParams.device_id && queryParams.device_name) {
    const redirectURI = queryParams.redirect_uri
    const state = queryParams.state || '' // optional
    const deviceID = queryParams.device_id
    const deviceName = queryParams.device_name

    return {
      path,
      query: {
        redirect_uri: redirectURI,
        state,
        device_id: deviceID,
        device_name: deviceName
      }
    }
  }
  return {
    path
  }
}

// Call after a successfull authenticatio to check whethr the request originated from
// the mobile app and if so handle the request accordingly. Otherwise the method does nothing and returns false
export function handleMobileAuth (queryParams: LocationQuery): boolean {
  if (queryParams.redirect_uri && queryParams.device_id && queryParams.device_name) {
    const state = queryParams.state ? '&state=' + queryParams.state! : ''

    // We need to navigate to this api url instead of fetching it because the response of this endpoint
    // is always a redirect. This redirect must be followed by the browser, which in turn opens the native
    // app that is listening on the redirect callback uri
    const runtimeConfig = useRuntimeConfig()
    const { public: { apiClient } } = runtimeConfig
    window.location.href = apiClient + '/mobile/authorize?redirect_uri=' + queryParams.redirect_uri + '&device_id=' + queryParams.device_id + '&device_name=' + queryParams.device_name + state
    return true
  }
  return false
}
