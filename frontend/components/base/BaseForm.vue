<script lang="ts" setup></script>

<template>
  <form
    class="flex flex-col gap-y-1"
    novalidate
    @submit.prevent="$emit('submit')"
  >
    <slot />
  </form>
</template>
