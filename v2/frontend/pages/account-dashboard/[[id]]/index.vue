<script setup lang="ts">
import type { DashboardCreationController } from '#components'

const { dashboardKey } = useDashboardKeyProvider('account')

// TODO: This duplicates code from the validator dashboard page
// Once the account dashboard page is tackled, improve this
const dashboardCreationControllerModal = ref<typeof DashboardCreationController>()
function showDashboardCreation () {
  dashboardCreationControllerModal.value?.show()
}

</script>

<template>
  <div v-if="dashboardKey==''">
    <BcPageWrapper>
      <DashboardCreationController
        ref="dashboardCreationControllerPanel"
        class="panel-controller"
        :display-type="'panel'"
        :initially-visislbe="true"
      />
    </BcPageWrapper>
  </div>
  <div v-else>
    <DashboardCreationController ref="dashboardCreationControllerModal" class="modal-controller" :display-type="'modal'" />
    <BcPageWrapper>
      <template #top>
        <DashboardHeader @show-creation="showDashboardCreation()" />
      </template>
      <DashboardControls type="account" />
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
