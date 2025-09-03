<script lang="ts" setup>
</script>

<template>
  <div>
    <NuxtLayout name="base">
      Primary Button
      <section class="grid gap-sm">
        <div>
          <BaseButton
            variant="primary"
          >
            Login
          </BaseButton>
        </div>
        <div>
          <BaseButton
            aria-disabled="true"
            variant="primary"
          >
            Login
          </BaseButton>
          (aria-disabled=true)
        </div>
        <div>
          <BaseButton
            disabled
            variant="primary"
          >
            Login
          </BaseButton>
          (disabled)
        </div>
      </section>
    </NuxtLayout>
  </div>
</template>
