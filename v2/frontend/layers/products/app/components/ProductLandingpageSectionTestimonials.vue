<template>
  <div class="py-3xl w-full">
    <BaseCard>
      <div class="flex flex-col text-center gap-y-4xl">
        <p class="text-xl">
          {{ $t('products.landing_page.testimonials.title') }}
        </p>
        <div class="flex flex-wrap sm:flex-row gap-x-5xl gap-y-3xl justify-center">
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
