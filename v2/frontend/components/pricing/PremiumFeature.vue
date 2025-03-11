<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faArrowUpRightFromSquare } from '@fortawesome/pro-solid-svg-icons'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { Feature } from '~/types/pricing'

interface Props {
  feature: Feature,
}
defineProps<Props>()
</script>

<template>
  <div class="feature-container">
    <div class="main-row">
      <BcFeatureCheck :available="feature.available" />
      <div
        class="description"
        :class="{ unavailable: !feature.available }"
      >
        <div class="name">
          {{ feature.name }}
          <BcTooltip
            v-if="feature.tooltip"
            position="bottom"
            class="tooltip"
            fit-content
          >
            <FontAwesomeIcon :icon="faInfoCircle" />
            <template #tooltip>
              <div class="tooltip-content">
                {{ feature.tooltip }}
              </div>
            </template>
          </BcTooltip>
        </div>
        <div
          v-if="feature.subtext || feature.tooltip"
          class="additional-info-row"
        >
          <div v-if="feature.subtext">
            {{ feature.subtext }}
          </div>
        </div>
      </div>
      <BcLink
        v-if="feature.link"
        class="link"
        :to="feature.link"
        target="_blank"
      >
        <FontAwesomeIcon
          class="popout"
          :icon="faArrowUpRightFromSquare"
        />
      </BcLink>
    </div>
    <BcFractionBar
      v-if="feature.percentage"
      :fill-percentage="feature.percentage"
      class="fraction-bar-container"
    />
  </div>
</template>

<style lang="scss" scoped>
.feature-container {
  display: flex;
  flex-direction: column;

  .main-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding-left: 10px;
    margin-bottom: 10px;

    .description {
      display: flex;
      flex-direction: column;
      gap: 5px;
      text-align: left;
      flex-grow: 1;

      &.unavailable {
        color: var(--text-color-discreet);
      }

      .name {
        font-size: 15px;
        display: flex;
        justify-content: space-between;
      }

      .additional-info-row {
        display: flex;
        justify-content: space-between;
        color: var(--text-color-discreet);
        font-size: 12px;
        font-weight: 400;
      }
    }

    .popout {
      width: 14px;
      height: auto;
      margin-left: 7px;
    }
  }

  @media (max-width: 1360px) {
    .main-row {
      margin-bottom: 5px;

      .check {
        width: 12px;
      }

      .description {
        gap: 0;

        .name {
          font-size: 12px;
        }

        .additional-info-row {
          font-size: 10px;
        }
      }
    }
  }

  .fraction-bar-container {
    height: 11px;
  }
}

.tooltip {
  display: flex;
  align-items: center;
}

.tooltip-content {
  width: 130px;
  text-align: left;
}
</style>
