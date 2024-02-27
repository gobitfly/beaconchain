<script setup lang="ts">
import type { ApiPagingResponse } from '~/types/api/common'
import type { Cursor } from '~/types/datatable'

interface Props {
  cursor: Cursor,
  dataKey: string, // Unique identifier for a data row
  pageSize: number,
  data?: ApiPagingResponse<any>,
  expandable?: boolean,
  tableClass?: string
}
const props = defineProps<Props>()

const emit = defineEmits<{(e: 'setCursor', value: Cursor): void, (e: 'setPageSize', value: number): void }>()

const expandedRows = ref<Record<any, boolean>>({ })

const allExpanded = computed(() => {
  if (!props.expandable) {
    return false
  }
  return !!props.data?.data.every((item) => {
    return !!expandedRows.value[item[props.dataKey]]
  })
})

const toggleAll = () => {
  const wasExpanded = allExpanded.value
  props.data?.data.forEach((item) => {
    if (wasExpanded) {
      delete expandedRows.value[item[props.dataKey]]
    } else {
      expandedRows.value[item[props.dataKey]] = true
    }
  })
  expandedRows.value = { ...expandedRows.value }
}

const toggleItem = (item: any) => {
  if (expandedRows.value[item[props.dataKey]]) {
    if (expandedRows.value) {
      delete expandedRows.value[item[props.dataKey]]
    }
  } else {
    expandedRows.value[item[props.dataKey]] = true
  }
  expandedRows.value = { ...expandedRows.value }
}

</script>
<template>
  <DataTable
    v-model:expandedRows="expandedRows"
    sort-mode="multiple"
    lazy
    :value="props.data?.data"
    :data-key="dataKey"
  >
    <Column v-if="props.expandable" expander class="expander">
      <template #header>
        <IconChevron class="toggle" :direction="allExpanded ? 'bottom' : 'right'" @click.stop.prevent="toggleAll" />
      </template>
      <template #body="slotProps">
        <IconChevron class="toggle mine" :direction="expandedRows[slotProps.data[props.dataKey]] ? 'bottom' : 'right'" @click.stop.prevent="toggleItem(slotProps.data)" />
      </template>
    </Column>
    <slot />
    <template #expansion="slotProps">
      <slot v-if="expandedRows[slotProps.data[props.dataKey]]" name="expansion" v-bind="slotProps" />
    </template>
    <template #footer>
      <BcTablePager
        :page-size="pageSize"
        :paging="props.data?.paging"
        :cursor="cursor"
        @set-cursor="(cursor) => emit('setCursor', cursor)"
        @set-page-size="(size) => emit('setPageSize', size)"
      />
    </template>
  </DataTable>
</template>

<style lang="scss" scoped>
:deep(.expander) {
  width: 32px;
}

.toggle {
  cursor: pointer;
}
</style>
