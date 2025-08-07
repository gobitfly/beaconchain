<script lang="ts" setup>
import { ERROR_CODE } from '~/shared/utils/helper'

definePageMeta({
  layout: false,
  middleware: [ 'dashboard' ],
})

const route = useRoute()
onBeforeMount(() => {
  const validatorListFromIndexQueryParameter = route.query.index
  if (validatorListFromIndexQueryParameter) {
    navigateTo({
      query: {
        validators: validatorListFromIndexQueryParameter,
      },
      replace: true,
    })
  }
})

const { validators } = useDashboard()

const dashboardId = computed(() => {
  if (validators.value.length) return encodeBase64Url(validators.value.join(','))
  return encodeBase64Url('1')
})

const {
  data: overview,
  error: overviewError,
  // refresh: refreshOverview,
} = useApi(() => `/api/validator-dashboards/${dashboardId.value}`, {
  immediate: validators.value.length > 0,
  key: 'dashboardOverview',
  lazy: true,
  watch: [ dashboardId ],
})

const {
  data: slotVizEpochs,
  // error: slotVizError,
  // refresh: refreshSlotViz,
} = await useApi(() => `/api/validator-dashboards/${dashboardId.value}/slot-viz`, {
  key: 'slotViz',
  transform: (response) => {
    if (validators.value.length) return response
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
  watch: [ dashboardId ],
})

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

const { setValidators } = useDashboard()

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
      @change-validators="onChangeValidators($event)"
    />
  </div>
</template>
