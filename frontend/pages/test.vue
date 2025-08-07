<script setup lang="ts">
// const dashboardId = ref(25)
// const { $api } = useNuxtApp()
// const {
//   data,
//   error,
//   refresh,
//   status,
// // } = useFetch('/api/latest-state', {
// } = useApi('/api/latest-state', {
// })

// const response = $api('/api/validator-dashboards/:dashboardId/consensus-layer-deposits')
// const response = await $api('/api/latest-state')
// const response = await $api({})

// const requestFetch = useRequestFetch()
// const {
//   data,
//   error,
//   refresh,
//   status,
// } = await useAsyncData(() => requestFetch('/api/validator-dashboards/25/consensus-layer-deposits'))
// const input = ref('')
// const router = useRouter()
// const updateQueryParam = (value: InputEvent) => {
//   router.push({
//     query: {
//       ...router.currentRoute.value.query,
//       validators: value.data,
//     },
//   })
// }
// const test = useFetchedData('user')
const route = useRoute()
const router = useRouter()
const counter = ref(0)
// const { data } = useApi(() => `/api/test/${counter.value}`, {
//   // baseURL: 'https://v2-staging-hoodi.beaconcha.in/api/i',
//   // query: {
//   //   counter,
//   // },
//   // watch: [ () => route.query ],
// })
const nuxtApp = useNuxtApp()
const key = ref('MSwyLDMsNCw1LDYsNyw4LDksMTAsMTEsMTIsMTMsMTQsMTUsMTYsMTcsMTgsMTk')
const {
  data,
  error,
  refresh,
// } = useApi(() => `/api/validator-dashboards/${key.value}`, {
} = useApi(() => '/api/validator-dashboards/22', {
  // getCachedData: key => nuxtApp.payload.state[key] ?? nuxtApp.payload.data[key],
  // body: {
  //   dashboardKey: key.value,
  //   groupId: 0,
  // },
  // immediate: true,
  // key: 'dashboardSummaryDetails',
  query: {
    period: 'last_24h',
  },
})
// const v1Domain = useV1Domain()
const onClick = async () => {
  // counter.value++
  // router.push({
  //   query: {
  //     test: counter.value,
  //   },
  // })
  // refresh()
  // await navigateTo({
  //   external: true,
  //   path: `${v1Domain}/login`,
  //   query: { promoCode: 'test' },
  // })
}
const expandedRows = ref({})
const query = ref({})
</script>

<template>
  <div>
    Page: test
    {{ route.query.test }}
    {{ data }}
    {{ error }}
    <button
      @click="onClick"
    >
      click
    </button>
    {{ expandedRows }}
    <ClientOnly>
      <BcTable
        v-model:expanded-rows="expandedRows"
        :query
        :data
        data-key="epoch"
      >
        <Column
          expander
        />
        <Column
          field="epoch"
        />
        <Column
          field="group_id"
        />
        <template #expansion="slotProps">
          <TestComponent
            :test="slotProps.data.epoch"
          />
        </template>
      </BcTable>
    </ClientOnly>
  </div>
</template>

<style scoped></style>
