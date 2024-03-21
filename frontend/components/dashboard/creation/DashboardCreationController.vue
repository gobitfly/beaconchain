<script lang="ts" setup>
import { type DashboardType, type ValidatorDashboardNetwork } from '~/types/dashboard'
import { type DashboardCreationDisplayType, type DashboardCreationState } from '~/types/dashboard/creation'

const router = useRouter()

const store = useUserDashboardStore()
const { createValidatorDashboard, createAccountDashboard } = store

interface Props {
  displayType: DashboardCreationDisplayType,
}
const props = defineProps<Props>()

const visible = ref<boolean>(false)

const state = ref<DashboardCreationState>('')
const type = ref<DashboardType | ''>('')
const name = ref<string>('')
// TODO: replace network types once we have them
const network = ref<ValidatorDashboardNetwork>()

function show () {
  visible.value = true

  state.value = 'type'
  type.value = ''
  name.value = ''
  network.value = undefined
}

defineExpose({
  show
})

function onNext () {
  if (state.value === 'type') {
    if (type.value === 'account') {
      createDashboard()
    } else {
      state.value = 'network'
    }
  } else if (state.value === 'network') {
    createDashboard()
  }
}

function onBack () {
  if (state.value === 'network') {
    state.value = 'type'
  }
}

async function createDashboard () {
  let newDashboardId = -1
  if (type.value === 'account') {
    if (!name.value) {
      return
    }
    const response = await createAccountDashboard(name.value)
    newDashboardId = response?.id || 1
  } else if (type.value === 'validator') {
    if (!name.value || !network.value) {
      return
    }
    const response = await createValidatorDashboard(name.value, network.value)
    newDashboardId = response?.id || 1
  }

  visible.value = false

  router.push(`/dashboard/${newDashboardId}`)
}
</script>

<template>
  <BcDialog v-if="visible && props.displayType === 'modal'" v-model="visible">
    <DashboardCreationTypeMask
      v-if="state === 'type'"
      v-model:state="state"
      v-model:type="type"
      v-model:name="name"
      @next="onNext()"
    />
    <DashboardCreationNetworkMask
      v-else-if="state === 'network'"
      v-model:state="state"
      v-model:network="network"
      @next="onNext()"
      @back="onBack()"
    />
  </BcDialog>
  <div v-else-if="visible && props.displayType === 'panel'">
    <div class="panel-container">
      <DashboardCreationTypeMask
        v-if="state === 'type'"
        v-model:state="state"
        v-model:type="type"
        v-model:name="name"
        @next="onNext()"
      />
      <DashboardCreationNetworkMask
        v-else-if="state === 'network'"
        v-model:state="state"
        v-model:network="network"
        @next="onNext()"
        @back="onBack()"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.panel-container {
  border: 1px solid var(--primary-orange);
  border-radius: var(--border-radius);
  padding: var(--padding-large);
  box-sizing: border-box;
  width: 460px;
  max-width: calc(100% - 42px);

  @media (max-width: 400px) {
    padding: var(--padding);
    max-width: calc(100% - 22px);
  }
}
</style>
