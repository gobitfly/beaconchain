import type { NumberOrString } from '~/types/value'
import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'

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

  // const data = useFetchedData('validators')

  /**
   * TODO: fix for private dashboards
   */
  const validatorIds = computed(() => {
    if (validatorList.value) return validatorList.value.split(',')
    return []
  })
  const hasValidators = computed(() => validatorIds.value.length > 0)

  /**
   * Integer (private Dashboard) or base64url encoded list of `validator id`s or `validator public key`s
   */
  const key = computed(() => {
    const id = computed(() => route.params.id as string | undefined)
    if (id.value?.length) return id.value
    if (validatorList.value) return validatorListEncoded.value
    return undefined
  })

  const navigateToDashboard = async (id: number | string) => {
    await navigateTo({
      name: 'dashboard-id',
      params: { id },
    })
  }
  /**
   * Pushes route to page with new query parameters for `?validatorIds=`
   */
  const setValidators = async (list: NumberOrString[]) => {
    const sortedList = list.sort((a, b) => `${a}`.localeCompare(`${b}`, 'en', { numeric: true }))
    await router.push({
      query: {
        ...route.query,
        validators: sortedList.length ? sortedList.join(',') : undefined,
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
  const { t: $t } = useTranslation()
  const name = computed(() => overview.value?.name ?? $t('dashboard.public_validator_dashboard'))

  const { validatorDashboards } = usePrivateDashboards()
  const currentDashboard = computed(() => {
    return validatorDashboards.value.find(dashboard => `${dashboard.id}` === key.value)
  })
  const publicId = computed(() => {
    // currently only one public id is supported
    return currentDashboard.value?.public_ids?.[0]?.public_id
  })
  const publicName = computed(() => {
    // currently only one public id is supported
    return currentDashboard.value?.public_ids?.[0]?.name
  })
  const updateGroups = (newGroups: VDBOverviewGroup[]) => {
    if (!overview.value) return

    overview.value.groups = newGroups
  }

  return {
    /**
     * number of seconds the user is allowed to query the chart history
     * into the past (depending on the user's tier)
     */
    chartHistorySeconds,
    groups,
    hasValidators,
    /**
     * arbitrary value that was defined by the product team
     */
    isLargeDashboard: computed(() => totalValidators.value > 64),
    key,
    name,
    // isPrivateDashboard,
    // isSharedDashboard,
    navigateToDashboard,
    /**
     * A string starting with `v-`, which is basically an alias for the private dashboard id
     * it belongs to.
     */
    publicId,
    publicName,
    setValidators,
    totalValidators,
    updateGroups,
    validatorIds,
    //   /**
    //  * List of `validator-id`s or `validator public key`s from the `?validators=` query parameter
    //  */
    //   validatorList,
    /**
     * Base64UrlEncoded list of `validator-id`s or `validator public key`s from the `?validators=` query parameter
     * @example 1,2 -> `MSwy`
     */
    validatorListEncoded,
    variant,
  }
}
