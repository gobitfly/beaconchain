<script setup lang="ts">
import type { Paging } from '~/types/api/common'
import type { Cursor } from '~/types/datatable'

interface Props {
  cursor: Cursor
  pageSize: number
  paging?: Paging
}
const props = defineProps<Props>()

const emit = defineEmits<{(e: 'setCursor', value: Cursor): void, (e: 'setPageSize', value: number): void }>()

const pageSizes = [5, 10, 25, 50, 100]

const currentOffset = computed<number>(() => typeof props.cursor === 'number' ? props.cursor : 0)

const data = computed(() => {
  if (!props.paging) {
    return {
      mode: 'waiting'
    }
  }
  if (props.paging.total_count === undefined) {
    return {
      mode: 'cursor',
      prev_cursor: props.paging.prev_cursor,
      next_cursor: props.paging.next_cursor
    }
  }
  const page = 1 + Math.floor(currentOffset.value / props.pageSize)
  const from = props.paging.total_count > 0 ? currentOffset.value + 1 : 0
  const to = Math.min(currentOffset.value + props.pageSize, props.paging.total_count)
  const lastPage = Math.ceil(props.paging.total_count / props.pageSize)

  return { mode: 'offset', page, from, to, lastPage }
})

const next = () => {
  emit('setCursor', Math.min(currentOffset.value + props.pageSize, ((data.value.lastPage ?? 1) - 1) * props.pageSize))
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
watch(() => data.value.lastPage && data.value.lastPage < data.value.page, (match) => {
  if (data.value.lastPage !== undefined && match) {
    last()
  }
})

</script>
<template>
  <div class="bc-pageinator">
    <template v-if="data.mode === 'offset'">
      <div class="pager">
        <div class="item button" :disabled="!currentOffset" @click="first">
          {{ $t('table.first') }}
        </div>
        <div class="item button" :disabled="!currentOffset" @click="prev">
          <IconChevron class="toggle" direction="left" />
        </div>
        <div class="item">
          {{ data.page }} {{ $t('table.of') }} {{ data.lastPage }}
        </div>
        <div class="item button" :disabled="data.page! >= data.lastPage!" @click="next">
          <IconChevron class="toggle" direction="right" />
        </div>
        <div class="item button" :disabled="data.page! >= data.lastPage!" @click="last">
          {{ $t('table.last') }}
        </div>
        <Dropdown
          :model-value="props.pageSize"
          :options="pageSizes"
          class="table small"
          @change="(event) => setPageSize(event.value)"
        />
      </div>
      <div class="left-info">
        {{ $t('table.showing', { from: data.from, to: data.to, total: props.paging?.total_count }) }}
      </div>
    </template>
    <div v-else-if="data.mode === 'cursor'" class="pager">
      <div class="item button" :disabled="!data.prev_cursor" @click="first">
        {{ $t('table.first') }}
      </div>
      <div class="item button" :disabled="!data.prev_cursor" @click="emit('setCursor', data.prev_cursor)">
        <IconChevron class="toggle" direction="left" />
      </div>
      <div class="item button" :disabled="data.next_cursor" @click="emit('setCursor', data.next_cursor)">
        <IconChevron class="toggle" direction="right" />
      </div>
      <Dropdown
        :model-value="props.pageSize"
        :options="pageSizes"
        class="table small"
        @change="(event) => setPageSize(event.value)"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';

.bc-pageinator {
  position: relative;
  width: 100%;
  height: 78px;
  display: flex;
  justify-content: center;
  align-items: center;
  font-weight: var(--standard_text_bold_font_weight);
  padding: var(--padding);

  .left-info {
    position: absolute;
    left: 0;
    top: 0;
    height: 100%;
    display: flex;
    align-items: center;
    padding-left: var(--padding);
  }

  .pager {
    display: flex;
    gap: 3px;

    .table{
      @include main.container;
      border-top-left-radius: 0;
      border-bottom-left-radius: 0;
    }

    .item {
      @include main.container;
      display: flex;
      justify-content: center;
      align-items: center;
      height: 30px;
      padding: 0 22px;
      border-radius: 0;

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
    }
  }

  @media screen and (max-width: 1399px) {
    flex-direction: column;
    gap: var(--padding);

    .left-info {
      position: relative;
      height: unset;
    }
  }
}
</style>
