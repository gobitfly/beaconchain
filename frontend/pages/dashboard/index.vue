<script lang="ts" setup>
import { ERROR_CODE } from '~/shared/utils/helper'

definePageMeta({
  layout: false,
  // middleware: [ 'dashboard' ],
})

const route = useRoute()
const validatorListFromIndexQueryParameter = route.query.index
if (validatorListFromIndexQueryParameter) {
  await navigateTo({
    query: {
      validators: validatorListFromIndexQueryParameter,
    },
    replace: true,
  })
}

const {
  key,
  setValidators,
  validatorIds,
} = useDashboard()

const {
  data: overview,
  error: overviewError,
  // refresh: refreshOverview,
} = useApi(() => `/api/bff/validator-dashboards/${key.value}`, {
  immediate: !!key.value,
  key: 'dashboardOverview',
  lazy: true,
  watch: [ key ],
})

const {
  data: slotVizEpochs,
  // error: slotVizError,
  // refresh: refreshSlotViz,
} = await useApi(() => `/api/bff/validator-dashboards/${key.value || encodeBase64Url('1')}/slot-viz`, {
  key: 'slotViz',
  transform: (response) => {
    if (validatorIds.value.length) return response
    // We use this hacky solution as we don't have an api endpoint to load a slot viz without validators
    // So we load it for a small guest dashboard and then remove the validator informations from it.
    const filteredSlotVizEpochs = response.map(epoch => ({
      ...epoch,
      slots: epoch.slots?.map(({
        slot,
        status,
      }) => ({
        slot,
        status,
      })),
    }))
    return filteredSlotVizEpochs
  },
  watch: [ key ],
})

// using useFetch instead of useApi here, as transform() is altering the return type of `data`
// which currently results in an `typescript error`
const {
  data: privateDashboards,
} = await useApi('/api/bff/users/me/dashboards', {
  getCachedData: (key, nuxtApp) => nuxtApp.payload[key] ?? nuxtApp.payload.data[key],
  key: 'privateDashboards',
  // transform: response => response.validator_dashboards,
})
// const {
//   data: validatorDashboards,
// } = await useApi('/api/bff/users/me/dashboards', {
//   getCachedData: (key, nuxtApp) => nuxtApp.payload[key] ?? nuxtApp.payload.data[key],
//   transform: response => response.validator_dashboards,
// })

onMounted(() => {
  if (overviewError.value?.statusMessage === ERROR_CODE.EFFECTIVE_BALANCE_EXCEEDS_LIMIT) {
    return showError({
      statusCode: overviewError.value.statusCode,
      statusMessage: ERROR_CODE.EFFECTIVE_BALANCE_EXCEEDS_LIMIT,
    })
  }
  if (overviewError.value?.statusCode === 400) {
    return showError({
      statusCode: overviewError.value.statusCode,
      statusMessage: ERROR_CODE.GUEST_DASHBOARD_ID_INVALID,
    })
  }
  if (overviewError.value) {
    return showError({
      statusCode: overviewError.value?.statusCode,
      statusMessage: overviewError.value?.statusMessage,
    })
  }
})

const onChangeValidators = async (newValidators: string[]) => {
  await setValidators(newValidators)
}

// const onClick = () => {
//   useRouter().push({
//     query: {
//       validators: '1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22',
//     },
//   })
//   // dashboardCreationControllerModal.value?.show()
// }
// const test = ref(1)
// const { data: fetchedData } = useNuxtData('test')
// const {
//   data,
//   // refresh,
// } = useFetch(() => `https://jsonplaceholder.typicode.com/todos/${test.value}`, {
//   watch: [ test ],
//   // getCachedData: () => undefined,
//   // key: 'test',
//   // default: () => ({
//   //   completed: false,
//   //   id: 1,
//   //   title: 'this is a test',
//   //   userId: 1,
//   // }),
//   // default: ,
// })
// const {
//   data,
//   // refresh,
// } = useAsyncData('test', () => $fetch(`https://jsonplaceholder.typicode.com/todos/${test.value}`), {
//   // watch: [ test ],
//   // getCachedData: () => undefined,
//   // key: 'test',
//   // default: () => ({
//   //   completed: false,
//   //   id: 1,
//   //   title: 'this is a test',
//   //   userId: 1,
//   // }),
//   // default: ,
// })
// const { $api } = useNuxtApp()
</script>

<template>
  <div>
    <DashboardIndex
      :overview
      :slot-viz-epochs
      :validator-dashboards="privateDashboards?.validator_dashboards ?? null"
      @change-validators="onChangeValidators($event)"
    />
  </div>
</template>
