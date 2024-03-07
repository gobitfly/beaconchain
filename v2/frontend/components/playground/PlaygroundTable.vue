<script lang="ts" setup>
import { warn } from 'vue'

const data = {
  paging: {
    total_count: 4
  },
  data: [
    {
      id: 1,
      name: 'item 1',
      value1: 'value 1 1',
      value2: 'value 2 1'
    },
    {
      id: 2,
      name: 'item 2',
      value1: 'value 1 2',
      value2: 'value 2 2'
    },
    {
      id: 3,
      name: 'item 3',
      value1: 'value 1 3',
      value2: 'value 2 3'
    },
    {
      id: 4,
      name: 'item 4',
      value1: 'value 1 4',
      value2: 'value 2 4'
    }
  ]
}

const selected = ref()

watch(selected, () => {
  warn('selected', selected)
}, { deep: true })

</script>

<template>
  <div class="container">
    <BcTable
      v-model:selection="selected"
      data-key="id"
      selection-mode="multiple"
      :data="data"
      :cursor="0"
      :page-size="10"
      :expandable="true"
    >
      <Column field="id" header="Id" />
      <Column field="name" header="Name" />
      <Column field="value1" haeader="Value 1" />
      <Column field="value2" header="Value 2" />
    </BcTable>

    Selected: {{ selected }}
  </div>
</template>

<style lang="scss" scoped>
.container {
  display: fex;
  gap: 20px;
  padding: 20px;
}
</style>
