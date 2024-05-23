<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faArrowUpRightFromSquare } from '@fortawesome/pro-solid-svg-icons'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import BcTooltip from '../bc/BcTooltip.vue'
import type { Feature } from '~/types/pricing'

interface Props {
  feature: Feature
}
defineProps<Props>()

</script>

<template>
  <div class="feature-container">
    <div class="main-row">
      <BcFeatureCheck :available="feature.available" class="check" />
      <div class="text" :class="{ 'unavailable': !feature.available }">
        <div class="name">
          {{ feature.name }}
          <BcTooltip v-if="feature.tooltip" position="top" :fit-content="true" :text="feature.tooltip">
            <FontAwesomeIcon :icon="faInfoCircle" />
          </BcTooltip>
        </div>
        <div v-if="feature.subtext" class="subtext">
          {{ feature.subtext }}
        </div>
      </div>
      <NuxtLink v-if="feature.link" class="link" :to="feature.link" target="_blank">
        <FontAwesomeIcon
          class="popout"
          :icon="faArrowUpRightFromSquare"
        />
      </NuxtLink>
    </div>
    <BcFractionBar v-if="feature.percentage" :fill-percentage="feature.percentage" class="fraction-bar-container" />
  </div>
</template>

<style lang="scss" scoped>
.feature-container {
  display: flex;
  flex-direction: column;

  .main-row {
    display: flex;
    align-items: center;
    gap: 7px;
    padding-left: 14px;
    margin-bottom: 14px;

    .check {
      width: 22px;
      height: auto;
    }

    .text {
      display: flex;
      flex-direction: column;
      gap: 5px;
      text-align: left;

      &.unavailable {
        color: var(--text-color-discreet);
      }

      .name {
        font-size: 17px;

        .slot_container {
          margin-left: 8px;
        }
      }

      .subtext {
        color: var(--text-color-discreet);
        font-size: 14px;
        font-weight: 300;
      }
    }

    .popout {
      width: 14px;
      height: auto;
      margin-left: 7px;
    }
  }

  @media (max-width: 600px) {
    .main-row {
      margin-bottom: 5px;

      .check {
        width: 12px;
      }

      .text {
        gap: 0;

        .name {
          font-size: 10px;
        }

        .subtext {
          font-size: 8px;
        }
      }
    }
  }

  .fraction-bar-container {
    height: 11px;
  }
}
</style>
