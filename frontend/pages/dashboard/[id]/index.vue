<script lang="ts" setup>
definePageMeta({
  layout: false,
  // middleware: [ 'dashboard' ],
})
// const dashboardData = useDashboardData()

// const elDepositsQueryParams = ref<TableQueryParams>({
//   limit: 5,
//   sort: 'timestamp:desc',
// })
// const {
//   data: elDepositsData,
//   refresh: refreshElDepositsData,
//   status: elDepositsDataStatus,
// } = useAsyncData('el_deposits', () => {
//   return Promise.all([
//     dashboardData.fetchELDeposits(
//       dashboardKey.value,
//       elDepositsQueryParams.value,
//     ),
//     dashboardData.fetchELDpositsTotalAmount(
//       dashboardKey.value,
//       { search: elDepositsQueryParams.value.search },
//     ),
//   ])
// },
// {
//   immediate: false,
//   watch: [ elDepositsQueryParams ],
// })
// const elDeposits = computed(() => {
//   return elDepositsData.value?.[0]
// })
// const elDepositsTotalAmount = computed(() => elDepositsData.value?.[1])

// const clDepositsQueryParams = ref<TableQueryParams>({
//   limit: 5,
//   sort: 'timestamp:desc',
// })
// const {
//   data: clDepositsData,
//   refresh: refreshClDepositsData,
//   status: clDepositsDataStatus,
// } = useAsyncData('cl_deposits', () => {
//   return Promise.all([
//     dashboardData.fetchClDeposits(
//       dashboardKey.value,
//       clDepositsQueryParams.value,
//     ),
//     dashboardData.fetchClDpositsTotalAmount(
//       dashboardKey.value,
//       { search: clDepositsQueryParams.value.search },
//     ),
//   ])
// },
// {
//   immediate: false,
//   watch: [ clDepositsQueryParams ],
// })
// const clDeposits = computed(() => {
//   return clDepositsData.value?.[0]
// })
// const clDepositsTotalAmount = computed(() => clDepositsData.value?.[1])

// const elWithdrawalsQueryParams = ref<TableQueryParams>({
//   limit: 5,
//   sort: 'timestamp:desc',
// })
// const {
//   data: elWithdrawalsData,
//   refresh: refreshElWithdrawalsData,
//   status: elWithdrawalsDataStatus,
// } = useAsyncData('el_withdrawals', () => {
//   return Promise.all([
//     dashboardData.fetchElWithdrawals(
//       dashboardKey.value,
//       elWithdrawalsQueryParams.value,
//     ),
//     dashboardData.fetchElWithdrawalsTotalAmount(
//       dashboardKey.value,
//       { search: elWithdrawalsQueryParams.value.search },
//     ),
//   ])
// },
// {
//   immediate: false,
//   watch: [ elWithdrawalsQueryParams ],
// })
// const elWithdrawals = computed(() => {
//   return elWithdrawalsData.value?.[0]
// })
// const elWithdrawalsTotalAmount = computed(() => elWithdrawalsData.value?.[1])

// const clWithdrawalsQueryParams = ref<TableQueryParams>({
//   limit: 5,
//   sort: 'timestamp:desc',
// })
// const {
//   data: clWithdrawalsData,
//   refresh: refreshClWithdrawalsData,
//   status: clWithdrawalsDataStatus,
// } = useAsyncData('cl_withdrawals', () => {
//   return Promise.all([
//     dashboardData.fetchClWithdrawals(
//       dashboardKey.value,
//       clWithdrawalsQueryParams.value,
//     ),
//     dashboardData.fetchClWithdrawalsTotalAmount(
//       dashboardKey.value,
//       { search: clWithdrawalsQueryParams.value.search },
//     ),
//   ])
// },
// {
//   immediate: false,
//   watch: [ clWithdrawalsQueryParams ],
// })

// const clWithdrawals = computed(() => {
//   return clWithdrawalsData.value?.[0]
// })
// const clWithdrawalsTotalAmount = computed(() => clWithdrawalsData.value?.[1])
// const elConsolidationsQueryParams = ref<TableQueryParams>({
//   limit: 5,
//   sort: 'timestamp:desc',
// })
// const {
//   data: elConsolidationsData,
//   refresh: refreshElConsolidationsData,
//   status: elConsolidationsDataStatus,
// } = useAsyncData('el_consolidations', () => {
//   return dashboardData.fetchElConsolidations(
//     dashboardKey.value,
//     elConsolidationsQueryParams.value,
//   )
// },
// {
//   immediate: false,
//   watch: [ elConsolidationsQueryParams ],
// })
// const elConsolidations = computed(() => {
//   return elConsolidationsData.value || undefined
// })
// const clConsolidationsQueryParams = ref<TableQueryParams>({
//   limit: 5,
//   sort: 'timestamp:desc',
// })
// const {
//   data: clConsolidationsData,
//   refresh: refreshClConsolidationsData,
//   status: clConsolidationsDataStatus,
// } = useAsyncData('cl_consolidations', () => {
//   return dashboardData.fetchClConsolidations(
//     dashboardKey.value,
//     clConsolidationsQueryParams.value,
//   )
// },
// {
//   immediate: false,
//   watch: [ clConsolidationsQueryParams ],
// })
// const clConsolidations = computed(() => {
//   return clConsolidationsData.value || undefined
// })
// const route = useRoute()
// const dashboardId = computed(() => {
//   return (route.params.id as string)
// })

const {
  key,
  updateGroups,
} = useDashboard()
const {
  data: overview,
  refresh: refreshOverview,
} = await useApi(() => `/api/bff/validator-dashboards/${key.value}`, {
  key: 'dashboardOverview',
  watch: [],
})
const {
  data: slotVizEpochs,
  // refresh: refreshSlotViz,
} = await useApi(() => `/api/bff/validator-dashboards/${key.value}/slot-viz`)

const {
  data: privateDashboards,
} = await useApi('/api/bff/users/me/dashboards', {
  getCachedData: (key, nuxtApp) => nuxtApp.payload[key] ?? nuxtApp.payload.data[key],
  key: 'privateDashboards',
  // transform: response => response.validator_dashboards,
})
// const { $api } = useNuxtApp()
// const { data } = await useAsyncData(() => {
//   return Promise.all([
//     $api(`/api/bff/validator-dashboards/${dashboardId.value}`),
//     $api(`/api/bff/validator-dashboards/${dashboardId.value}/slot-viz`),
//   ])
// })
// const overview = computed(() => data.value?.[0] ?? null)
// const slotVizEpochs = computed(() => data.value?.[1] ?? null)

const onChangeValidators = (validators: string[]) => {
  // const encodedValidators = encodeBase64Url(validators.join(','))
  // refreshOverview()
  // refreshSlotViz()
  console.log('👉', validators)

  refreshOverview()
}
</script>

<template>
  <div>
    <DashboardIndex
      :overview
      :slot-viz-epochs
      :validator-dashboards="privateDashboards?.validator_dashboards ?? null"
      @change-validators="onChangeValidators($event)"
      @change-groups="updateGroups($event)"
    />
  </div>
</template>
