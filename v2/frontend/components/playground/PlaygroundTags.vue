<script setup lang="ts">
import {
  faAnchor,
  faTrophy,
} from '@fortawesome/pro-solid-svg-icons'
import type { BlockStatus } from '~/types/block'
import { TagColors } from '~/types/tag'

const blockStati: BlockStatus[] = ['missed', 'orphaned', 'scheduled', 'success']
</script>

<template>
  <h1>Block table status tags</h1>
  <div class="container">
    <div
      v-for="status in blockStati"
      :key="status"
      class="item"
    >
      <BlockTableStatus :status="status" />
      <BlockTableStatus
        :status="status"
        :mobile="true"
      />
    </div>
  </div>
  <h1>Generic tags</h1>
  <div class="container">
    <div
      v-for="color in TagColors"
      :key="color"
      class="item"
    >
      <BcTableTag
        :color="color"
        :label="color"
      />
      <BcTableTag
        :color="color"
        :label="color"
        :icon="faAnchor"
      />
      <BcTableTag
        :color="color"
        :label="color"
        size="compact"
      />
      <BcTableTag
        :color="color"
        :label="color"
        size="compact"
      />
      <BcTableTag
        :color="color"
        :label="color"
        :icon="faAnchor"
        size="compact"
      />
      <BcTableTag
        :color="color"
        :label="color.substring(0, 2)"
        size="circle"
        :tooltip="color"
      />
      <BcTableTag
        :color="color"
        :icon="faAnchor"
        size="circle"
        :tooltip="color"
      />
    </div>
  </div>
  <h1>Custom colors</h1>
  <div class="container">
    <div class="item">
      <BcTableTag
        label="this is gold"
        :icon="faTrophy"
        :style="{ 'background-color': 'gold', 'color': 'silver' }"
        tooltip="we can also set custom colors on our tags!"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.container{
  display: flex;
  flex-wrap: wrap;

  .item{
    flex-grow: 0;
    margin: 10px;
    display: flex;
    align-items: center;
    gap: 5px;
  }
}
</style>
