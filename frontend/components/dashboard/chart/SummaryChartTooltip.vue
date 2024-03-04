<script lang="ts" setup>
import { type ComposerTranslation } from 'vue-i18n'
import { formatEpochToDate } from '~/utils/format'

interface Props {
  t: ComposerTranslation, // required as dynamically created components via render do not have the proper app context
  startEpoch: number,
  groupInfos: {
    name: string,
    efficiency: number,
    color: string
  }[]
}

const props = defineProps<Props>()

const { epochsPerDay } = useNetwork()

const dateText = computed(() => {
  const date = formatEpochToDate(props.startEpoch, props.t('locales.date'))
  if (date === undefined) {
    return undefined
  }
  return `${date}`
})

const epochText = computed(() => {
  const endEpoch = props.startEpoch + epochsPerDay()
  return `${props.t('common.epoch')} ${props.startEpoch} - ${endEpoch}`
})

</script>

<template>
  <div class="tooltip-container">
    <div>
      {{ dateText }}
    </div>
    <div>
      {{ epochText }}
    </div>
    <div v-for="(entry, index) in props.groupInfos" :key="index" class="line-container">
      <div class="circle" :style="{ 'background-color': entry.color }" />
      <div>
        {{ entry.name }}:
      </div>
      <div class="efficiency">
        {{ entry.efficiency }}%
      </div>
    </div>
  </div>
</template>

<style lang="scss">
@use '~/assets/css/fonts.scss';

.tooltip-container {
  @include fonts.tooltip_text_bold;
  background-color: var(--tooltip-background);
  color: var(--tooltip-text-color);
  line-height: 1.5;
  padding: var(--padding);

  .line-container{
    display: flex;
    align-items: center;
    gap: 3px;

    .circle{
      width: 10px;
      height: 10px;
      border-radius: 50%;
    }

    .efficiency{
      @include fonts.tooltip_text;
    }
  }
}
</style>
