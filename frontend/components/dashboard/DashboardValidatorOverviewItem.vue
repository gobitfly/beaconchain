<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { NumberOrString } from '~/types/value'

const props = defineProps<{ infos: { label: string, value: NumberOrString }[], title: string }>()
</script>

<template>
  <div class="box">
    <div class="main">
      <div class="big_text_label">
        {{ props.title }}
      </div>
      <div class="big_text dashbaord-validator-overview-item__value">
        <slot />
      </div>
    </div>
    <div
      class="additional small_text"
    >
      <slot name="additionalInfo" />
    </div>
    <div
      class="info"
    >
      <BcTooltip :fit-content="true">
        <FontAwesomeIcon :icon="faInfoCircle" />
        <template #tooltip>
          <div class="info-label-list">
            <div
              v-for="info in props.infos"
              :key="info.label"
            >
              <span class="bold">{{ info.label }}:</span> {{ info.value }}
            </div>
          </div>
        </template>
      </BcTooltip>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.box {
  display: flex;
  align-items: center;

  .popout {
    flex-shrink: 0;
  }

  .main,
  .additional {
    display: flex;
    flex-direction: column;

    div {
      white-space: nowrap;
      text-wrap: nowrap;
    }

    .dashbaord-validator-overview-item__value {
      display: flex;
      gap: var(--padding);

    }
  }

  .additional {
    margin-left: 8px;

    &:nth-child(2) {
      margin-left: var(--padding);
    }
  }
}

.info-label-list {
  text-align: left;
}

.info {
  margin-left: var(--padding);
}
</style>
