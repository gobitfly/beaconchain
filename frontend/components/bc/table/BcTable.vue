<script setup lang="ts" generic="T extends ApiPagingResponse<Record<string, any>> | null">
import type { DataTableSortEvent } from 'primevue/datatable'
import type { ApiPagingResponse } from '~/types/api/common'
import type { Cursor } from '~/types/datatable'

const props = defineProps<{
  addSpacer?: boolean,
  cursor?: Cursor,
  data: T,
  dataKey: string, // Required Unique identifier for a data row
  expandable?: boolean,
  hidePager?: boolean,
  isLoading?: boolean,
  isRowExpandable?: (item: any) => boolean,
  // pageSize?: number,
  // selectedSort?: string,
  selectionMode?: 'multiple' | 'single',
  tableClass?: string,
}>()

// const emit = defineEmits<{
//   // (e: 'changeCursor', value: Cursor): void,
//   (e: 'setPageSize', value: number): void,
// }>()

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
// Todo toggle all
  console.log('Todo', forceClose)
  return
  // if (!props.dataKey) {
  //   return
  // }
  // const wasExpanded = allExpanded.value
  // props.data?.data?.forEach((item) => {
  //   if (wasExpanded || forceClose) {
  //     // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
  //     delete expandedRows.value[item[props.dataKey!]]
  //   }
  //   else if (!props.isRowExpandable || props.isRowExpandable(item)) {
  //     expandedRows.value[item[props.dataKey!]] = true
  //   }
  // })
  // expandedRows.value = { ...expandedRows.value }
}

// const toggleItem = (item: any) => {
//   if (!props.dataKey) {
//     return
//   }
//   if (expandedRows.value[item[props.dataKey]]) {
//     if (expandedRows.value) {
//       // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
//       delete expandedRows.value[item[props.dataKey]]
//     }
//   }
//   else {
//     expandedRows.value[item[props.dataKey]] = true
//   }
//   expandedRows.value = { ...expandedRows.value }
// }

const changeCursor = (value: string | undefined) => {
  // toggleAll(true)
  query.value.cursor = value
}

// const setPageSize = (value: number) => {
//   // toggleAll(true)
//   emit('setPageSize', value)
// }

// watch(
//   () => props.expandable,
//   (expandable) => {
//     if (!expandable) {
//       // toggleAll(true)
//     }
//   },
// )
// watch(
//   () => props.data,
//   () => {
//     // toggleAll(true)
//   },
// )
/**
 * Use useDefaultQuery in parent component to set the right defaults
 */
const query = defineModel<Query>('query', {
  required: true,
})
const onSort = (event: DataTableSortEvent) => {
  const {
    sortField,
    sortOrder,
  } = event
  if (query.value) {
    query.value.sort = `${sortField}:${sortOrder === -1 ? 'asc' : 'desc'}`
  }
}
</script>

<template>
  <DataTable
    v-model:expanded-rows="expandedRows"
    class="bc-table"
    sort-mode="single"
    lazy
    :value="data?.data"
    :data-key
    :loading="isLoading"
    :table-class
    @sort="onSort"
  >
    <Column
      v-if="selectionMode"
      :selection-mode
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

      <!-- <template #body="slotProps">
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
      </template> -->
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
      <div class="empty">
        <slot name="empty">
          <DashboardTableEmpty
            v-if="!isLoading"
          />
        </slot>
      </div>
    </template>

    <template #expansion="slotProps">
      <!-- <slot
        v-if="dataKey && expandedRows[slotProps.data[dataKey]]"
        name="expansion"
        v-bind="slotProps"
      /> -->
      <slot
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
        v-if="!hidePager && data?.paging"
        v-model:limit="query.limit"
        :paging="data?.paging"
        @change-cursor="changeCursor"
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

  :deep(.selection) {
    width: 20px;
  }

  :deep(.p-datatable-emptymessage) {
    height: 140px;
    background: transparent;

    > td {
      border: none;
    }
  }

  :deep(.p-datatable-column-header-content) {
    text-wrap: balance;
  }
  .empty {
    min-height: 400px;
  }
}

.toggle {
  cursor: pointer;
}
</style>
