<script setup lang="ts">
import type { MenuItem } from 'primevue/menuitem'
import type { IconName } from '~/layers/base/app/components/BaseIcon.vue'

import {
  mobileHeaderThreshold, smallHeaderThreshold,
} from '~/types/header'

const { t: $t } = useTranslation()
const { width } = useWindowSize()
const {
  doLogout,
  hasV1Notifications,
  isLoggedIn,
} = useUserStore()
const route = useRoute()
const megaMenu = ref<null | {
  mobileActive: boolean,
  toggle: (evt: Event) => void,
}>(null)

const breakpoint = `${smallHeaderThreshold}px`
const isSmallScreen = computed(() => width.value < smallHeaderThreshold)
const isMobile = computed(() => width.value < mobileHeaderThreshold)

const v1Domain = useV1Domain()

const items = computed(() => {
  let list: MenuItem[] = []

  list = [
    {
      items: [ [ {
        items: [
          {
            icon: 'key-filled',
            label: $t('header.megamenu.api_key_management'),
            url: `${v1Domain}/user/settings#api`,
          },
          {
            icon: 'book-filled',
            label: $t('header.megamenu.api_docs'),
            url: `${v1Domain}/api/v1/docs`,
          },
          {
            icon: 'coins',
            label: $t('header.megamenu.api_pricing'),
            url: `${v1Domain}/pricing`,
          },
        ],
      } ] ],
      label: $t('products.api.name'),
      root: true,
    },
    {
      label: $t('header.megamenu.dashboard'),
      root: true,
      url: '/dashboard',
    },
    {
      label: $t('header.megamenu.explorer'),
      root: true,
      url: `${v1Domain}`,
    },
    {
      label: $t('header.megamenu.premium'),
      root: true,
      url: '/premium',
    },
    ...(hasV1Notifications.value
      ? [
          {
            label: $t('header.megamenu.notifications_v1'),
            root: true,
            url: `${v1Domain}/user/notifications`,
          },
          {
            label: $t('header.megamenu.notifications_v2'),
            root: true,
            url: '/notifications',
          },
        ]
      : [ {
          label: $t('header.megamenu.notifications'),
          root: true,
          url: '/notifications',
        } ]
    ),
  ]

  if (isMobile.value) {
    if (isLoggedIn.value) {
      list.push({
        command: async () => {
          await navigateTo(`${v1Domain}/user/settings`, { external: true })
        },
        label: $t('header.settings'),
      })
    }
  }
  if (isSmallScreen.value && isLoggedIn.value) {
    list.push({
      command: () => doLogout(),
      label: $t('header.logout'),
    })
  }
  return list
})

const isMobileMenuOpen = computed(() => megaMenu.value?.mobileActive)

const toggleMegaMenu = (evt: Event) => {
  megaMenu.value?.toggle(evt)
  document.body.focus()
}

defineExpose({
  isMobileMenuOpen,
  toggleMegaMenu,
})
</script>

<template>
  <ClientOnly>
    <PvMegaMenu
      ref="megaMenu"
      :model="items"
      :breakpoint
    >
      <template #item="{ item }">
        <BcLink
          v-if="item.root && item.url"
          class="p-menuitem-link"
          :to="item.url"
          :replace="route.path.startsWith(item.url)"
        >
          <span
            :class="[item.class]"
            class="p-menuitem-text"
          >
            <span>{{ item.label }}</span>
          </span>
        </BcLink>

        <BcButton
          v-else-if="item.root"
          :class="[item.class, 'bc-header-mega-menu__item-button']"
          @click="item.command?.(null as any)"
        >
          {{ item.label }}
          <template #icon>
            <BaseIcon
              name="chevron-down"
            />
          </template>
        </BcButton>

        <BcLink
          v-else-if="item.url"
          :to="item.url"
          :replace="route.path.startsWith(item.url)"
          class="p-menuitem-link"
        >
          <BaseIcon
            :name="item.icon as IconName"
          />
          <span
            :class="[item.class]"
            class="p-menuitem-text"
          >
            <span>{{ item.label }}</span>
          </span>
        </BcLink>
      </template>
    </PvMegaMenu>
  </ClientOnly>
</template>

<style lang="scss" scoped>
</style>
