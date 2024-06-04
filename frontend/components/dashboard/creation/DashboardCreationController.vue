<script lang="ts" setup>
import { type DashboardType } from '~/types/dashboard'
import { type DashboardCreationDisplayType, type DashboardCreationState } from '~/types/dashboard/creation'
import { ChainIDs } from '~/types/networks'
import { API_PATH } from '~/types/customFetch'

const { createValidatorDashboard, createAccountDashboard } = useUserDashboardStore()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const { dashboards } = useUserDashboardStore()
const { user } = useUserStore()

interface Props {
  displayType: DashboardCreationDisplayType,
  initiallyVisislbe?: boolean
}
const props = defineProps<Props>()

const visible = ref<boolean>(false)

const state = ref<DashboardCreationState>('')
const type = ref<DashboardType | ''>('')
const name = ref<string>('')
const network = ref<ChainIDs>()
const { dashboardKey, publicEntities } = useDashboardKey()
const { isLoggedIn } = useUserStore()
const { fetch } = useCustomFetch()
const route = useRoute()

const maxDashboards = computed(() => {
  // TODO: currently there is no value for "amount of account dashboards", using "amount of validator dashboards" instead for now
  return user.value?.premium_perks.validator_dashboards ?? 1
})
const accountsDisabled = computed(() => {
  return !showInDevelopment || (dashboards.value?.account_dashboards?.length ?? 0) >= maxDashboards.value
})
const validatorsDisabled = computed(() => {
  return (dashboards.value?.validator_dashboards?.length ?? 0) >= maxDashboards.value
})

function show () {
  visible.value = true

  state.value = 'type'
  type.value = ''
  if (!validatorsDisabled.value) {
    type.value = 'validator'
  } else if (!accountsDisabled.value) {
    type.value = 'account'
  }
  name.value = isLoggedIn.value ? '' : 'cookie'
  network.value = undefined
}

defineExpose({
  show
})
if (props.initiallyVisislbe) {
  show()
}

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
  visible.value = false
  const matchingType = route.name === 'dashboard-id' && type.value === 'validator'

  const publicKey = matchingType && !isLoggedIn.value ? dashboardKey.value : undefined
  if (type.value === 'account') {
    if (!name.value) {
      return
    }
    const response = await createAccountDashboard(name.value, publicKey)

    await navigateTo(`/account-dashboard/${response?.hash ?? response?.id ?? 1}`)
  } else if (type.value === 'validator') {
    if (!name.value || !network.value) {
      return
    }

    const response = await createValidatorDashboard(name.value, network.value, publicKey)
    if (matchingType && publicEntities.value?.length && response?.id && response.id > 0) {
      await fetch(API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT, { method: 'POST', body: { validators: publicEntities.value, group_id: '0' } }, { dashboardKey: response.id })
    }
    await navigateTo(`/dashboard/${response?.hash ?? response?.id ?? 1}`)
  }
}
</script>

<template>
  <BcDialog v-if="visible && props.displayType === 'modal'" v-model="visible">
    <DashboardCreationTypeMask
      v-if="state === 'type'"
      v-model:state="state"
      v-model:type="type"
      v-model:name="name"
      :accounts-disabled="accountsDisabled"
      :validators-disabled="validatorsDisabled"
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
        :accounts-disabled="accountsDisabled"
        :validators-disabled="validatorsDisabled"
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
