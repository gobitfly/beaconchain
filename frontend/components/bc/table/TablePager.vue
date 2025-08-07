<script setup lang="ts">
import type { Paging } from '~/types/api/common'

const props = defineProps<{
  paging: Paging,
}>()

const emit = defineEmits<{
  (e: 'changeCursor', value: string | undefined): void,
}>()

// const currentOffset = computed<number>(() =>
//   typeof props.cursor === 'number' ? props.cursor : 0,
// )

// const data = computed(() => {
//   if (!props.paging) {
//     return { mode: 'waiting' }
//   }
//   if (props.paging.total_count === undefined) {
//     return {
//       mode: 'cursor',
//       next_cursor: props.paging.next_cursor,
//       prev_cursor: props.paging.prev_cursor,
//     }
//   }
//   const page
//     = props.paging.total_count > 0
//       ? 1 + Math.floor(currentOffset.value / props.pageSize)
//       : 0
//   const from = props.paging.total_count > 0 ? currentOffset.value + 1 : 0
//   const to = Math.min(
//     currentOffset.value + props.pageSize,
//     props.paging.total_count,
//   )
//   const lastPage = Math.ceil(props.paging.total_count / props.pageSize)

//   return {
//     from,
//     lastPage,
//     mode: 'offset',
//     page,
//     to,
//   }
// })

// const next = () => {
//   emit(
//     'setCursor',
//     Math.min(
//       currentOffset.value + props.pageSize,
//       ((data.value.lastPage ?? 1) - 1) * props.pageSize,
//     ),
//   )
// }

// const setPageSize = (size: number) => {
//   if (data.value.mode === 'offset') {
//     // in case we increase the page size we must adjust the offset
//     const off = currentOffset.value % size
//     if (off > 0) {
//       emit('setCursor', currentOffset.value - off)
//     }
//   }
//   emit('setPageSize', size)
// }

const limit = defineModel<Query['limit']>('limit')
</script>

<template>
  <div class="bc-pageinator">
    <div class="pager">
      <!-- <template v-if="data.mode === 'offset'">
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
          <BcIcon
            name="chevron-left"
            class="toggle"
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
          <BcIcon
            name="chevron-right"
            class="toggle"
          />
        </div>
        <div
          class="item button"
          :disabled="data.page! >= data.lastPage!"
          @click="last"
        >
          {{ $t("table.last") }}
        </div>
      </template> -->
      <!-- <template v-else-if="data.mode === 'cursor'"> -->
      <BcButtonIcon
        name="chevrons-left"
        screenreader-text="table.navigation.first"
        class="item button"
        :disabled="!props.paging?.prev_cursor"
        @click="emit('changeCursor', undefined)"
      />
      <BcButtonIcon
        screenreader-text="table.navigation.previous"
        name="chevron-left"
        class="item button"
        :disabled="!props.paging?.prev_cursor"
        @click="emit('changeCursor', props.paging?.prev_cursor ?? '')"
      />

      <BcButtonIcon
        screenreader-text="table.navigation.next"
        name="chevron-right"
        class="item button"
        :disabled="!props.paging?.next_cursor"
        @click="emit('changeCursor', props.paging?.next_cursor ?? '')"
      />
      <!-- </template> -->
      <!-- <Select
        v-model="limit"
        :options="[...limits]"
        class="table small"
        @update:model-value="emit('setLimit', $event)"
      /> -->
      <Select
        v-model="limit"
        :options="[...limits]"
        class="table small"
      />
    </div>
    <div
      class="left-info"
    >
      <slot name="bc-table-footer-left">
        <!-- <span v-if="props.paging?.total_count">
          {{
            $t("table.showing", {
              from: data.from,
              to: data.to,
              total: props.paging?.total_count,
            })
          }}
        </span> -->
      </slot>
    </div>
    <div
      v-if="$slots['bc-table-footer-right']"
      class="right-info"
    >
      <slot name="bc-table-footer-right" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";

.bc-pageinator {
  margin-block-start: 20px;
  display: grid;
  gap: var(--padding-small);
  justify-items: center;
  font-weight: var(--standard_text_medium_font_weight);

  @media screen and (min-width: 768px) {
      grid-template-rows: auto 1fr;
      grid-template-columns: 1fr 1fr;

      .pager {
        grid-column: span 2
      }
      .left-info {
        margin-inline-start: 16px;
        justify-self: start;
      }
      .right-info{
        justify-self: end;
        margin-inline-end: 16px;
      }
    }
  @media screen and (min-width: 1024px) {
    grid-template-rows: 1fr ;
    grid-template-columns: 1fr auto 1fr;
    .pager {
      grid-column:  2/3;
      grid-row: 1/2
    }
    .left-info {
      grid-column: 1/2;
      grid-row: 1/2
    }
    .right-info {
      grid-column: 3/4;
      grid-row: 1/2
    }

  }

  .pager {
    display: flex;
    align-items: center;
    gap: var(--padding-small, 3px);

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
        margin: 0;
        &[disabled] {
          color: var(--text-color-disabled);
          cursor: not-allowed;
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

  .left-info {
    display: none;
    @media screen and (min-width: 640px) {
      display: block;
    }
  }
  .right-info {
    font-weight: 400;
    font-size: 12px;
    @media screen and (min-width: 640px) {
      font-weight: inherit;
      font-size: inherit;
    }
  }
}
</style>
