<script setup lang="ts">
import type { MenuItem } from 'primevue/menuitem'

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
      label: $t('header.megamenu.dashboard'),
      url: '/dashboard',
    },
    {
      label: $t('header.megamenu.explorer'),
      url: `${v1Domain}`,

    },
    {
      label: $t('header.megamenu.pricing'),
      url: '/pricing',
    },
    ...(hasV1Notifications.value
      ? [
          {
            label: $t('header.megamenu.notifications_v1'),
            url: `${v1Domain}/user/notifications`,
          },
          {
            label: $t('header.megamenu.notifications_v2'),
            url: '/notifications',
          },
        ]
      : [ {
          label: $t('header.megamenu.notifications'),
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
    <MegaMenu
      ref="megaMenu"
      :model="items"
      :breakpoint
    >
      <template #item="{ item }">
        <span class="p-menuitem-link">
          <BcLink
            v-if="item.url"
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
          <div
            v-else
            class="pointer p-menuitem-text"
            :class="[item.class]"
            @click="item.command?.(null as any)"
          >
            {{ item.label }}
          </div>
        </span>
      </template>
    </MegaMenu>
  </ClientOnly>
</template>

<style lang="scss" scoped>
.iconSpacing {
  width: 25px;
  position: relative;

  img,
  svg,
  i {
    position: absolute;
    transform: translateY(-50%);
    max-width: 16px;
  }
}
</style>
