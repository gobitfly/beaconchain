<script setup lang="ts" generic="ChipItemValue extends string">
type ChipItem = {
  label: string,
  value: ChipItemValue,
}
const {
  ariaLabel,
  items,
} = defineProps<{
  ariaLabel: string,
  items: ChipItem[],
}>()

const { t: $t } = useTranslation()

const selectedItems = defineModel<ChipItem['value'][]>({ default: () => [] })

const itemValues = computed(() => items.map(item => item.value))
const isAllSelected = computed(() => {
  if (!selectedItems.value?.length) return false

  return items.every(filter => selectedItems.value?.includes(filter.value))
})

const getIsFilterSelected = (value: ChipItem['value']) => {
  return selectedItems.value?.includes(value)
}
const selectAllFilters = () => {
  selectedItems.value = itemValues.value
}
const selectOnlyFilter = (value: ChipItem['value']) => {
  selectedItems.value = [ value ]
}
const deselectFilter = (value: ChipItem['value']) => {
  return selectedItems.value = selectedItems.value?.filter(type => type !== value)
}
const addFilter = (value: ChipItem['value']) => {
  selectedItems.value = [
    ...selectedItems.value,
    value,
  ]
}

const handleSelectFilter = (value: ChipItem['value']) => {
  if (isAllSelected.value) {
    return selectOnlyFilter(value)
  }
  if (getIsFilterSelected(value)) {
    const isOnlySelectedFilter = selectedItems.value?.length === 1

    if (isOnlySelectedFilter) {
      return selectAllFilters()
    }

    return deselectFilter(value)
  }

  return addFilter(value)
}
</script>

<template>
  <ul
    class="flex flex-wrap gap-md px-2xl py-lg"
    role="group"
    :aria-label
  >
    <li>
      <BaseChip
        class="px-xl"
        :is-selected="isAllSelected"
        @click.prevent="selectAllFilters"
      >
        {{ $t('base.common.all') }}
      </BaseChip>
    </li>
    <li
      v-for="item in items"
      :key="item.value"
    >
      <BaseChip
        :is-selected="!isAllSelected && getIsFilterSelected(item.value)"
        @click.prevent="handleSelectFilter(item.value)"
      >
        {{ item.label }}
      </BaseChip>
    </li>
  </ul>
</template>
