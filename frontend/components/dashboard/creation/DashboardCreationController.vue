<script lang="ts" setup>
import { type DashboardType, type DashboardCreationDisplayType, type DashboardCreationState } from '~/types/dashboard/creation'

interface Props {
  displayType: DashboardCreationDisplayType;
}
const props = defineProps<Props>()

watch(() => props.displayType, () => {
  if (props.displayType === 'panel') {
    modalVisibility.value = false
  } else {
    modalVisibility.value = true
  }
})

const modalVisibility = ref(false)

const state = ref<DashboardCreationState>('')
const changeState = (newState: DashboardCreationState) => {
  state.value = newState
  if (newState === '') {
    modalVisibility.value = false
  }
}

const type = ref<DashboardType>('')
const name = ref<string>('')
const network = ref<string>('')
</script>

<template>
  <div class="settings_container">
    <h1>
      Dashboard Creation Controller
    </h1>
    <div>
      Type: {{ type }}
    </div>
    <div>
      Name: {{ name }}
    </div>
    <div>
      Network: {{ network }}
    </div>
    Current State: {{ state }}
    <div class="button_container">
      <Button @click="changeState('')">
        State: None
      </Button>
      <Button @click="changeState('type')">
        State: Type
      </Button>
      <Button @click="changeState('network')">
        State: Network
      </Button>
    </div>
  </div>
  <BcDialog v-if="displayType === 'modal'" v-model="modalVisibility">
    <DashboardCreationTypeMask v-if="state === 'type'" v-model:state="state" v-model:type="type" v-model:name="name" />
    <DashboardCreationNetworkMask v-else-if="state === 'network'" v-model:state="state" v-model:network="network" />
  </BcDialog>
  <div v-else-if="displayType === 'panel'">
    <DashboardCreationTypeMask v-if="state === 'type'" v-model:state="state" v-model:type="type" v-model:name="name" />
    <DashboardCreationNetworkMask v-else-if="state === 'network'" v-model:state="state" v-model:network="network" />
  </div>
</template>

<style lang="scss" scoped>
  .settings_container {
    padding: 10px;

    .button_container {
      display: flex;
      padding: 10px;
      gap: 10px;
    }
  }
</style>
