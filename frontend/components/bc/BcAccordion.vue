<script setup lang="ts" generic="T">
defineProps<{
  items?: T[],
}>()
const isOpen = ref(false)
</script>

<template>
  <details
    class="bc-accordion"
    @click="isOpen = !isOpen"
  >
    <summary>
      <span class="bc-accordion__heading">
        <IconChevron
          :direction="isOpen ? 'bottom' : 'right'"
        />
        <slot name="headingIcon" />
        <slot name="heading" />
      </span>
    </summary>
    <BcCard class="bc-accordion__content">
      <ul class="bc-accordion-list">
        <li
          v-for="item in items"
          :key="`${item}`"
          class="bc-accordion-list__item"
        >
          <slot name="item" :item />
        </li>
      </ul>
    </BcCard>
  </details>
</template>

<style scoped lang="scss">
// summary::marker {
//   display: none;
// }
.bc-accordion {
  summary {
    list-style: none
  }
  summary::-webkit-details-marker{
    display: none;
  }
}
.bc-accordion__heading {
  display: inline-flex;
  align-items: center;
  gap: 0.625rem;
}
.bc-accordion__content {
margin-top: 0.625rem;
}
.bc-accordion-list {
  list-style: none;
}
.bc-accordion-list__item {
  display: inline;
}
.bc-accordion-list__item:not(:last-child)::after {
content: ', ';
}
</style>
