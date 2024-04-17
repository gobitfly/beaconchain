import { defineStore } from 'pinia'
import { warn } from 'vue'
import type { GetUserDashboardsResponse } from '~/types/api/dashboard'
import type { LoginResponse } from '~/types/user'

const userStore = defineStore('user-store', () => {
  const data = ref<{user_id: number, user_name: string} | undefined | null>()
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
  }

  const setUser = (id?: number, name: string = '') => {
    if (!id) {
      data.value = null
    } else {
      data.value = {
        user_id: id,
        user_name: name
      }
    }
  }

  const getUser = async () => {
    try {
      // TODO: replace once we have an endpoint to get a real user
      const res = await fetch<GetUserDashboardsResponse>(API_PATH.USER_DASHBOARDS)
      if (res.data) {
        setUser(1, 'My temp sollution')
      }
    } catch (e) {
      warn('get user error: ', e)
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
