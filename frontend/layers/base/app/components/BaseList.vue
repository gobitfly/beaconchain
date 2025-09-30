<script setup lang="ts">
defineProps<
  {
    items: string[],
  }>()
</script>

<template>
  <ul>
    <li
      v-for="item in items"
      :key="item"
      class="flex gap-xs"
    >
      <slot name="marker">
        <BaseIcon
          name="check"
          aria-hidden="true"
          class="shrink-0 text-lg mt-[.125rem]"
        />
      </slot>
      <span class="">
        {{ item }}
      </span>
    </li>
  </ul>
</template>

<style scoped></style>
