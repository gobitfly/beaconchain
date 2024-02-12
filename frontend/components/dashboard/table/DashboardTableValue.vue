<script setup lang="ts">
import { type TableValue } from '~/types/dashboard/table'
interface Props {
  data: TableValue
}
const props = defineProps<Props>()

</script>
<template>
  <div>{{ props }}</div>
</template>
