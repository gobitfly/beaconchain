import { defineStore } from 'pinia'
import type { LoginResponse } from '~/types/user'

export const useUserStore = defineStore('user-store', () => {
  const { public: { xUserId } } = useRuntimeConfig()
  const { fetch } = useCustomFetch()

  async function doLogin (email: string, password: string) {
    await fetch<LoginResponse>(API_PATH.LOGIN, {
      body: {
        email,
        password
      }
    })
  }

  // TODO: Faking logged in User for now, if xUserId is set
  const user = computed(() => {
    return xUserId ? { user_id: xUserId, user_name: `Test User [${xUserId}]` } : undefined
  })

  const isLoggedIn = computed(() => !!user.value)

  return { doLogin, user, isLoggedIn }
})
