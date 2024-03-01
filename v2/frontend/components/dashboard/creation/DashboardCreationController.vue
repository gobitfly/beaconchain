<script lang="ts" setup>
import { type DisplayType, type State } from '~/types/dashboard/creation'

interface Props {
  displayType: DisplayType;
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

const state = ref<State>('none')
const changeState = (newState: State) => {
  state.value = newState
  if (newState === 'none') {
    modalVisibility.value = false
  }
}
</script>

<template>
  <div class="settings_container">
    <h1>
      Dashboard Creation Controller
    </h1>
    Current State: {{ state }}
    <div class="button_container">
      <Button @click="changeState('none')">
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
    <DashboardCreationNetworkMask v-if="state === 'network'" />
    <DashboardCreationTypeMask v-else-if="state === 'type'" />
  </BcDialog>
  <div v-else-if="displayType === 'panel'">
    <DashboardCreationNetworkMask v-if="state === 'network'" />
    <DashboardCreationTypeMask v-else-if="state === 'type'" />
  </div>
</template>

<style lang="scss">
  .settings_container {
    padding: 10px;

    .button_container {
      display: flex;
      padding: 10px;
      gap: 10px;
    }
  }
</style>
