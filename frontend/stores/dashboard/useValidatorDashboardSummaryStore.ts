import { defineStore } from 'pinia'
import { filter, get, orderBy } from 'lodash-es'
import type { InternalGetValidatorDashboardSummaryResponse, VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'
import { API_PATH } from '~/types/customFetch'
import { getGroupLabel } from '~/utils/dashboard/group'

const validatorDashboardSummaryStore = defineStore('validator_dashboard_sumary_store', () => {
  const data = ref < InternalGetValidatorDashboardSummaryResponse>()
  const query = ref < TableQueryParams>()
  const lastDashboardKey = ref<DashboardKey>()

  return { data, query, lastDashboardKey }
})

export function useValidatorDashboardSummaryStore () {
  const { fetch } = useCustomFetch()
  const { groups } = useValidatorDashboardGroups()
  const { t: $t } = useI18n()

  const { data, query: storedQuery, lastDashboardKey } = storeToRefs(validatorDashboardSummaryStore())
  const isLoading = ref(false)

  const query = computed(() => storedQuery.value)

  const groupNameLabel = (groupId?: number) => {
    return getGroupLabel($t, groupId, groups.value)
  }

  const summary = computed<InternalGetValidatorDashboardSummaryResponse | undefined>(() => {
    if (!data.value) {
      return
    }

    const search = query.value?.search?.toLowerCase()
    let list: VDBSummaryTableRow[] = !search
      ? data.value.data
      : filter(data.value.data, (item) => {
        const name = groupNameLabel(item.group_id).toLowerCase()
        return name.includes(search)
      })
    const [sort, order] = query.value?.sort?.split(':') || [undefined, undefined]
    if (sort && order) {
      if (sort === 'group_id') {
        list = orderBy(list, [item => groupNameLabel(item.group_id).toLowerCase()], order as 'asc' | 'desc')
      } else if (sort.startsWith('efficiency')) {
        list = orderBy(list, [(item) => {
          const path = sort.replace('efficiency_', 'efficiency.')
          return get(item, path)
        }], order as 'asc' | 'desc')
      }
    }
    const totalCount = list.length
    let cursor = query.value?.cursor as (number | undefined)
    const limit = query.value?.limit ?? 10
    if (cursor && cursor as number >= totalCount) {
      cursor = 0
    }
    list = list.slice(cursor, (cursor ?? 0) + limit)
    return {
      paging: {
        total_count: totalCount
      },
      data: list
    }
  }
  )

  async function getSummary (dashboardKey: DashboardKey, query?: TableQueryParams) {
    if (!dashboardKey) {
      data.value = undefined
      return undefined
    }
    storedQuery.value = query
    isLoading.value = true
    if (lastDashboardKey.value !== dashboardKey || !data.value) {
      lastDashboardKey.value = dashboardKey
      const res = await fetch<InternalGetValidatorDashboardSummaryResponse>(API_PATH.DASHBOARD_SUMMARY, undefined, { dashboardKey })

      data.value = res
    }
    isLoading.value = false
    return summary.value
  }

  return { summary, query, isLoading, getSummary, groupNameLabel }
}
