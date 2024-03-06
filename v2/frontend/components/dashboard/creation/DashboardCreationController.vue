<script lang="ts" setup>
import { type DashboardType, type DashboardCreationDisplayType, type DashboardCreationState } from '~/types/dashboard/creation'

interface Props {
  displayType: DashboardCreationDisplayType;
}
const props = defineProps<Props>()

watch(() => props.displayType, () => {
  if (props.displayType === 'panel') {
    modalVisibility.value = false
    state.value = 'type'
  } else if (props.displayType === 'modal') {
    modalVisibility.value = true
    state.value = 'type'
  } else {
    state.value = ''
    type.value = ''
    name.value = ''
    network.value = ''
  }
})

function onCreate () {
  if (type.value === 'account') {
    console.log(`Creating ${type.value} dashboard ${name.value} via ${API_PATH.DASHBOARD_CREATE_ACCOUNT}`)
  } else if (type.value === 'validator') {
    console.log(`Creating ${type.value} dashboard ${name.value} on ${network.value} via ${API_PATH.DASHBOARD_CREATE_VALIDATOR}`)
  }
}

const modalVisibility = ref(false)

const state = ref<DashboardCreationState>('')

const type = ref<DashboardType>('')
const name = ref<string>('')
const network = ref<string>('')
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
