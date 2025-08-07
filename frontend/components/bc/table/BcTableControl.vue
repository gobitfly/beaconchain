<script setup lang="ts">
const props = defineProps<{
  chartDisabled?: boolean,
  disabledFilter?: boolean,
  searchPlaceholder?: string,
  title?: string,
}>()

// const emit = defineEmits<{ (e: 'setSearch', value?: string): void }>()

const search = defineModel<string>('search', {
  default: '',
})
const tabView = defineModel<'chart' | 'table'>('tabView', {
  default: 'table',
})
const isTable = computed(() => {
  return tabView.value === 'table'
})
</script>

<template>
  <slot name="bc-table-header">
    <div class="bc-table-header">
      <div class="side left">
        <BcToggleIcon
          v-if="$slots.chart"
          v-model="tabView"
          :true-value="'table'"
          :false-value="'chart'"
          :disabled="chartDisabled"
        >
          <template #trueIcon>
            <BcIcon
              name="table"
              size="sm"
            />
          </template>
          <template #falseIcon>
            <BcIcon
              name="chart-column"
              size="sm"
            />
          </template>
        </BcToggleIcon>
        <slot
          v-if="isTable"
          name="value-format"
        />

        <slot name="header-left" />
      </div>

      <slot
        name="header-center"
        :is-table
      >
        <div
          v-if="props.title"
          class="h1"
        >
          {{ props.title }}
        </div>
      </slot>
      <div class="side right">
        <slot
          name="header-right"
          :is-table
        />
        <BcContentFilter
          v-if="props.searchPlaceholder && isTable"
          v-model="search"
          :search-placeholder="props.searchPlaceholder"
          :disabled-filter
          class="search"
        />
      </div>
    </div>
  </slot>
  <slot name="bc-table-sub-header" />
  <slot
    v-if="isTable"
    name="table"
  />
  <slot
    v-else
    name="chart"
    v-bind="{ isTable }"
  />
</template>

<style lang="scss" scoped>
.bc-table-header {
  height: 70px;
  padding: 0 var(--padding-large);
  width: 100%;
  display: flex;
  align-items: center;
  gap: var(--padding);
  flex-shrink: 0;

  .side {
    flex-grow: 1;
    flex-basis: 0;
    display: flex;
    & + h1 {
      width: 180px;
    }

    &.left {
      gap: var(--padding);
    }

    &.right {
      justify-content: flex-end;
      .search {
        z-index: 3;
      }
    }
  }
}

.toggle {
  cursor: pointer;
}
</style>
