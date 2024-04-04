<script setup lang="ts">
import type { DashboardCreationController } from '#components'
import { type DashboardCreationDisplayType } from '~/types/dashboard/creation'

const { dashboardKey } = useDashboardKeyProvider('account')

// TODO: This duplicates code from the validator dashboard page
// Once the account dashboard page is tackled, improve this
const dashboardCreationControllerPanel = ref<typeof DashboardCreationController>()
const dashboardCreationControllerModal = ref<typeof DashboardCreationController>()
function showDashboardCreation (type: DashboardCreationDisplayType) {
  if (type === 'panel') {
    dashboardCreationControllerPanel.value?.show()
  } else {
    dashboardCreationControllerModal.value?.show()
  }
}

</script>

<template>
  <div v-if="dashboardKey==''">
    <BcPageWrapper>
      <DashboardCreationController ref="dashboardCreationControllerPanel" class="panel-controller" :display-type="'panel'" />
    </BcPageWrapper>
  </div>
  <div v-else>
    <DashboardCreationController ref="dashboardCreationControllerModal" class="modal-controller" :display-type="'modal'" />
    <BcPageWrapper>
      <template #top>
        <DashboardHeader @show-creation="showDashboardCreation('modal')" />
      </template>
      <h1>Account Dashboard {{ dashboardKey }}</h1>
    </BcPageWrapper>
  </div>
</template>

<style lang="scss" scoped>

.panel-controller {
  display: flex;
  justify-content: center;
  padding: 60px 0px;
}

:global(.modal_controller) {
  max-width: 460px;
  width: 100%;
}

</style>
