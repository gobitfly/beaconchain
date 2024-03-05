import { defineStore } from 'pinia'
import type { LoginResponse } from '~/types/user'
import { useCustomFetch } from '~/composables/useCustomFetch'

export const useUserStore = defineStore('user-store', () => {
  const { public: { xUserId } } = useRuntimeConfig()

  async function doLogin (email: string, password: string) {
    await useCustomFetch<LoginResponse>(API_PATH.LOGIN, {
      body: {
        email,
        password
      },
      method: 'POST'
    })
  }

  const user = computed(() => {
    return xUserId ? { user_id: xUserId, user_name: `Test User [${xUserId}]` } : undefined
  })

  const isLoggedIn = computed(() => !!user.value)

  return { doLogin, user, isLoggedIn }
})
