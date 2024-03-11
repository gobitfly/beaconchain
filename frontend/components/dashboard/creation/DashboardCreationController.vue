<script lang="ts" setup>
import { type DashboardType } from '~/types/dashboard'
import { type DashboardCreationDisplayType, type DashboardCreationState } from '~/types/dashboard/creation'
import { type VDBPostReturnData } from '~/types/api/validator_dashboard'

const router = useRouter()

interface Props {
  displayType: DashboardCreationDisplayType,
}
const props = defineProps<Props>()

const visible = ref<boolean>(false)

const state = ref<DashboardCreationState>('')
const type = ref<DashboardType | ''>('')
const name = ref<string>('')
const network = ref<string>('')

function show () {
  visible.value = true

  state.value = 'type'
  type.value = ''
  name.value = ''
  network.value = ''
}

defineExpose({
  show
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

  visible.value = false

  router.push(`/dashboard/${newDashboardId}`)
}
</script>

<template>
  <div v-if="visible">
    <BcDialog v-if="props.displayType === 'modal'" v-model="visible">
      <DashboardCreationTypeMask v-if="state === 'type'" v-model:state="state" v-model:type="type" v-model:name="name" @create-pressed="onCreate()" />
      <DashboardCreationNetworkMask v-else-if="state === 'network'" v-model:state="state" v-model:network="network" @create-pressed="onCreate()" />
    </BcDialog>
    <div v-else-if="props.displayType === 'panel'">
      <div class="panel_container">
        <DashboardCreationTypeMask v-if="state === 'type'" v-model:state="state" v-model:type="type" v-model:name="name" @create-pressed="onCreate()" />
        <DashboardCreationNetworkMask v-else-if="state === 'network'" v-model:state="state" v-model:network="network" @create-pressed="onCreate()" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .panel_container {
    border: 1px solid var(--primary-orange);
    border-radius: var(--border-radius);
    max-width: 460px;
    padding: var(--padding-large);

    @media (max-width: 400px) {
      padding: var(--padding);
    }
  }
</style>
