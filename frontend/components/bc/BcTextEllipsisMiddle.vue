<script setup lang="ts">
const props = defineProps<{
  text: string,
}>()
const root = ref(null)

const {
  update,
  width,
} = useElementBounding(root, {
  immediate: false,
})
onMounted(() => {
  update()
})
const characterTotal = computed(() => props.text.length)

// this is estimated by manually messuring for 10_0000 characters 🙉
const averageCharacterWidth = 9

const allowedCharacterCount = computed(() => Math.round(width.value / averageCharacterWidth))

const difference = computed(() => {
  const result = characterTotal.value - allowedCharacterCount.value
  if (result <= 0) return 0
  return result
})

const indexOfCenter = computed(() => Math.round(characterTotal.value / 2))

const leftPart = computed(() => Math.round(difference.value / 2))
const rightPart = computed(() => difference.value - leftPart.value)
const left = computed(() => props.text.substring(0, indexOfCenter.value - leftPart.value))
const right = computed(() => props.text.substring(indexOfCenter.value + rightPart.value))
</script>

<template>
  <span
    ref="root"
    class="bc-text-ellipsis-middle"
  >
    <span
      class="bc-text-ellipsis-middle-left"
    >
      {{ left }}
    </span>
    <span
      v-if="difference"
    >…</span>
    <span
      class="bc-text-ellipsis-middle-left"
    >{{ right }}</span>
  </span>
</template>

<style scoped lang="scss">
.bc-text-ellipsis-middle {
  white-space: nowrap;
}
.bc-text-ellipsis-middle-left,
.bc-text-ellipsis-middle-right {
  white-space: nowrap;
}
</style>
