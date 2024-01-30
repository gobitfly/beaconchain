<script setup lang="ts">
interface Props {
  value?: string
}

const props = defineProps<Props>()

const label = computed(() => {
  if (!props.value) {
    return ''
  }
  // TODO: handly currency conversion depending on the selected currency
  return formatWeiToEth(props.value)
})

</script>
<template>
  {{ label }}
</template>
