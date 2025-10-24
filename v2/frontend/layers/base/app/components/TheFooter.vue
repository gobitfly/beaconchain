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
    href: '/products/api/docs', icon: 'file-code', text: $t('products.api.name'),
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
    href: externalLink.beaconchain.shop, icon: 'shirt', text: $t('base.footer.resources.links.swag_shop'),
  },
  {
    href: '/products/pricing', icon: 'code', text: $t('base.footer.resources.links.api_pricing'),
  },
  {
    href: '/products/contact-sales', icon: 'mail-up', text: $t('base.footer.resources.links.contact_sales'),
  },
  {
    href: externalLink.beaconchain.status, icon: 'circle-check-filled', text: $t('base.footer.resources.links.site_status'),
  },
]
</script>

<template>
  <div
    class="mx-auto flex max-w-24xl flex-col gap-3xl bg-gray-50 px-xl pt-3xl pb-7xl sm:mb-xl sm:gap-5xl sm:rounded-4xl sm:px-5xl sm:py-7xl dark:bg-gray-950 [&_*:focus-visible]:rounded-3xl [&_*:focus-visible]:outline-offset-4"
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

    <div class="flex flex-col gap-3xl sm:flex-row">
      <div class="flex flex-col gap-lg sm:w-1/4">
        <BaseHeading
          is="h3"
          size="xxs"
        >
          {{ $t('base.footer.services.title') }}
        </BaseHeading>
        <ul class="flex flex-col gap-md">
          <li
            v-for="link in productLinks"
            :key="link.href"
            class="w-fit text-xs"
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
          size="xxs"
        >
          {{ $t('base.footer.resources.title') }}
        </BaseHeading>
        <ul class="flex flex-col gap-md">
          <li
            v-for="link in resourceLinks"
            :key="link.href"
            class="w-fit text-xs"
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

      <div class="my-lg flex flex-col gap-xl sm:my-unset sm:ml-auto">
        <BaseHeading
          is="h3"
          size="xxs"
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
              :to="externalLink.beaconchain.discord"
              name="brand-discord"
              variant="secondary"
              screenreader-text="base.footer.links.text_discord"
            />
          </li>
          <li>
            <BaseButtonIcon
              :to="externalLink.beaconchain.x"
              name="brand-x"
              variant="secondary"
              screenreader-text="base.footer.links.text_x"
            />
          </li>
          <li>
            <BaseButtonIcon
              :to="externalLink.beaconchain.github.organization"
              name="brand-github"
              variant="secondary"
              screenreader-text="base.footer.links.text_github"
            />
          </li>
        </ul>
        <BaseColormode
          class="-ml-1"
        />
      </div>
    </div>

    <div class="flex flex-col gap-lg">
      <hr class="dark:text-gray-400">
      <div class="flex flex-col gap-xl text-xs sm:flex-row sm:justify-between">
        <div class="flex justify-around sm:order-2 sm:w-1/2">
          <NuxtLink
            :to="externalLink.beaconchain.imprint"
            class="hover:text-link-500"
          >
            {{ $t('base.footer.imprint') }}
          </NuxtLink>
          <NuxtLink
            :to="externalLink.beaconchain.termsOfService"
            class="hover:text-link-500"
          >
            {{ $t('base.footer.terms_of_service') }}
          </NuxtLink>
          <NuxtLink
            :to="externalLink.beaconchain.privacyPolicy"
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
