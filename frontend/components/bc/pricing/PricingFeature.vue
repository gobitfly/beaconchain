<script lang="ts" setup>
// TODO: Checkmark is too small

interface Props {
  name: string,
  available: boolean,
  barFillPercentage?: number,
  subtext?: string,
}
defineProps<Props>()

</script>

<template>
  <div class="feature-container">
    <div class="main-row">
      <BcFeatureCheck :available="available" class="check" />
      <div class="text">
        <div class="name">
          {{ name }}
        </div>
        <div v-if="subtext" class="subtext">
          {{ subtext }}
        </div>
      </div>
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
      height: 17px;
    }

    .text {
      display: flex;
      flex-direction: column;
      gap: 5px;
      text-align: left;

      .name {
        font-size: 17px;
      }

      .subtext {
        color: var(--text-color-discreet);
        font-size: 14px;
        font-weight: 300;
      }
    }
  }

  .fraction-bar-container {
    height: 11px;
  }
}
</style>
