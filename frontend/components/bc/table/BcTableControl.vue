<script setup lang="ts">
interface Props {
  chartDisabled?: boolean,
  disabledFilter?: boolean,
  searchPlaceholder?: string,
  title?: string,
}
const props = defineProps<Props>()

const emit = defineEmits<{ (e: 'setSearch', value?: string): void }>()

const tableIsShown = ref(true)

const useAbsoluteValues = defineModel<boolean | null>({ default: null })

const onInput = (value: string) => {
  emit('setSearch', value)
}
</script>

<template>
  <slot name="bc-table-header">
    <div class="bc-table-header">
      <div class="side left">
        <BcIconToggle
          v-if="$slots.chart"
          v-model="tableIsShown"
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
        </BcIconToggle>
        <BcIconToggle
          v-if="useAbsoluteValues !== null && tableIsShown"
          v-model="useAbsoluteValues"
        >
          <template #trueIcon>
            <BcIcon
              size="sm"
              name="hashtag"
            />
          </template>
          <template #falseIcon>
            <BcIcon
              size="sm"
              name="percent"
            />
          </template>
        </BcIconToggle>
        <slot name="header-left" />
      </div>

      <slot
        name="header-center"
        :table-is-shown
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
          :table-is-shown
        />
        <BcContentFilter
          v-if="props.searchPlaceholder && tableIsShown"
          :search-placeholder="props.searchPlaceholder"
          :disabled-filter
          class="search"
          @filter-changed="onInput"
        />
      </div>
    </div>
  </slot>
  <slot name="bc-table-sub-header" />
  <slot
    v-if="tableIsShown"
    name="table"
  />
  <slot
    v-else
    name="chart"
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
