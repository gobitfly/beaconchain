<script setup lang="ts">
import type { BaseNavigationItem } from '#layers/base/app/components/BaseNavigationItem.vue'

const { navigateToV1Login } = useV1Login()
const { t: $t } = useTranslation()
const emit = defineEmits<{
  (e: 'open'): void,
}>()
const handleClick = () => {
  emit('open')
}
defineProps<{
  items: BaseNavigationItem[],
}>()
</script>

<template>
  <nav class="p-2xl">
    <div class="max-w-8xl mx-auto grid grid-cols-[auto_1fr_auto] gap-2xl items-center">
      <BaseButtonIcon
        screenreader-text="base.common.open_navigation"
        class="sm:hidden"
        name="menu-2"
        variant="secondary"
        @click="handleClick"
      />
      <NuxtLink
        to="/"
        class="flex gap-sm p-md rounded-full focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300 w-fit"
      >
        <BaseScreenreaderOnly screenreader-text="base.beaconchain_homepage" />
        <TheLogoMark
          width="1.25rem"
          class="aspect-square"
        />
        <TheLogoType
          class="hidden md:inline"
          width="6.25rem"
        />
      </NuxtLink>
      <ul class="hidden sm:flex gap-xl justify-center">
        <li
          v-for="{ icon, label, to } in items"
          :key="label"
        >
          <BaseNavigationItem
            :icon
            :label
            :to
          />
        </li>
      </ul>
      <BaseButton
        @click="navigateToV1Login"
      >
        {{ $t('base.common.log_in') }}
      </BaseButton>
    </div>
  </nav>
</template>
