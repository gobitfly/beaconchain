import { defineStore } from 'pinia'
import type { LoginResponse } from '~/types/user'
import { useCustomFetch } from '~/composables/useCustomFetch'

export const useUserStore = defineStore('user-store', () => {
  const refreshToken = ref<string | undefined | null>()
  const accessToken = ref<string | undefined | null>()

  async function doLogin (email: string, password: string) {
    const res = await useCustomFetch<LoginResponse>('/login', {
      body: {
        email,
        password
      },
      method: 'POST'
    })
    refreshToken.value = res.refresh_token
    accessToken.value = res.access_token

    return { refreshToken: refreshToken.value, accessToken: accessToken.value }
  }

  return { refreshToken, accessToken, doLogin }
})
