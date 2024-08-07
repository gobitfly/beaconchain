<script setup lang="ts">
import {
  DashboardGroupSelectionDialog,
  DashboardValidatorEpochDutiesModal,
} from '#components'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'

const { groups } = useValidatorDashboardGroups()

const selectedGroupId = ref<number>(DAHSHBOARDS_ALL_GROUPS_ID)

const dialog = useDialog()

function onClose(groupId: boolean) {
  setTimeout(() => {
    alert('new group: ' + groupId)
  }, 100)
}

const openGroupSelection = (withPreselection: boolean) => {
  dialog.open(DashboardGroupSelectionDialog, {
    data: {
      groupId: withPreselection ? groups.value?.[0]?.id : undefined,
      selectedValidators: withPreselection ? 1 : 10,
      totalValidators: 123,
    },
    onClose: response => onClose(response?.data),
  })
}

const openEpochDuties = () => {
  dialog.open(DashboardValidatorEpochDutiesModal, {
    data: {
      dashboardKey: 5003,
      epoch: 1370,
      groupId: 4,
      groupName: 'My test group',
    },
  })
}
</script>

<template>
  <Button
    label="Open Epoch Duties"
    @click="openEpochDuties"
  />
  <div class="icon-holder">
    <div class="premium-row">
      Come on, you cheap friend, buy that premium<BcPremiumGem
        style="margin-left: 10px"
      />
    </div>
    <DashboardGroupSelection
      v-model="selectedGroupId"
      class="group-selection"
    />
    <DashboardGroupSelection
      v-model="selectedGroupId"
      class="group-selection"
      :include-all="true"
    />
  </div>
  <div class="status-holder">
    <div class="status">
      <ValidatorTableStatus status="slashed" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="exited" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="deposited" />
    </div>
    <div class="status">
      <ValidatorTableStatus
        status="pending"
        :position="12345"
      />
    </div>
    <div class="status">
      <ValidatorTableStatus status="slashing_offline" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="slashing_online" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="exiting_offline" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="exiting_online" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="active_offline" />
    </div>
    <div class="status">
      <ValidatorTableStatus status="active_online" />
    </div>
  </div>
  <div class="icon_holder">
    <Button
      class="group_selection"
      label="Open Group Selection preselected"
      @click="openGroupSelection(true)"
    />
    <Button
      class="group_selection"
      label="Open Group Selection"
      @click="openGroupSelection(false)"
    />
  </div>
</template>

<style lang="scss" scoped>
.premium-row {
  display: inline-flex;
  gap: 10px;
}
.group-selection {
  width: 200px;
}
.icon-holder {
  margin: 10px;
  display: flex;
  flex-direction: column;
  gap: var(--padding);
}

.status-holder {
  display: flex;
  flex-wrap: wrap;
  padding: 10px;
  .status {
    width: 140px;
    padding: 5px;
    border: 1px solid black;
  }
}
</style>
