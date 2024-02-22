<script lang="ts" setup>
import {
  faMagnifyingGlass
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

defineEmits<{(e: 'filter-changed', value: string): void }>()

const filterVisible = ref(false)
const filter = ref<string>('')
</script>

<template>
  <span class="filter_elements_container">
    <InputText v-model="filter" placeholder="Index" :class="{visible:filterVisible}" :disabled="!filterVisible" @input="$emit('filter-changed', filter)" />
    <Button class="p-button-icon-only" :class="{filter_visible:filterVisible}" @click="filterVisible=!filterVisible">
      <FontAwesomeIcon :icon="faMagnifyingGlass" />
    </Button>
  </span>
</template>

<style lang="scss">
  .filter_elements_container {
    display: flex;
    justify-content: flex-end;

    > :first-child{
      border-top-right-radius: 0;
      border-bottom-right-radius: 0;
      height: var(--default-button-height);
      width: 0;
      opacity: 0;
      padding: 0;
      transition:
        width 0.2s ease-in-out,
        opacity 0.01s ease-in-out 0.19s,
        padding 0.2s ease-in-out;

      &.visible {
        width: 100%;
        opacity: 100%;
        padding: 4px;

        transition:
          width 0.2s ease-in-out,
          opacity 0.01s ease-in-out,
          padding 0.2s ease-in-out;
      }
    }

    > :last-child{
      flex-shrink: 0;
      border-top-left-radius: var(--border-radius);
      border-bottom-left-radius: var(--border-radius);
      transition: all 0.2s ease-in-out;

      &.filter_visible {
        border-top-left-radius: 0;
        border-bottom-left-radius: 0;
      }
    }
  }
</style>
