<script setup lang="ts" generic="WORKAROUND_FOR_CONDITIONAL_PROPS">
// https://github.com/vuejs/core/issues/8952
import type { BaseHeadings } from '~/layers/base/app/components/BaseHeading.vue'
import type { IconName } from '~/layers/base/app/components/BaseIcon.vue'

defineProps<
  {
    titleIcon?: IconName,
  } & (
    | {
      title: string,
      titleIs: BaseHeadings,
    }
    | {
      title?: never,
      titleIs?: never,
    }
  )
>()
</script>

<template>
  <div class="dark:bg-gray-950 p-3xl rounded-4xl flex flex-col gap-2xl">
    <div
      v-if="$slots.header || title"
      class="flex gap-md pb-3xl"
    >
      <slot name="header">
        <LazyBaseIcon
          v-if="titleIcon"
          :name="titleIcon"
          class="text-2xl"
        />
        <LazyBaseHeading
          :is="titleIs"
          v-if="title && titleIs"
          size="xs"
        >
          {{ title }}
        </LazyBaseHeading>
      </slot>
    </div>
    <slot />
    <div
      v-if="$slots.footer"
      class="pt-3xl mt-auto"
    >
      <slot name="footer" />
    </div>
  </div>
</template>
