<script setup lang="ts">
const { url } = useV1Login()

const videoRef = useTemplateRef('videoRef')

const playVideo = async () => {
  if (videoRef.value) {
    try {
      await videoRef.value.play()
    }
    catch {
      // Handle autoplay restrictions silently
    }
  }
}

// Play animation when entering viewport to attract user attention
const { stop: stopIntersectionObserver } = useIntersectionObserver(
  videoRef,
  ([ entry ]) => {
    if (entry && entry.isIntersecting) {
      playVideo()
    }
  },
  {
    rootMargin: '0px 0px -50px 0px',
    threshold: 0.5,
  },
)

const playOnInteraction = () => {
  playVideo()
}

// Cleanup on unmount
onUnmounted(() => {
  stopIntersectionObserver()
})
</script>

<template>
  <div class="w-full flex flex-col gap-4xl justify-center items-center">
    <BaseHeading
      is="h2"
      id="staking-hub"
      size="lg"
      class="text-center "
    >
      {{ $t('products.landing_page.staking_hub.title') }}
    </BaseHeading>
    <p>{{ $t('products.landing_page.staking_hub.description') }}</p>
    <BaseButton
      leading-icon="coins"
      trailing-icon="arrow-up-right"
      variant="branded"
      size="xl"
      :to="url"
    >
      {{ $t('products.landing_page.staking_hub.action.go_to_stakinghub') }}
    </BaseButton>
    <div
      class="grid grid-cols-1 md:grid-cols-3 md:flex-row gap-xl min-h-[var(--stakinghub-card-height)] lg:min-h-[var(--stakinghub-card-height-lg)]"
      style="--stakinghub-card-height: 31rem; --stakinghub-card-height-lg: 38.75rem;"
    >
      <BaseCard
        title-is="h3"
        title-icon="file-code"
        :title="$t('products.landing_page.staking_hub.cards.notifications.title')"
        class="min-h-[var(--stakinghub-card-height)]"
      >
        <p class="text-md font-semibold">
          {{ $t('products.landing_page.staking_hub.cards.notifications.subtitle') }}
        </p>
        <p class="text-sm">
          {{ $t('products.landing_page.staking_hub.cards.notifications.description') }}
        </p>
        <div class="flex-grow px-5xl flex items-center justify-center">
          <BaseIcon
            name="bell"
            class="size-[6.25rem]"
          />
        </div>
        <template #footer>
          <BaseButton
            variant="branded"
            size="xl"
            full
            to="/notifications"
          >
            {{ $t('products.landing_page.staking_hub.cards.notifications.action') }}
          </BaseButton>
        </template>
      </BaseCard>
      <BaseCard
        title-is="h3"
        title-icon="file-code"
        :title="$t('products.landing_page.staking_hub.cards.staking_mobile_app.title')"
        class="min-h-[var(--stakinghub-card-height)] relative overflow-hidden focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300 [&>div:first-child]:relative [&>div:first-child]:z-10"
        tabindex="0"
        @mouseenter="playOnInteraction"
        @focus="playOnInteraction"
      >
        <video
          ref="videoRef"
          poster="/assets-2usdf/img/rotating-mobile-frame.webp"
          class="absolute top-[0] left-[0] w-full h-full object-cover z-0"
          muted
          playsinline
          preload="metadata"
        >
          <source
            src="/assets-2usdf/img/rotating-mobile.mp4"
            type="video/mp4"
          >
        </video>
        <!-- Dark overlay for text readability -->
        <div class="absolute top-[0] left-[0] w-full h-full bg-black/70 z-5" />
        <div class="relative z-10 p-6 text-white flex flex-col gap-2xl">
          <p class="text-md font-semibold">
            {{ $t('products.landing_page.staking_hub.cards.staking_mobile_app.subtitle') }}
          </p>
          <p>{{ $t('products.landing_page.staking_hub.cards.staking_mobile_app.description') }}</p>
        </div>
        <template #footer>
          <div class="flex gap-2xl justify-center items-start relative z-10">
            <NuxtLink
              to="https://apps.apple.com/app/beaconchain-dashboard/id1541822121"
              class="focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
              target="_blank"
            >
              <span class="sr-only">{{ $t('products.landing_page.staking_hub.cards.staking_mobile_app.action-download-appstore') }}</span>
              <img
                src="/assets-2usdf/img/app-store-btn.svg"
                class="w-auto h-[2.56rem]"
              >
            </NuxtLink>
            <NuxtLink
              to="https://play.google.com/store/apps/details?id=in.beaconcha.mobile"
              class="focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
              target="_blank"
            >
              <span class="sr-only">{{ $t('products.landing_page.staking_hub.cards.staking_mobile_app.action-download-playstore') }}</span>
              <img
                src="/assets-2usdf/img/play-store-btn.svg"
                class="w-auto h-[2.875rem]"
              >
            </NuxtLink>
          </div>
        </template>
      </BaseCard>
      <BaseCard
        title-is="h3"
        title-icon="file-code"
        :title="$t('products.landing_page.staking_hub.cards.validator_dashboards.title')"
        class="min-h-[var(--stakinghub-card-height)] bg-[url('/assets-2usdf/img/validator-bg.svg')] bg-[length:440px_440px] bg-[position:center_calc(100%+180px)] bg-no-repeat"
      >
        <p class="text-md font-semibold">
          {{ $t('products.landing_page.staking_hub.cards.validator_dashboards.subtitle') }}
        </p>
        <p>{{ $t('products.landing_page.staking_hub.cards.validator_dashboards.description') }}</p>
        <template #footer>
          <BaseButton
            variant="branded"
            size="xl"
            full
            to="/dashboard"
          >
            {{ $t('products.landing_page.staking_hub.cards.validator_dashboards.action') }}
          </BaseButton>
        </template>
      </BaseCard>
    </div>
  </div>
</template>
