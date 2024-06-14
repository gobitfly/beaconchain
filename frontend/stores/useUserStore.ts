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
  const router = useRouter()

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

  async function getUser () : Promise<UserInfo|undefined> {
    try {
      const res = await fetch<InternalGetUserInfoResponse>(API_PATH.USER, undefined, undefined, undefined, true)
      if (!res.data) {
        return undefined
      }
      setUser(res.data)
      return res.data
    } catch {
      setUser(undefined)
      return undefined
    }
  }

  const doLogout = async () => {
    await fetch(API_PATH.LOGOUT, undefined, undefined, undefined, true)
    setUser(undefined)
    router.replace('/')
  }

  const user = computed(() => {
    return data.value
  })

  const isLoggedIn = computed(() => {
    return !!user.value
  })

  return { doLogin, doLogout, user, isLoggedIn, getUser }
}
