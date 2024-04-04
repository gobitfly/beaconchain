import { defineStore } from 'pinia'
import type { LoginResponse } from '~/types/user'

const userStore = defineStore('user-store', () => {
  const { public: { xUserId } } = useRuntimeConfig()
  const data = ref(xUserId ?? '')
  return { data }
})

export function useUserStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(userStore())

  const xUserId = computed(() => data?.value)

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
    return xUserId.value ? { user_id: xUserId, user_name: `Test User [${xUserId}]` } : undefined
  })

  const isLoggedIn = computed(() => !!user.value)

  return { doLogin, user, isLoggedIn }
}
