<script setup lang="ts">
import { useBreakpoints } from '#layers/base/app/composables/useBreakpoints'
import type { BaseNavigationItem } from '~/layers/base/app/components/BaseNavigationItem.vue'

const navigation = useTemplateRef('navigation')
defineProps<{
  isOpen: boolean,
  items: BaseNavigationItem[],
}>()
onClickOutside(navigation, () => {
  emit('close')
})

const emit = defineEmits<{
  (e: 'close'): void,
}>()

const close = () => {
  emit('close')
}
const { sm } = useBreakpoints()

const previousActiveElement = ref<Element | null>()
const closeButton = useTemplateRef('closeButton')
onBeforeMount(() => {
  previousActiveElement.value = document.activeElement
})
onMounted(() => {
  closeButton.value?.$el?.focus()
})
onUnmounted(() => {
  if (previousActiveElement.value instanceof HTMLElement) {
    previousActiveElement.value.focus()
  }
})

watchEffect(() => {
  if (sm.value) {
    emit('close')
  }
})
</script>

<template>
  <div
    ref="navigation"
    :open="isOpen"
    class="fixed inset-[0] right-auto h-screen w-3xs border-r border-r-gray-200 bg-white p-2xl dark:border-r-gray-700 dark:bg-black"
    @keydown.escape="close"
  >
    <BaseScreenreaderOnly
      is="h2"
      screenreader-text="base.common.side_navigation"
    />
    <section class="flex justify-between">
      <span class="flex flex-row gap-sm">
        <TheLogoMark
          width="1.25rem"
          class="aspect-square"
        />
        <TheLogoType
          width="6.25rem"
        />
      </span>
      <BaseButtonIcon
        ref="closeButton"
        screenreader-text="base.common.close"
        class=""
        name="x"
        variant="secondary"
        @click="close"
      />
    </section>
    <nav class="mt-xl">
      <ul class="flex flex-col gap-md">
        <li
          v-for="{ icon, label, to } in items"
          :key="label"
        >
          <BaseNavigationLeftItem
            :icon
            :label
            :to
          />
        </li>
      </ul>
    </nav>
  </div>
</template>

<style scoped></style>
