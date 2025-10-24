<template>
  <div class="w-full py-3xl">
    <BaseCard>
      <div class="flex flex-col gap-y-4xl text-center">
        <p class="text-xl">
          {{ $t('products.landing_page.testimonials.title') }}
        </p>
        <div class="flex flex-wrap justify-center gap-x-5xl gap-y-3xl sm:flex-row">
          <div
            v-for="i in 5"
            :key="i"
            class="flex items-center gap-md"
          >
            <TheLogoMark
              width="3rem"
              height="3rem"
              class="aspect-square"
            />
            <TheLogoType
              width="6.25rem"
            />
          </div>
        </div>
      </div>
    </BaseCard>
  </div>
</template>

<script setup lang="ts">

</script>

<style scoped>

</style>
