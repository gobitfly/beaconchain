<script setup lang="ts">
import { type TableResponse } from '~/types/dashboard/summary'
import type { Cursor } from '~/types/datatable'

interface Props {
  cursor: Cursor,
  dataKey: string, // Unique identifier for a data row
  pageSize: number,
  data?: TableResponse<any>,
  expandable?: boolean,
  tableClass?: string
}
const props = defineProps<Props>()

const emit = defineEmits<{(e: 'setCursor', value: Cursor): void, (e: 'setPageSize', value: number): void }>()

const expandedRows = ref<any[]>([])

const allExpanded = computed(() => {
  return !!props.data?.data.every((item) => {
    return !!expandedRows.value[item[props.dataKey]]
  })
})

const toggleAll = () => {
  const wasExpanded = allExpanded.value
  const rows = { ...expandedRows.value }
  props.data?.data.forEach((item) => {
    if (wasExpanded) {
      delete rows[item[props.dataKey]]
    } else {
      rows[item[props.dataKey]] = item
    }
  })
  expandedRows.value = rows
}

</script>
<template>
  <DataTable
    v-model:expandedRows="expandedRows"
    sort-mode="multiple"
    lazy
    :total-records="props.data?.paging.total_count"
    :rows="pageSize"
    :value="props.data?.data"
    :data-key="dataKey"
  >
    <Column v-if="props.expandable" expander class="expander">
      <template #header>
        <IconChevron class="toggle" :direction="allExpanded ? 'bottom' : 'right'" @click="toggleAll" />
      </template>
      <template #rowtogglericon="slotProps">
        <IconChevron class="toggle" :direction="slotProps.rowExpanded ? 'bottom' : 'right'" />
      </template>
    </Column>
    <slot />
    <template #expansion="slotProps">
      <slot name="expansion" v-bind="slotProps" />
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
