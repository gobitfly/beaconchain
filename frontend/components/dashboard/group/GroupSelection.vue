<script setup lang="ts">
const { overview } = storeToRefs(useValidatorDashboardOverviewStore())

const list = computed(() => {
  return [{ id: -1, name: '' }].concat(overview.value?.groups ?? [])
})

const selected = defineModel<number>({ required: true })

const selectedGroup = computed(() => {
  return list.value.find(item => item.id === selected.value)
})

</script>
<template>
  <BcDropdown v-model="selected" :options="list" option-value="id" option-label="name">
    <template v-if="selectedGroup" #value>
      <span>{{ $t('dashboard.group.selection.group') }}: <DashboardGroupLabel :group="selectedGroup" /></span>
    </template>
    <template #option="slotProps">
      <DashboardGroupLabel :group="slotProps" />
    </template>
  </BcDropdown>
</template>
<style lang="scss" scoped>
</style>
