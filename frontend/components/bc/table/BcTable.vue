<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type { ApiPagingResponse } from '~/types/api/common'

const props = defineProps<{
  addSpacer?: boolean,
  // cursor?: Cursor,
  data?: ApiPagingResponse<any>,
  dataKey: string, // Required Unique identifier for a data row
  expandable?: boolean,
  hasSelectionMode?: boolean,
  hidePager?: boolean,
  isRowExpandable?: (item: any) => boolean,
  // limit?: number,
  loading?: boolean,
  pageSize?: number,
  tableClass?: string,
}>()
const query = defineModel<Query>('query')

const expandedRows = ref<Record<any, boolean>>({})

const allExpanded = computed(() => {
  if (!props.expandable || !props.dataKey || !props.data?.data?.length) {
    return false
  }
  return !!props.data?.data?.every((item) => {
    if (props.isRowExpandable && !props.isRowExpandable(item)) {
      return true // ignore rows that can't be expanded
    }
    return !!expandedRows.value[item[props.dataKey!]]
  })
})

const toggleAll = (forceClose = false) => {
  if (!props.dataKey) {
    return
  }
  const wasExpanded = allExpanded.value
  props.data?.data?.forEach((item) => {
    if (wasExpanded || forceClose) {
      // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
      delete expandedRows.value[item[props.dataKey!]]
    }
    else if (!props.isRowExpandable || props.isRowExpandable(item)) {
      expandedRows.value[item[props.dataKey!]] = true
    }
  })
  expandedRows.value = { ...expandedRows.value }
}

const toggleItem = (item: any) => {
  if (!props.dataKey) {
    return
  }
  if (expandedRows.value[item[props.dataKey]]) {
    if (expandedRows.value) {
      // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
      delete expandedRows.value[item[props.dataKey]]
    }
  }
  else {
    expandedRows.value[item[props.dataKey]] = true
  }
  expandedRows.value = { ...expandedRows.value }
}

const setCursor = (value: string | undefined) => {
  if (!query.value) return
  toggleAll(true)
  query.value.cursor = value
}

const setLimit = (value: number) => {
  if (!query.value) return
  toggleAll(true)
  query.value.limit = value
}
const onSort = (event: DataTableSortEvent) => {
  if (!query.value) return
  toggleAll(true)
  const order = event.sortOrder === -1 ? 'asc' : 'desc'
  const { sortField } = event
  query.value.sort = `${sortField}:${order}` as const
}

watch(
  () => props.expandable,
  (expandable) => {
    if (!expandable) {
      toggleAll(true)
    }
  },
)
watch(
  () => props.data,
  () => {
    toggleAll(true)
  },
)
</script>

<template>
  <DataTable
    v-model:expanded-rows="expandedRows"
    class="bc-table"
    sort-mode="single"
    lazy
    :value="data?.data"
    :data-key
    :loading
    @sort="onSort"
  >
    <Column
      v-if="hasSelectionMode"
      selection-mode="multiple"
      class="selection"
    />
    <Column
      v-if="expandable"
      expander
      class="expander"
    >
      <template #header>
        <BcButtonIcon
          screenreader-text="dashboard.table.action.toggle_all_row_details"
          class="toggle"
          name="chevron-right"
          :rotation="allExpanded ? '90deg' : '0deg'"
          @click.stop.prevent="toggleAll()"
        />
      </template>

      <template #body="slotProps">
        <BcButtonIcon
          v-if="!isRowExpandable || isRowExpandable(slotProps.data)"
          screenreader-text="dashboard.table.action.toggle_row_detail"
          name="chevron-right"
          class="toggle"
          :rotation="
            dataKey && expandedRows[slotProps.data[dataKey]]
              ? '90deg'
              : '0deg'
          "
          @click.stop.prevent="toggleItem(slotProps.data)"
        />
      </template>
    </Column>
    <slot />
    <Column
      v-if="addSpacer"
      field="space_filler"
    >
      <template #body>
        <span />
        <!-- used to fill up the empty space so that the last column does not strech endlessly -->
      </template>
    </Column>
    <template #empty>
      <slot
        v-if="!loading"
        name="empty"
      >
        <DashboardTableEmpty />
      </slot>
    </template>

    <template #expansion="slotProps">
      <slot
        v-if="dataKey && expandedRows[slotProps.data[dataKey]]"
        name="expansion"
        v-bind="slotProps"
      />
    </template>

    <template #loading>
      <BcLoadingSpinner
        class="spinner"
        :loading="true"
        alignment="center"
      />
    </template>
    <template #footer>
      <BcTablePager
        v-if="!hidePager && data?.paging && query?.limit"
        :page-size="query.limit"
        :paging="data?.paging"
        @set-cursor="setCursor"
        @set-page-size="setLimit"
      >
        <template #bc-table-footer-left>
          <slot name="bc-table-footer-left" />
        </template>
        <template
          v-if="$slots['bc-table-footer-right']"
          #bc-table-footer-right
        >
          <slot name="bc-table-footer-right" />
        </template>
      </BcTablePager>
    </template>
  </DataTable>
</template>

<style lang="scss" scoped>
.bc-table {
  :deep(.expander) {
    width: 32px;
  }

  :deep(.p-datatable-emptymessage) {
    height: 140px;
    background: transparent;

    > td {
      border: none;
    }
  }
}

.bc-table-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-block: var(--padding-medium);
}

.toggle {
  cursor: pointer;
}
</style>
