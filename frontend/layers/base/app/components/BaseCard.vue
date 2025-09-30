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
  <div class="dark:bg-gray-950 p-3xl rounded-4xl flex flex-col gap-2xl w-full">
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
