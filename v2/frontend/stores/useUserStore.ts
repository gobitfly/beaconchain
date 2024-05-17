import { defineStore } from 'pinia'
import type { LoginResponse } from '~/types/user'
import type { InternalGetUserInfoResponse, UserInfo } from '~/types/api/user'
import { API_PATH } from '~/types/customFetch'

const userStore = defineStore('user-store', () => {
  const data = ref<UserInfo | undefined | null>()
  return { data }
})

export function useUserStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(userStore())

  async function doLogin (email: string, password: string) {
    await fetch<LoginResponse>(API_PATH.LOGIN, {
      body: {
        email,
        password
      }
    })
    await getUser()
  }

  const setUser = (user?: UserInfo) => {
    data.value = user
  }

  const getUser = async () => {
    try {
      const res = await fetch<InternalGetUserInfoResponse>(API_PATH.USER, undefined, undefined, undefined, true)
      setUser(res.data)
    } catch (e) {
      setUser(undefined)
    }
  }

  const user = computed(() => {
    return data.value
  })

  const isLoggedIn = computed(() => {
    return !!user.value
  })

  return { doLogin, user, isLoggedIn, getUser }
}
