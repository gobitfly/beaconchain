import { defineStore } from 'pinia'
import type { LoginResponse } from '~/types/user'

const userStore = defineStore('user-store', () => {
  const sessionCookie = useCookie('session_id')
  return { data: sessionCookie }
})

export function useUserStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(userStore())

  const sessionCookie = computed(() => data.value)

  async function doLogin (email: string, password: string) {
    await fetch<LoginResponse>(API_PATH.LOGIN, {
      body: {
        email,
        password
      }
    })
  }

  // TODO: Find a way to check if the user is logged in
  const user = computed(() => {
    return sessionCookie.value ? { user_id: sessionCookie, user_name: 'Logged in User' } : undefined
  })

  const isLoggedIn = computed(() => !!user.value)

  return { doLogin, user, isLoggedIn }
}
