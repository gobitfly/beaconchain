<script lang="ts" setup>
</script>

<template>
  <div class="font-urbanist">
    <NuxtLayout name="base">
      hello from test
      <div>
        <BaseButton
          aria-disabled="true"
          variant="primary"
        >
          Login
        </BaseButton>
      </div>
      <div>
        <BaseButton variant="secondary">
          Login
        </BaseButton>
      </div>
    </NuxtLayout>
  </div>
</template>
