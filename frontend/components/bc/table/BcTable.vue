<script setup lang="ts">
import { type TableResponse } from '~/types/dashboard/summary'
import type { Cursor } from '~/types/datatable'

interface Props {
  cursor: Cursor,
  dataKey: string, // Unique identifier for a data row
  pageSize: number,
  data?: TableResponse<any>,
  expandable?: boolean,
  title?: string,
  searchPlaceholder?:string,
  tableClass?: string
}
const props = defineProps<Props>()

const tableIsShown = ref(true)

// TODO: implement page size selection and search input
const emit = defineEmits<{(e: 'setCursor', value: Cursor): void, (e: 'setPageSize', value: number): void, (e: 'setSearch', value?: string): void }>()

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

const onInput = (event: Event) => {
  if (event.target) {
    emit('setSearch', (event.target as HTMLInputElement).value)
  }
}

</script>
<template>
  <slot name="header">
    <div class="bc-table-header">
      <div class="side">
        <BcIconToggle v-if="$slots.chart" v-model="tableIsShown" true-icon="fas fa-table" false-icon="far fa-chart-column" />
        <slot id="header-left" />
      </div>
      <div v-if="props.title" class="h1">
        {{ props.title }}
      </div>
      <div class="side right">
        <slot id="header-right" />
        <!--TODO: replace input with styled input-->
        <input v-if="props.searchPlaceholder && tableIsShown" type="text" :placeholder="props.searchPlaceholder" @input="onInput">
      </div>
    </div>
  </slot>
  <DataTable
    v-if="tableIsShown"
    v-model:expandedRows="expandedRows"
    :class="props.tableClass"
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
      <BcTablePager :page-size="pageSize" :paging="props.data?.paging" :cursor="cursor" @set-cursor="(cursor)=>emit('setCursor', cursor)" @set-page-size="(size) => emit('setPageSize', size)" />
    </template>
  </DataTable>
  <slot v-else name="chart" />
</template>

<style lang="scss" scoped>
.bc-table-header{
  height: 70px;
  padding: 0 var(--padding-large);
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  .side{
    width: 180px;
    &.right{
      display: flex;
      justify-content: flex-end;
    }
  }
}
:deep(.expander) {
  width: 32px;
}

.toggle {
  cursor: pointer;
}
</style>
