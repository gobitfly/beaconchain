<script setup lang="ts">
interface Props {
  validators: number[],
  groupId: number,
  context: 'group' | 'sync' | 'propsoal'
}
const props = defineProps<Props>()
// TODO: render validators as links
// TODO: Add icon and link to validators modal and pass title and subtitle based on groupId (mapped with with group name, from dashboard overview store) and context

</script>
<template>
  <div>{{ props.validators?.join('') }}</div>
</template>
