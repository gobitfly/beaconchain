<script lang="ts" setup>
import { type DashboardType, type DashboardCreationDisplayType, type DashboardCreationState } from '~/types/dashboard/creation'
import { type VDBPostReturnData } from '~/types/api/validator_dashboard'

const router = useRouter()

const displayType = defineModel<DashboardCreationDisplayType>({ required: true })
const modalVisibility = ref(false)

const state = ref<DashboardCreationState>('')
const type = ref<DashboardType>('')
const name = ref<string>('')
const network = ref<string>('')

watch(() => displayType.value, () => {
  if (displayType.value === 'panel') {
    modalVisibility.value = false
    state.value = 'type'
  } else if (displayType.value === 'modal') {
    modalVisibility.value = true
    state.value = 'type'
  } else {
    state.value = ''
    type.value = ''
    name.value = ''
    network.value = ''
  }
}, { immediate: true })

watch(() => modalVisibility.value, () => {
  if (modalVisibility.value === false) {
    displayType.value = ''
  }
})

async function onCreate () {
  let newDashboardId = -1
  if (type.value === 'account') {
    await useCustomFetch<undefined>(API_PATH.DASHBOARD_CREATE_ACCOUNT, { // TODO: Use correct type once available
      body: {
        name: name.value
      }
    })
    newDashboardId = 1
  } else if (type.value === 'validator') {
    const response = await useCustomFetch<VDBPostReturnData>(API_PATH.DASHBOARD_CREATE_VALIDATOR, {
      body: {
        name: name.value,
        network: network.value
      }
    })
    newDashboardId = response.id || 1
  }

  displayType.value = ''

  // TODO
  console.log('New dashboard ID:', newDashboardId)
  router.push(`/dashboard/${newDashboardId}`)
}
</script>

<template>
  <BcDialog v-if="displayType === 'modal'" v-model="modalVisibility">
    <DashboardCreationTypeMask v-if="state === 'type'" v-model:state="state" v-model:type="type" v-model:name="name" @create-pressed="onCreate()" />
    <DashboardCreationNetworkMask v-else-if="state === 'network'" v-model:state="state" v-model:network="network" @create-pressed="onCreate()" />
  </BcDialog>
  <div v-else-if="displayType === 'panel'">
    <div class="panel_container">
      <DashboardCreationTypeMask v-if="state === 'type'" v-model:state="state" v-model:type="type" v-model:name="name" @create-pressed="onCreate()" />
      <DashboardCreationNetworkMask v-else-if="state === 'network'" v-model:state="state" v-model:network="network" @create-pressed="onCreate()" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .panel_container {
    border: 1px solid var(--primary-orange);
    border-radius: var(--border-radius);
    max-width: 460px;
    padding: var(--padding-large);
  }
</style>
