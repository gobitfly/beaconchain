<script lang="ts" setup>
import type { DashboardType } from '~/types/dashboard'
import type {
  DashboardCreationDisplayMode,
  DashboardCreationState,
} from '~/types/dashboard/creation'
import type { ChainId } from '~/types/network'

const userDashboardStore = useUserDashboardStore()
const {
  createValidatorDashboard,
} = userDashboardStore

const {
  dashboards,
} = storeToRefs(userDashboardStore)

const {
  isLoggedIn,
  user,
} = useUserStore()
const { currentNetwork } = useNetwork()

const visible = ref<boolean>(false)
const type = ref<'' | DashboardType>('')
const name = ref<string>('')
const network = ref<ChainId>(0)
const forcedDashboardType = ref<'' | DashboardType>('')
const {
  dashboardKey,
  publicEntities,
} = useDashboardKey()
const { fetch } = useCustomFetch()

const maxDashboards = computed(() => {
  // TODO: currently there is no value for "amount of account dashboards", using
  //  "amount of validator dashboards" instead for now
  return user.value?.premium_perks.validator_dashboards ?? 1
})

const validatorsDisabled = computed(() => {
  return (
    (dashboards.value?.validator_dashboards?.length ?? 0)
    >= maxDashboards.value
    || (!!forcedDashboardType.value && forcedDashboardType.value !== 'validator')
  )
})

async function createDashboard() {
  visible.value = false

  const publicKey
    = !isLoggedIn.value ? dashboardKey.value : undefined

  if (!name.value || !network.value) {
    return
  }

  const response = await createValidatorDashboard(
    name.value,
    network.value,
    publicKey,
  )
  if (
    publicEntities.value?.length
    && response?.id
    && response.id > 0
  ) {
    await fetch(
      'DASHBOARD_VALIDATOR_MANAGEMENT',
      {
        body: {
          group_id: '0',
          validators: publicEntities.value,
        },
        method: 'POST',
      },
      { dashboardKey: response.id },
    )
  }
  await navigateTo(`/dashboard/${response?.key ?? response?.id ?? 1}`)
}

// defineExpose({ show })
// if (props.initiallyVisible) {
//   show()
// }

function show(
  forcedType: '' | DashboardType = '',
) {
  visible.value = true
  type.value = forcedDashboardType.value = forcedType
  if (!type.value) {
    if (!validatorsDisabled.value) {
      type.value = 'validator'
    }
  }
  network.value = currentNetwork.value ?? 1
  name.value = isLoggedIn.value ? '' : 'cookie'
}
</script>

<template>
  <DashboardCreationTypeMask
    v-model:type="type"
    v-model:name="name"
    :validators-disabled
    accounts-disabled
    @next="createDashboard"
  />
  <!-- <div v-else-if="visible && props.displayMode === 'panel'">
    <div class="panel-container">
      <DashboardCreationTypeMask
        v-if="state === 'type'"
        v-model:state="state"
        v-model:type="type"
        v-model:name="name"
        :validators-disabled
        @next=" createDashboard"
      />
    </div>
  </div> -->
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
