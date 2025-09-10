<script setup lang="ts">
const { navigateToV1Login } = useV1Login()
const { t: $t } = useTranslation()
</script>

<template>
  <nav class="p-2xl flex justify-between items-center">
    <NuxtLink to="/">
      <span class="sr-only">
        {{ $t('base.beaconchain_homepage') }}
      </span>
      <TheLogo
        class="text-black dark:text-white"
      />
    </NuxtLink>
    <ul class="flex gap-xl">
      <li>
        <a
          class="font-semibold flex gap-md items-center p-md bg-linear-to-b rounded-full from-gray-300 to-gray-200 text-black dark:from-charcoal-600 dark:to-charcoal-700 dark:text-white opacity-90 hover:opacity-95 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
          href=""
        >
          <BaseIcon name="file-code-2" />
          {{ $t('base.products.api') }}
        </a>
        <!-- <BaseButton
          variant="secondary"
          leading-icon="file-code-2"
        >
          {{ $t('base.products.api') }}
        </BaseButton> -->
      </li>
      <li>
        <a
          class="font-semibold flex gap-md items-center p-md bg-linear-to-b rounded-full from-gray-300 to-gray-200 text-black dark:from-charcoal-600 dark:to-charcoal-700 dark:text-white opacity-90 hover:opacity-95 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
          href=""
        >
          <BaseIcon name="coins" />
          {{ $t('base.products.stacking_hub') }}
        </a>
        <!-- <BaseButton
          leading-icon="coins"
          variant="secondary"
        >
          {{ $t('base.products.stacking_hub') }}
        </BaseButton> -->
      </li>
      <li>
        <a
          class="font-semibold flex gap-md items-center p-md bg-linear-to-b rounded-full from-gray-300 to-gray-200 text-black dark:from-charcoal-600 dark:to-charcoal-700 dark:text-white opacity-90 hover:opacity-95 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
          href=""
        >
          <BaseIcon name="compass" />
          {{ $t('base.products.explorer') }}
        </a>
        <!-- <BaseButton
          leading-icon="compass"
          variant="secondary"
        >
          {{ $t('base.products.explorer') }}
        </BaseButton> -->
      </li>
    </ul>
    <BaseButton
      @click="navigateToV1Login"
    >
      {{ $t('base.common.log_in') }}
    </BaseButton>
  </nav>
</template>
