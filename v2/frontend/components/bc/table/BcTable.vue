<script setup lang="ts">
import type { ApiPagingResponse } from '~/types/api/common'
import type { Cursor } from '~/types/datatable'

interface Props {
  cursor?: Cursor,
  dataKey?: string, // Unique identifier for a data row
  pageSize?: number,
  data?: ApiPagingResponse<any>,
  expandable?: boolean,
  isRowExpandable?: (item: any) => boolean,
  selectionMode?: 'multiple' | 'single'
  tableClass?: string
  addSpacer?: boolean
}
const props = defineProps<Props>()

const emit = defineEmits<{(e: 'setCursor', value: Cursor): void, (e: 'setPageSize', value: number): void }>()

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
      delete expandedRows.value[item[props.dataKey!]]
    } else if (!props.isRowExpandable || props.isRowExpandable(item)) {
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
      delete expandedRows.value[item[props.dataKey]]
    }
  } else {
    expandedRows.value[item[props.dataKey]] = true
  }
  expandedRows.value = { ...expandedRows.value }
}

const setCursor = (value: Cursor) => {
  toggleAll(true)
  emit('setCursor', value)
}

const setPageSize = (value: number) => {
  toggleAll(true)
  emit('setPageSize', value)
}

watch(() => props.expandable, (expandable) => {
  if (!expandable) {
    toggleAll(true)
  }
})

</script>

<template>
  <DataTable
    v-model:expandedRows="expandedRows"
    class="bc-table"
    sort-mode="single"
    lazy
    :value="data?.data"
    :data-key="dataKey"
  >
    <Column v-if="selectionMode" :selection-mode="selectionMode" class="selection" />
    <Column v-if="expandable" expander class="expander">
      <template #header>
        <IconChevron class="toggle" :direction="allExpanded ? 'bottom' : 'right'" @click.stop.prevent="toggleAll()" />
      </template>

      <template #body="slotProps">
        <IconChevron
          v-if="!isRowExpandable || isRowExpandable(slotProps.data)"
          class="toggle"
          :direction="dataKey && expandedRows[slotProps.data[dataKey]] ? 'bottom' : 'right'"
          @click.stop.prevent="toggleItem(slotProps.data)"
        />
      </template>
    </Column>
    <slot />
    <Column v-if="addSpacer" field="space_filler">
      <template #body>
        <span /> <!--used to fill up the empty space so that the last column does not strech endlessly -->
      </template>
    </Column>

    <template #expansion="slotProps">
      <slot v-if="dataKey && expandedRows[slotProps.data[dataKey]]" name="expansion" v-bind="slotProps" />
    </template>

    <template #loading>
      <BcLoadingSpinner class="spinner" :loading="true" alignment="center" />
    </template>
    <template #footer>
      <BcTablePager
        v-if="data?.paging"
        :page-size="pageSize ?? 0"
        :paging="data?.paging"
        :cursor="cursor"
        @set-cursor="setCursor"
        @set-page-size="setPageSize"
      />
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
  }
}

.toggle {
  cursor: pointer;
}
</style>
