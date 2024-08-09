<script setup lang="ts">
import type { Paging } from '~/types/api/common'
import type { Cursor } from '~/types/datatable'

interface Props {
  cursor: Cursor,
  pageSize: number,
  paging?: Paging,
  stepperOnly?: boolean,
}
const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'setCursor', value: Cursor): void,
  (e: 'setPageSize', value: number): void,
}>()

const pageSizes = [
  5,
  10,
  25,
  50,
  100,
]

const currentOffset = computed<number>(() =>
  typeof props.cursor === 'number' ? props.cursor : 0,
)

const data = computed(() => {
  if (!props.paging) {
    return { mode: 'waiting' }
  }
  if (props.paging.total_count === undefined) {
    return {
      mode: 'cursor',
      next_cursor: props.paging.next_cursor,
      prev_cursor: props.paging.prev_cursor,
    }
  }
  const page
    = props.paging.total_count > 0
      ? 1 + Math.floor(currentOffset.value / props.pageSize)
      : 0
  const from = props.paging.total_count > 0 ? currentOffset.value + 1 : 0
  const to = Math.min(
    currentOffset.value + props.pageSize,
    props.paging.total_count,
  )
  const lastPage = Math.ceil(props.paging.total_count / props.pageSize)

  return {
    from,
    lastPage,
    mode: 'offset',
    page,
    to,
  }
})

const next = () => {
  emit(
    'setCursor',
    Math.min(
      currentOffset.value + props.pageSize,
      ((data.value.lastPage ?? 1) - 1) * props.pageSize,
    ),
  )
}

const prev = () => {
  emit('setCursor', Math.max(0, currentOffset.value - props.pageSize))
}

const first = () => {
  emit('setCursor', undefined)
}

const last = () => {
  emit('setCursor', (data.value.lastPage! - 1) * props.pageSize)
}

const setPageSize = (size: number) => {
  if (data.value.mode === 'offset') {
    // in case we increase the page size we must adjust the offset
    const off = currentOffset.value % size
    if (off > 0) {
      emit('setCursor', currentOffset.value - off)
    }
  }
  emit('setPageSize', size)
}

// in case the totalCount decreased
watch(
  () => data.value.lastPage && data.value.lastPage < data.value.page,
  (match) => {
    if (data.value.lastPage !== undefined && match) {
      last()
    }
  },
)
</script>

<template>
  <div class="bc-pageinator">
    <div class="pager">
      <template v-if="data.mode === 'offset'">
        <div
          class="item button"
          :disabled="!currentOffset"
          @click="first"
        >
          {{ $t("table.first") }}
        </div>
        <div
          class="item button"
          :disabled="!currentOffset"
          @click="prev"
        >
          <IconChevron
            class="toggle"
            direction="left"
          />
        </div>
        <div class="item current-page">
          {{ data.page }} {{ $t("table.of") }} {{ data.lastPage }}
        </div>
        <div
          class="item button"
          :disabled="data.page! >= data.lastPage!"
          @click="next"
        >
          <IconChevron
            class="toggle"
            direction="right"
          />
        </div>
        <div
          class="item button"
          :disabled="data.page! >= data.lastPage!"
          @click="last"
        >
          {{ $t("table.last") }}
        </div>
      </template>
      <template v-else-if="data.mode === 'cursor'">
        <div
          class="item button"
          :disabled="!data.prev_cursor"
          @click="first"
        >
          {{ $t("table.first") }}
        </div>
        <div
          class="item button"
          :disabled="!data.prev_cursor"
          @click="emit('setCursor', data.prev_cursor)"
        >
          <IconChevron
            class="toggle"
            direction="left"
          />
        </div>
        <div
          class="item button"
          :disabled="!data.next_cursor"
          @click="emit('setCursor', data.next_cursor)"
        >
          <IconChevron
            class="toggle"
            direction="right"
          />
        </div>
      </template>
      <Dropdown
        v-if="props.pageSize && !stepperOnly"
        :model-value="props.pageSize"
        :options="pageSizes"
        class="table small"
        @change="(event) => setPageSize(event.value)"
      />
    </div>
    <div class="very-last">
      <div
        v-if="!stepperOnly"
        class="left-info"
      >
        <slot name="bc-table-footer-left">
          <span v-if="props.paging?.total_count">
            {{
              $t("table.showing", {
                from: data.from,
                to: data.to,
                total: props.paging?.total_count,
              })
            }}
          </span>
        </slot>
      </div>
      <div
        v-if="$slots['bc-table-footer-right']"
        class="right-info"
      >
        <slot name="bc-table-footer-right" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";

.bc-pageinator {
  position: relative;
  height: 78px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  font-weight: var(--standard_text_medium_font_weight);
  margin: var(--padding) var(--padding-large);

  .very-last {
    position: absolute;
    display: flex;
    flex-direction: row;
    width: 100%;
    .left-info {
      margin-right: auto;
    }
    .right-info {
      margin-left: auto;
    }
  }

  .pager {
    display: flex;
    gap: 3px;

    .table {
      @include main.container;
      border-top-left-radius: 0;
      border-bottom-left-radius: 0;
      height: 30px;

      &.p-overlay-open {
        border-bottom-right-radius: 0;
      }
    }

    .item {
      @include main.container;
      display: flex;
      justify-content: center;
      align-items: center;
      height: 30px;
      padding: 0 22px;
      border-radius: 0;
      white-space: nowrap;

      &:has(svg) {
        padding: 0 15px;
      }

      &.button {
        &:not([disabled="true"]) {
          cursor: pointer;
        }

        &[disabled="true"] {
          pointer-events: none;
        }

        &[disabled="true"] {
          color: var(--text-color-disabled);
        }
      }

      &:first-child {
        border-top-left-radius: var(--border-radius);
        border-bottom-left-radius: var(--border-radius);
      }

      &:last-child {
        border-top-right-radius: var(--border-radius);
        border-bottom-right-radius: var(--border-radius);
      }

      @media screen and (max-width: 1399px) {
        &.current-page {
          display: none;
        }
      }
    }
  }

  @media screen and (max-width: 1399px) {
    gap: var(--padding);
    height: unset;
    .very-last {
      @media (max-width: 600px) {
        position: relative;
      }
    }
  }
}
</style>
