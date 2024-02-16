import type { EmitFlags } from 'typescript';
import type Column from 'primevue/column';

import type { Column } from '#build/components';
<script setup lang="ts">
interface Props {
  currentOffset: number
  pageSize: number
  totalCount: number
}
const props = defineProps<Props>()

const emit = defineEmits<{(e: 'setOffset', value: number): void }>()

const current = computed(() => {
  const page = 1 + Math.floor(props.currentOffset / props.pageSize)
  const from = props.currentOffset
  const to = props.currentOffset + props.pageSize

  return { page, from, to }
})

const lastPage = computed(() => Math.ceil(props.totalCount / props.pageSize))

const next = () => {
  emit('setOffset', props.currentOffset + props.pageSize)
}

const prev = () => {
  emit('setOffset', props.currentOffset - props.pageSize)
}

const first = () => {
  emit('setOffset', 0)
}

const last = () => {
  emit('setOffset', (lastPage.value - 1) * props.pageSize)
}

</script>
<template>
  <div class="bc-pageinator">
    <div v-if="props.totalCount" class="pager">
      <div class="item button" :disabled="!props.currentOffset" @click="first">
        {{ $t('table.first') }}
      </div>
      <div class="item button" :disabled="!props.currentOffset" @click="prev">
        <IconChevron class="toggle" direction="left" />
      </div>
      <div class="item">
        {{ current.page }} {{ $t('table.of') }} {{ lastPage }}
      </div>
      <div class="item button" :disabled="current.page === lastPage" @click="next">
        <IconChevron class="toggle" direction="right" />
      </div>
      <div class="item button" :disabled="current.page === lastPage" @click="last">
        {{ $t('table.last') }}
      </div>
    </div>
    <div v-if="props.totalCount" class="left-info">
      {{ $t('table.showing', { from: current.from, to: current.to, total: props.totalCount }) }}
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

    .item {
      @include main.container;
      display: flex;
      justify-content: center;
      align-items: center;
      height: 30px;
      padding: 0 15px;
      border-radius: 0;

      &.button {
        &:not([disabled="true"]) {
          cursor: pointer;
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
        border-top-righ-radius: var(--border-radius);
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
