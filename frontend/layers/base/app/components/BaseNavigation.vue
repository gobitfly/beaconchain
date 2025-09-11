<script setup lang="ts">
const { navigateToV1Login } = useV1Login()
const { t: $t } = useTranslation()
</script>

<template>
  <nav class="p-2xl">
    <div class="max-w-8xl mx-auto grid grid-cols-[auto_1fr_auto] gap-2xl items-center">
      <BaseButtonIcon
        class="sm:hidden"
        name="menu-2"
        variant="secondary"
      />
      <NuxtLink to="/">
        <span class="sr-only">
          {{ $t('base.beaconchain_homepage') }}
        </span>
        <span class="flex gap-sm text-black dark:text-white">
          <TheLogoMark
            width="1.25rem"
            class="aspect-square"
          />
          <TheLogoType
            class="hidden md:inline"
            width="6.25rem"
          />
        </span>
      </NuxtLink>
      <ul class="hidden sm:flex gap-xl justify-center">
        <li>
          <a
            class="font-semibold flex gap-md items-center p-md bg-linear-to-b rounded-full from-gray-300 to-gray-200 text-black dark:from-charcoal-600 dark:to-charcoal-700 dark:text-white opacity-90 hover:opacity-95 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
            href=""
          >
            <BaseIcon name="file-code-2" />
            {{ $t('base.products.api') }}
          </a>
        </li>
        <li>
          <a
            class="font-semibold flex gap-md items-center p-md bg-linear-to-b rounded-full from-gray-300 to-gray-200 text-black dark:from-charcoal-600 dark:to-charcoal-700 dark:text-white opacity-90 hover:opacity-95 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
            href=""
          >
            <BaseIcon name="coins" />
            {{ $t('base.products.stacking_hub') }}
          </a>
        </li>
        <li>
          <a
            class="font-semibold flex gap-md items-center p-md bg-linear-to-b rounded-full from-gray-300 to-gray-200 text-black dark:from-charcoal-600 dark:to-charcoal-700 dark:text-white opacity-90 hover:opacity-95 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
            href=""
          >
            <BaseIcon name="compass" />
            {{ $t('base.products.explorer') }}
          </a>
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
