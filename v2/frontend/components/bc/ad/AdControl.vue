<script setup lang="ts">
const { ads } = useCurrentAds()

const top = computed(() => ads.value?.find(c => c.jquery_selector === 'top'))
const rest = computed(() =>
  ads.value?.filter(c => c.jquery_selector !== 'top'),
)
</script>

<template>
  <ClientOnly>
    <BcAdComponent
      v-if="top"
      id="revive_top"
      :ad="top"
    />
    <Teleport
      v-for="config in rest"
      :key="config.jquery_selector"
      :to="config.jquery_selector"
    >
      <BcAdComponent :ad="config" />
    </Teleport>
  </ClientOnly>
</template>

<style lang="scss" scoped></style>
