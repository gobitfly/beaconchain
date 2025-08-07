import type { NumberOrString } from '~/types/value'

export const useDashboard = () => {
  const route = useRoute()
  const router = useRouter()
  // const cookie = useBcCookie('bc-validator-dashboard-key')

  /**
   * string of comma separated `validator id`s or `validator public key`s
   *
   *  @example `1,2,3` or `0x80445169...,0x8000...`
   */
  const validatorList = computed(() => {
    if (typeof route.query.validators === 'string') return route.query.validators
    return ''
  })
  const validatorListEncoded = computed(() => encodeBase64Url(validatorList.value))
  /**
   * TODO: fix for private dashboards
   */
  const validators = computed(() => {
    if (validatorList.value) return validatorList.value.split(',')
    return []
  })
  /**
   * Integer (private Dashboard) or base64url encoded list of `validator id`s or `validator public key`s
   */
  const key = computed(() => {
    const id = route.params.id as string | undefined
    if (id?.length) return id
    if (validatorList.value) return encodeBase64Url(validatorList.value)
    return undefined
  })
  // const isSharedDashboard = computed(() => id.value?.startsWith('v-'))
  const navigateToDashboard = async (id: string) => {
    await navigateTo({
      name: 'dashboard-id',
      params: { id },
    })
  }

  const hasValidators = computed(() => validators.value.length > 0)

  /**
   * Pushes route to page with new query parameters for `?validators=`
   */
  const setValidators = async (list: NumberOrString[]) => {
    const sortedList = list.sort((a, b) => `${a}`.localeCompare(`${b}`, 'en', { numeric: true }))
    await router.push({
      query: {
        ...route.query,
        validators: sortedList.join(','),
      },
    })
  }

  const variant = computed(() => {
    if (key.value?.startsWith('v-')) return 'shared-dashboard'
    if (isInteger(key.value ?? '')) return 'private-dashboard'
    // if (validators.value.length) return 'guest-dashboard'
    // return undefined
    return 'guest-dashboard'
  })
  // const isPrivateDashboard = computed(() => {
  //   if(!dashboardId.value.length)
  // })

  const overview = useFetchedData('dashboardOverview')
  const totalValidators = computed(() => {
    return Object.values(overview.value?.validators ?? {})
      .reduce((previousValue, currentValue) => previousValue + currentValue, 0)
  })

  const chartHistorySeconds = computed(() => overview.value?.chart_history_seconds ?? {
    daily: 0,
    epoch: 0,
    hourly: 0,
    weekly: 0,
  })

  const groups = computed(() => overview.value?.groups ?? [])

  return {
    /**
     * number of seconds the user is allowed to query the chart history
     * in the past (depending on user's tier)
     */
    chartHistorySeconds,
    groups,
    hasValidators,
    /**
     * arbitrary value that was defined by the product team
     */
    isLargeDashboard: computed(() => totalValidators.value > 64),
    key,
    // isPrivateDashboard,
    // isSharedDashboard,
    navigateToDashboard,
    setValidators,
    totalValidators,
    //   /**
    //  * List of `validator-id`s or `validator public key`s from the `?validators=` query parameter
    //  */
    //   validatorList,
    /**
     * Base64UrlEncoded list of `validator-id`s or `validator public key`s from the `?validators=` query parameter
     * @example 1,2 -> `MSwy`
     */
    validatorListEncoded,
    validators,
    variant,
  }
}
