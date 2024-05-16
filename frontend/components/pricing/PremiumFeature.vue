<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faArrowUpRightFromSquare } from '@fortawesome/pro-solid-svg-icons'

interface Props {
  name: string,
  available?: boolean,
  barFillPercentage?: number,
  subtext?: string,
  link?: string
}
defineProps<Props>()

</script>

<template>
  <div class="feature-container">
    <div class="main-row">
      <BcFeatureCheck :available="available" class="check" />
      <div class="text" :class="{ 'unavailable': !available }">
        <div class="name">
          {{ name }}
        </div>
        <div v-if="subtext" class="subtext">
          {{ subtext }}
        </div>
      </div>
      <NuxtLink v-if="link" class="link" :to="link" target="_blank">
        <FontAwesomeIcon
          class="popout"
          :icon="faArrowUpRightFromSquare"
        />
      </NuxtLink>
    </div>
    <BcFractionBar v-if="barFillPercentage" :bar-fill-percentage="barFillPercentage" class="fraction-bar-container" />
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

  .fraction-bar-container {
    height: 11px;
  }
}
</style>
