<script setup lang="ts">
// https://github.com/vuejs/core/issues/8952
import type { NuxtLinkProps } from '#app'
import type { BaseHeadings } from '~/layers/base/app/components/BaseHeading.vue'
import type { IconName } from '~/layers/base/app/components/BaseIcon.vue'

defineProps<
  {
    titleIcon?: IconName,
  } & (
    | {
      subtitle?: never,
      title?: never,
      titleIs?: never,
      titleTo?: never,
    }
    | {
      subtitle?: string,
      title: string,
      titleIs: BaseHeadings,
      titleTo?: NuxtLinkProps['to'],
    }
  )
>()
const id = useId()
</script>

<template>
  <div class="flex w-full flex-col gap-2xl rounded-4xl border border-gray-100 p-3xl dark:border-gray-900 dark:bg-gray-950">
    <slot name="header">
      <header
        v-if="title"
        class="flex justify-between gap-md"
      >
        <LazyBaseIcon
          v-if="titleIcon"
          :name="titleIcon"
          class="shrink-0 justify-self-end text-2xl"
        />
        <span class="grow">
          <LazyBaseHeading
            :is="titleIs"
            v-if="title && titleIs"
            :id
            size="xs"
          >
            {{ title }}
          </LazyBaseHeading>
          <LazyBaseText
            v-if="subtitle"
            size="sm"
            class="mt-xs"
            variant="secondary"
          >
            {{ subtitle }}
          </LazyBaseText>
        </span>
        <LazyBaseButtonIcon
          v-if="titleTo"
          :aria-labelledby="id"
          name="arrow-up-right"
          variant="tertiary"
          size="lg"
          :to="titleTo"
        />
      </header>
    </slot>
    <slot />
    <div
      v-if="$slots.footer"
      class="mt-auto pt-3xl"
    >
      <slot name="footer" />
    </div>
  </div>
</template>
