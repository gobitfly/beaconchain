import type {
  InternalGetUserInfoResponse, UserInfo,
} from '~/types/api/user'
import { API_PATH } from '~/types/customFetch'

const userStore = defineStore('user-store', () => {
  const data = ref<null | undefined | UserInfo>()
  return { data }
})

export function useUserStore() {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(userStore())
  const router = useRouter()

  async function doLogin(email: string, password: string) {
    await fetch(API_PATH.LOGIN, {
      body: {
        email,
        password,
      },
    })
    await getUser()
  }

  const setUser = (user?: UserInfo) => {
    data.value = user
  }

  async function getUser() {
    try {
      const res = await fetch<InternalGetUserInfoResponse>(
        API_PATH.USER,
      )
      setUser(res.data)
      return res.data
    }
    catch {
      setUser(undefined)
      return null
    }
  }

  const doLogout = async () => {
    await fetch(API_PATH.LOGOUT)
    setUser(undefined)
    router.replace('/')
  }

  const user = computed(() => {
    return data.value
  })

  const isLoggedIn = computed(() => {
    return !!user.value
  })

  return {
    doLogin,
    doLogout,
    getUser,
    isLoggedIn,
    user,
  }
}
