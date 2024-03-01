<script lang="ts" setup>

import { type DisplayType } from '~/types/dashboard/creation'

const displayType = ref<DisplayType>('modal')

const changeDisplayType = (newDisplayType: DisplayType) => {
  displayType.value = newDisplayType
}
</script>

<template>
  <div class="settings_container">
    <h1>
      Display Type Settings
    </h1>
    Current Display Type: {{ displayType }}
    <div class="button_container">
      <Button @click="changeDisplayType('modal')">
        Display Type: Modal
      </Button>
      <Button @click="changeDisplayType('panel')">
        Display Type: Panel
      </Button>
    </div>
  </div>
  <DashboardCreationController :display-type="displayType" />
</template>

<style lang="scss">
  .settings_container {
    padding: 10px;
    border-bottom: 2px solid var(--primary-color);

    .button_container {
      display: flex;
      padding: 10px;
      gap: 10px;
    }
  }

</style>
