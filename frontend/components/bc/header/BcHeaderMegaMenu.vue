<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faCaretRight } from '@fortawesome/pro-solid-svg-icons'

import type { MenuItem } from 'primevue/menuitem'
import MegaMenu from 'primevue/megamenu'

import {
  mobileHeaderThreshold, smallHeaderThreshold,
} from '~/types/header'

const { t: $t } = useTranslation()
const { width } = useWindowSize()
const {
  doLogout, isLoggedIn,
} = useUserStore()
const route = useRoute()
const megaMenu = ref<null | {
  mobileActive: boolean,
  toggle: (evt: Event) => void,
}>(null)

const breakpoint = `${smallHeaderThreshold}px`
const isSmallScreen = computed(() => width.value < smallHeaderThreshold)
const isMobile = computed(() => width.value < mobileHeaderThreshold)

const { has } = useFeatureFlag()

const items = computed(() => {
  let list: MenuItem[] = []

  list = [
    {
      label: $t('header.megamenu.dashboard'),
      url: '/dashboard',
    },
    {
      label: $t('header.megamenu.pricing'),
      url: '/pricing',
    },
  ]
  if (has('feature-notifications')) {
    list.push({
      label: $t('header.megamenu.notifications'),
      url: '/notifications',
    })
  }
  if (isMobile.value) {
    if (isLoggedIn.value) {
      list.push({
        command: async () => {
          await navigateTo('../user/settings')
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
      :model="[items]"
      :breakpoint
    >
      <template #item="{ item, hasSubmenu }">
        <span class="p-menuitem-link">
          <span
            v-if="item.svg || item.icon"
            class="p-menuitem-icon iconSpacing"
            data-pc-section="icon"
          >
            <component
              :is="item.svg"
              v-if="item.svg"
              class="monochromatic"
            />
            <FontAwesomeIcon
              v-else-if="item.icon"
              class="icon"
              :icon="item.icon"
            />
          </span>
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
          <FontAwesomeIcon
            v-if="hasSubmenu"
            :icon="faCaretRight"
            class="p-icon p-submenu-icon"
          />
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
