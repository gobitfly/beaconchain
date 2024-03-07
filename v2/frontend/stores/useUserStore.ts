import { defineStore } from 'pinia'
import type { LoginResponse } from '~/types/user'
import { useCustomFetch } from '~/composables/useCustomFetch'

export const useUserStore = defineStore('user-store', () => {
  async function doLogin (email: string, password: string) {
    await useCustomFetch<LoginResponse>(API_PATH.LOGIN, {
      body: {
        email,
        password
      }
    })
  }

  return { doLogin }
})
