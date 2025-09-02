import type {
  InternalGetUserInfoResponse, UserInfo,
} from '~/types/api/user'

const userStore = defineStore('user-store', () => {
  const data = ref<null | undefined | UserInfo>()
  return { data }
})

export function useUserStore() {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(userStore())
  const router = useRouter()

  const setUser = (user?: UserInfo) => {
    data.value = user
  }

  async function getUser() {
    try {
      const res = await fetch<InternalGetUserInfoResponse>(
        'USER',
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
    await fetch('LOGOUT')
    setUser(undefined)
    router.replace('/')
  }

  const user = computed(() => {
    return data.value
  })

  const hasV1Notifications = computed(() => {
    return !!user.value?.has_v1_notifications
  })

  const isLoggedIn = computed(() => {
    return !!user.value
  })

  const premium_perks = computed(() => user.value?.premium_perks)

  const hasAbilityRewardsChartHistory = computed(() => ({
    daily: (user.value?.premium_perks.rewards_chart_history_seconds?.daily ?? 0) > 0,
    epoch: (user.value?.premium_perks.rewards_chart_history_seconds?.epoch ?? 0) > 0,
    hourly: (user.value?.premium_perks.rewards_chart_history_seconds?.hourly ?? 0) > 0,
    weekly: (user.value?.premium_perks.rewards_chart_history_seconds?.weekly ?? 0) > 0,
  }))

  const hasAbilityChartHistory = computed(() => ({
    daily: (user.value?.premium_perks.chart_history_seconds?.daily ?? 0) > 0,
    epoch: (user.value?.premium_perks.chart_history_seconds?.epoch ?? 0) > 0,
    hourly: (user.value?.premium_perks.chart_history_seconds?.hourly ?? 0) > 0,
    weekly: (user.value?.premium_perks.chart_history_seconds?.weekly ?? 0) > 0,
  }))

  return {
    doLogout,
    getUser,
    hasAbilityChartHistory,
    hasAbilityRewardsChartHistory,
    hasV1Notifications,
    isLoggedIn,
    premium_perks,
    user,
  }
}
