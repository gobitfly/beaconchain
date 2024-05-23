<script setup lang="ts">
import {
  faTable
} from '@fortawesome/pro-solid-svg-icons'
import {
  faChartColumn
} from '@fortawesome/pro-regular-svg-icons'

interface Props {
  title?: string,
  searchPlaceholder?: string,
  disabledFilter?: boolean,
  chartDisabled?: boolean
}
const props = defineProps<Props>()

const emit = defineEmits<{(e: 'setSearch', value?: string): void }>()

const tableIsShown = ref(true)

const onInput = (value: string) => {
  emit('setSearch', value)
}

</script>
<template>
  <slot name="bc-table-header">
    <div class="bc-table-header">
      <div class="side">
        <BcIconToggle v-if="$slots.chart" v-model="tableIsShown" :true-icon="faTable" :false-icon="faChartColumn" :disabled="chartDisabled" />
        <slot name="header-left" />
      </div>

      <slot name="header-center">
        <div v-if="props.title" class="h1">
          {{ props.title }}
        </div>
      </slot>
      <div class="side right">
        <slot name="header-right" />
        <BcContentFilter
          v-if="props.searchPlaceholder && tableIsShown"
          :search-placeholder="props.searchPlaceholder"
          :disabled-filter="disabledFilter"
          @filter-changed="onInput"
        />
      </div>
    </div>
  </slot>
  <slot name="bc-table-sub-header" />
  <slot v-if="tableIsShown" name="table" />
  <slot v-else name="chart" />
</template>

<style lang="scss" scoped>
.bc-table-header {
  height: 70px;
  padding: 0 var(--padding-large);
  width: 100%;
  display: flex;
  align-items: center;
  gap: var(--padding);

  .side {
    flex-grow: 1;
    flex-basis: 0;
    &+h1 {
      width: 180px;
    }

    &.right {
      display: flex;
      justify-content: flex-end;
    }
  }
}

.toggle {
  cursor: pointer;
}
</style>
