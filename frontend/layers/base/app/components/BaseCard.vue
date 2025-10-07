<script setup lang="ts">
// https://github.com/vuejs/core/issues/8952
import type { NuxtLinkProps } from '#app'
import type { BaseHeadings } from '~/layers/base/app/components/BaseHeading.vue'
import type { IconName } from '~/layers/base/app/components/BaseIcon.vue'

const {
  borderRadius = '4xl',
  elevation = '1',
  gap = '2xl',
} = defineProps<
  {
    borderRadius?: '4xl' | 'xl',
    elevation?: '1' | '2',
    gap?: '5xl' | 'xl',
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
  <div
    class="p-3xl rounded-4xl flex flex-col gap-2xl w-full border-1"
    :class="[
      elevation === '1' && 'dark:bg-gray-950 border-gray-900',
      elevation === '2' && 'dark:bg-gray-900 border-gray-900',
      borderRadius === '4xl' && 'rounded-4xl',
      borderRadius === 'xl' && 'rounded-xl',
      gap === 'xl' && 'gap-xl',
      gap === '5xl' && 'gap-5xl',
    ]"
  >
    <slot name="header">
      <header
        v-if="title"
        class="flex gap-md justify-between"
      >
        <LazyBaseIcon
          v-if="titleIcon"
          :name="titleIcon"
          class="shrink-0 text-2xl justify-self-end"
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
      class="pt-3xl mt-auto"
    >
      <slot name="footer" />
    </div>
  </div>
</template>
