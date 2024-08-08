<script setup lang="ts">
defineProps<{
  layoutAdaptability: 'low' | 'high'
}>()
</script>

<template>
  <BcLink
    to="/"
    :class="`${layoutAdaptability}-adaptability`"
  >
    <IconBeaconchainLogo alt="Beaconcha.in logo" />
    <span class="name">beaconcha.in</span>
  </BcLink>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

// do not change these three values without changing the values in MainHeader.vue and in types/header.ts accordingly
$mobileHeaderThreshold: 600px;
$smallHeaderThreshold: 1024px;
$largeHeaderThreshold: 1360px;

@mixin common() {
  display: flex;
  position: relative;
  gap: var(--padding);
  svg {
    margin-top: auto;
    height: 30px;
  }
  .name {
    display: inline-flex;
    position: relative;
    margin-top: auto;
    line-height: 22px;
    font-family: var(--logo_font_family);
    font-weight: var(--logo_font_weight);
    letter-spacing: var(--logo_small_letter_spacing);
    font-size: var(--logo_font_size);
  }
}

@mixin smaller() {
  gap: 6px;
  svg {
    height: 18px;
  }
  .name {
    line-height: 14px;
    font-size: var(--logo_small_font_size);
  }
}

.low-adaptability {
  @include common();
  svg {
    margin-bottom: 13px;
  }
  .name {
    margin-top: 14px;
    margin-bottom: 14px;
  }
  @media (max-width: $mobileHeaderThreshold) {
    @include smaller();
    svg {
      margin-bottom: 11px;
    }
    .name {
      margin-top: 11px;
      margin-bottom: 11px;
    }
  }
}

.high-adaptability {
  @include common();
  @media (max-width: $largeHeaderThreshold) {
    @include smaller();
    @media (max-width: $mobileHeaderThreshold) {
      svg {
        height: 30px;
      }
      .name {
        display: none;
      }
    }
  }
}
</style>
