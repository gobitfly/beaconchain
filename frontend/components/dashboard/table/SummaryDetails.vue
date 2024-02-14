<script setup lang="ts">
import type { SummaryDetailS, VDBGroupSummaryResponse } from '~/types/dashboard/summary'

interface Props {
  groupId: number,
}
const props = defineProps<Props>()

const { width } = useWindowSize()

type SummaryTableData = {
  property: string
  detail: SummaryDetailS
}

const data = computed(() => {
  if (!props.groupId) {
    return null
  }
  return {} as VDBGroupSummaryResponse
})

</script>
<template>
  <DataTable :value="slotProps.data.orders">
    <Column field="id" />
    <Column field="customer" />
    <Column field="date" />
    <Column field="amount">
      <template #body="slotProps">
        {{ slotProps.data.amount }}
      </template>
    </Column>
    <Column field="status">
      <template #body="slotProps">
        {{ slotProps.data }}
      </template>
    </Column>
    <Column>
      <template #body>
        <Button icon="pi pi-search" />
      </template>
    </Column>
  </DataTable>
</template>

<style lang="scss" scoped>
</style>
