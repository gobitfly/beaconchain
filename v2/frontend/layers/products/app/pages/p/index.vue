<script setup lang="ts">
import type { BaseNavigationItem } from '#layers/base/app/components/BaseNavigationItem.vue'

const isOpen = ref(false)
const { t: $t } = useTranslation()
const id = {
  api: 'api',
  explorer: 'explorer',
  stackingHub: 'stacking-hub',
}
const items: BaseNavigationItem[] = [
  {
    icon: 'file-code-2',
    label: $t('products.api'),
    to: `#${id.api}`,
  },
  {
    icon: 'coins',
    label: $t('products.stacking_hub'),
    to: `#${id.stackingHub}`,
  },
  {
    icon: 'compass',
    label: $t('products.explorer'),
    to: `#${id.explorer}`,
  },
]
const navigationLeft = useTemplateRef<HTMLDialogElement>('navigationLeft')
</script>

<template>
  <div>
    <NuxtLayout name="base">
      <template #header>
        <BaseNavigation
          :items
          @open="isOpen = !isOpen"
        />
        <BaseNavigationLeft
          ref="navigationLeft"
          :items
          :is-open
          @close="isOpen = false"
        />
        <section>
          <h2 class="text-center mt-[4rem]">
            {{ $t('products.landing_page.search.title') }}
          </h2>
        </section>
        <section>
          <h2
            :id="id.api"
            class="text-center mt-[40rem]"
          >
            {{ $t('products.landing_page.api.title') }}
          </h2>
        </section>
        <section>
          <h2
            :id="id.stackingHub"
            class="text-center mt-[40rem]"
          >
            {{ $t('products.landing_page.stacking_hub.title') }}
          </h2>
        </section>
        <section>
          <h2
            :id="id.explorer"
            class="text-center mt-[40rem]"
          >
            {{ $t('products.landing_page.explorer.title') }}
          </h2>
        </section>
      </template>
    </NuxtLayout>
  </div>
</template>
