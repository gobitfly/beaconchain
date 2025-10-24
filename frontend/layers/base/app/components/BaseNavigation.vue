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
  <nav class="isolate bg-gray-50 p-2xl dark:bg-transparent">
    <div class="mx-auto grid max-w-8xl gap-2xl *:[grid-area:1/1]">
      <span class="z-10 flex w-fit items-center gap-2xl">
        <BaseButtonIcon
          screenreader-text="base.common.open_navigation"
          class="sm:hidden"
          name="menu-2"
          variant="secondary"
          @click="handleClick"
        />
        <NuxtLink
          to="/"
          class="flex w-fit gap-sm rounded-full p-md"
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
      <ul class="hidden justify-center gap-xl sm:flex">
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
