<script setup lang="ts">
import type { BaseNavigationItem } from '~/layers/base/app/components/BaseNavigationItem.vue'

defineProps<{
  isOpen: boolean,
  items: BaseNavigationItem[],
}>()
const emit = defineEmits<{
  (e: 'close'): void,
}>()
</script>

<template>
  <Teleport to="#teleports">
    <transition
      enter-from-class="-translate-x-full sm:-translate-x-full"
      enter-active-class="transition transform duration-200 ease-in-out"
      leave-to-class="-translate-x-full sm:-translate-x-full"
      leave-active-class="transition transform duration-200 ease-in-out"
    >
      <LazyBaseNavigationLeftContent
        v-if="isOpen"
        v-bind="$attrs"
        :items
        :is-open
        @close="emit('close')"
      />
    </transition>
  </Teleport>
</template>

<style scoped></style>
