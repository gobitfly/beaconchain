<script setup lang="ts">
import type { BaseNavigationItem } from '#layers/base/app/components/BaseNavigationItem.vue'

defineProps<{
  items: BaseNavigationItem[],
}>()
const {
  navigateToV1Login,
} = useV1Login()
const { t: $t } = useTranslation()
const emit = defineEmits<{
  (e: 'open'): void,
}>()
const handleClick = () => {
  emit('open')
}
const { hasSession } = useUserSession()
const v1Domain = useV1Domain()
</script>

<template>
  <nav class="p-2xl isolate">
    <div class="max-w-8xl mx-auto grid gap-2xl [&>*]:[grid-area:1/1]">
      <span class="flex gap-2xl items-center z-10 w-fit">
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
      </span>
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
      <span class="justify-self-end">
        <BaseButtonIcon
          v-if="hasSession"
          variant="secondary"
          screenreader-text="base.action.open_user_menu"
          name="user"
          @click="navigateTo(`${v1Domain}/user/settings`, { external: true })"
        />
        <BaseButton
          v-else
          @click="navigateToV1Login"
        >
          {{ $t('base.common.log_in') }}
        </BaseButton>
      </span>
    </div>
  </nav>
</template>
