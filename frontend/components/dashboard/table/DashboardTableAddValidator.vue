<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faCirclePlus } from '@fortawesome/pro-regular-svg-icons'
const manageValidatorsModalVisisble = ref(false)

const addValidator = () => {
  manageValidatorsModalVisisble.value = true
}
</script>
<template>
  <div class="empty" @click="addValidator">
    <span>{{ $t('dashboard.validator.summary.add_validator') }}</span>
    <FontAwesomeIcon :icon="faCirclePlus" />
    <DashboardValidatorManagementModal v-model="manageValidatorsModalVisisble" />
  </div>
</template>

<style lang="scss" scoped>
.empty {
  width: 100%;
  height: 400px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: var(--text-color-disabled);
  gap: var(--padding);
  cursor: pointer;
}
</style>
