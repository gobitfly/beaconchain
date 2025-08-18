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
key.value = undefined
const {
  data,
  error,
  // refresh,
} = useFetch(() => `/api/bff/validator-dashboards/${key.value}/validators`, {
  // key: 'privateDashboards',
  // } = useApi(() => '/api/bff/validator-dashboards/MQ/validators', {
  // getCachedData: key => nuxtApp.payload.state[key] ?? nuxtApp.payload.data[key],
  // body: {
  //   dashboardKey: key.value,
  //   groupId: 0,
  // },
  // immediate: true,
  // key: 'dashboardSummaryDetails',
  query: {
    // period: 'last_24h',
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
  await navigateTo({
    name: 'dashboard-id',
    params: {
      id: 22,
    },
  })
  // await navigateTo({
  //   external: true,
  //   path: `${v1Domain}/login`,
  //   query: { promoCode: 'test' },
  // })
}
const expandedRows = ref({})
const query = ref({})
// const data = ref({
//   data: [
//     {
//       epoch: 1, group_id: 'group1', index: 1, public_key: '0x123',
//     },
//     {
//       epoch: 2, group_id: 'group1', index: 2, public_key: '0x456',
//     },
//     {
//       epoch: 3, group_id: 'group2', index: 3, public_key: '0x789',
//     },
//   ],
// })
// const {
//   dashboards,
//   validatorDashboards,
// } = usePrivateDashboards()
// validatorDashboards[0].name = 'test'

// const test = ref({
//   one: 'value1',
//   two: {
//     four: 'value4',
//     three: 'value3',
//   },
// })

// const test2 = toRefs(test.value)
// test2.two.value = {
//   four: 'value5',
//   three: 'value6',
// }
const { hasShareCustomDashboard } = usePremiumPerks()
// hasShareCustomDashboard.value = false
</script>

<template>
  <div>
    <pre>
      {{ hasShareCustomDashboard }}
      <!-- {{ dashboards }} -->
      <!-- {{ validatorDashboards }} -->
      <!-- {{ route.query.test }} -->
      {{ data }}
      {{ error }}
    </pre>
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
          header="Epoch"
          sortable
        />
        <Column
          header="Public Key"
          field="public_key"
        />
        <Column
          header="Group ID"
          field="group_id"
          sortable
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
