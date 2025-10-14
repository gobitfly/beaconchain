<script setup lang="ts">
import type { IconName } from '~/layers/base/app/components/BaseIcon.vue'

type Link = {
  href: string,
  icon: IconName,
  text: string,
}

const { t: $t } = useTranslation()

const productLinks: Link[] = [
  {
    href: '/products/api/docs', icon: 'file-code', text: $t('products.api'),
  },
  {
    href: '/dashboard', icon: 'coins', text: $t('products.staking_hub'),
  },
  {
    href: '/', icon: 'compass', text: $t('products.explorer'),
  },
]
const resourceLinks: Link[] = [
  {
    href: '/advertisewithus', icon: 'ad', text: $t('base.footer.resources.links.advertise'),
  },
  {
    href: '/premium', icon: 'ufo', text: $t('base.footer.resources.links.beaconchain_premium'),
  },
  {
    href: 'https://shop.beaconcha.in/#!/', icon: 'shirt', text: $t('base.footer.resources.links.swag_shop'),
  },
  {
    href: '/products/pricing', icon: 'code', text: $t('base.footer.resources.links.api_pricing'),
  },
  {
    href: '/products/contact-sales', icon: 'mail-up', text: $t('base.footer.resources.links.contact_sales'),
  },
  {
    href: 'https://status.beaconcha.in/', icon: 'circle-check-filled', text: $t('base.footer.resources.links.site_status'),
  },
]
const colorMode = useColorMode()
</script>

<template>
  <div
    class="bg-gray-50 dark:bg-gray-950 flex flex-col gap-3xl sm:gap-5xl pt-3xl px-xl pb-7xl sm:px-5xl sm:py-7xl sm:rounded-4xl sm:mb-xl max-w-24xl mx-auto [&_*:focus-visible]:rounded-3xl [&_*:focus-visible]:outline-offset-4"
  >
    <NuxtLink
      to="/"
      class="w-fit"
    >
      <span class="sr-only">
        {{ $t('base.beaconchain_homepage') }}
      </span>
      <span class="flex gap-xl">
        <TheLogoMark
          width="1.5rem"
        />
        <div>
          <TheLogoType
            class="inline stroke-3"
            width="6.25rem"
          />
          <span
            class="block text-xs text-gray-400"
            aria-hidden
          >
            {{ $t('base.footer.logo_text') }}
          </span>
        </div>
      </span>
    </NuxtLink>

    <div class="flex flex-col sm:flex-row gap-3xl">
      <div class="flex flex-col gap-lg sm:w-1/4">
        <BaseHeading
          is="h3"
          size="md"
        >
          {{ $t('base.footer.services.title') }}
        </BaseHeading>
        <ul class="flex flex-col gap-md">
          <li
            v-for="link in productLinks"
            :key="link.href"
            class="text-xs w-fit"
          >
            <NuxtLink
              :to="link.href"
              class="flex items-center gap-xs"
            >
              <BaseIcon
                :name="link.icon"
                size="1rem"
                class="opacity-60"
              />
              <span class="hover:text-link-500">{{ link.text }}</span>
            </NuxtLink>
          </li>
        </ul>
      </div>

      <div class="flex flex-col gap-lg sm:w-1/4">
        <BaseHeading
          is="h3"
          size="md"
        >
          {{ $t('base.footer.resources.title') }}
        </BaseHeading>
        <ul class="flex flex-col gap-md">
          <li
            v-for="link in resourceLinks"
            :key="link.href"
            class="text-xs w-fit"
          >
            <NuxtLink
              :to="link.href"
              class="flex items-center gap-xs"
            >
              <BaseIcon
                :name="link.icon"
                size="1rem"
                class="opacity-60"
                :class="[{ 'text-success-600 dark:text-success-500': link.text === 'Site Status' }]"
              />
              <span class="hover:text-link-500">{{ link.text }}</span>
            </NuxtLink>
          </li>
        </ul>
      </div>

      <div class="my-lg sm:my-unset sm:ml-auto flex flex-col gap-xl">
        <BaseHeading
          is="h3"
          size="md"
        >
          {{ $t('base.footer.links.title') }}
        </BaseHeading>
        <div
          class="w-[152px] text-xs text-gray-400"
        >
          {{ $t('base.footer.links.follow_us') }}
        </div>
        <ul class="flex gap-md">
          <li>
            <BaseButtonIcon
              to="https://dsc.gg/beaconchain"
              name="brand-discord"
              variant="secondary"
              screenreader-text="base.footer.links.text_discord"
            />
          </li>
          <li>
            <BaseButtonIcon
              to="https://x.com/beaconcha_in"
              name="brand-x"
              variant="secondary"
              screenreader-text="base.footer.links.text_x"
            />
          </li>
          <li>
            <BaseButtonIcon
              to="https://github.com/gobitfly"
              name="brand-github"
              variant="secondary"
              screenreader-text="base.footer.links.text_github"
            />
          </li>
        </ul>
        <BaseSwitch
          v-model="colorMode.value"
          class="w-fit -ml-[1px]"
          :values="[
            {
              label: $t('base.footer.color_mode.light'),
              key: 'light',
            },
            {
              label: $t('base.footer.color_mode.dark'),
              key: 'dark',
            },
          ]"
          :class-list="{
            trackItem: 'p-md +py-md +px-lg border border-transparent rounded-4xl text-gray-600 dark:text-gray-400 has-checked:text-black has-checked:dark:text-white has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-brand-300',
            track: 'border border-gray-100 dark:border-gray-900 flex gap-md [&>*]:grow text-center bg-white dark:bg-gray-950 rounded-4xl shadow-[0_1px_0.5px_0_rgba(255,255,255,0.08)_inset,0-1px_0_0_rgba(255,255,255,0.18)_inset]',
            thumb: 'bg-gray-50 dark:bg-gray-800 rounded-4xl shadow-[0_2px_2px_0_rgba(0,0,0,0.25),_0_0.5px_0.5px_0_rgba(255,255,255,0.12)_inset]',
          }"
          screenreader-title="base.footer.color_mode.title"
        >
          <template #first=" { label } ">
            <BaseIcon name="sun" />
            <span class="sr-only">{{ label }}</span>
          </template>
          <template #second=" { label } ">
            <BaseIcon name="moon" />
            <span class="sr-only">{{ label }}</span>
          </template>
        </BaseSwitch>
      </div>
    </div>

    <div class="flex flex-col gap-lg">
      <hr class="dark:text-gray-400">
      <div class="flex flex-col sm:flex-row gap-xl text-xs sm:justify-between">
        <div class="flex justify-around sm:order-2 sm:w-1/2">
          <NuxtLink
            to="https://beaconcha.in/imprint"
            class="hover:text-link-500"
          >
            {{ $t('base.footer.imprint') }}
          </NuxtLink>
          <NuxtLink
            to="https://storage.googleapis.com/legal.beaconcha.in/tos.pdf"
            class="hover:text-link-500"
          >
            {{ $t('base.footer.terms_of_service') }}
          </NuxtLink>
          <NuxtLink
            to="https://storage.googleapis.com/legal.beaconcha.in/privacy.pdf"
            class="hover:text-link-500"
          >
            {{ $t('base.footer.privacy') }}
          </NuxtLink>
        </div>
        <span class="flex self-center sm:order-1">{{ $t('base.footer.copyright', { year: new Date().getFullYear() }) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
