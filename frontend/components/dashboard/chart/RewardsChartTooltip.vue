<script lang="ts" setup>
import { type ComposerTranslation } from 'vue-i18n'
import { BigNumber } from '@ethersproject/bignumber'
import type { RewardChartGroupGroupData, RewardChartSeries } from '~/types/dashboard/rewards'
import { formatEpochToDate } from '~/utils/format'
import type { VaiToValue } from '~/types/value'

interface Props {
  t: ComposerTranslation, // required as dynamically created components via render do not have the proper app context,
  weiToValue: VaiToValue,
  startEpoch: number,
  dataIndex: number,
  series: RewardChartSeries[]
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

interface GroupValue {
  id: number,
  name: string,
  value: string
}

interface Group {
  name: string,
  value: string
  groups: GroupValue[]
}

const mapData = (groups: RewardChartGroupGroupData[]):GroupValue[] => {
  const sort = [...groups].sort((g1, g2) => {
    const v1 = g1.bigData[props.dataIndex] || BigNumber.from('0')
    const v2 = g2.bigData[props.dataIndex] || BigNumber.from('0')
    return v1.gt(v2) ? -1 : 1
  })
  return sort.map(g => ({
    name: g.name,
    id: g.id,
    value: `${props.weiToValue(g.bigData[props.dataIndex]).label}`
  }))
}

const data = computed<Group[]>(() => {
  const el:Group = {
    name: props.series[1].name,
    value: props.series[1].formatedData[props.dataIndex].label as string,
    groups: mapData(props.series[1].groups)
  }
  const cl:Group = {
    name: props.series[0].name,
    value: props.series[0].formatedData[props.dataIndex].label as string,
    groups: mapData(props.series[0].groups)
  }
  return [el, cl]
})

</script>

<template>
  <div class="tooltip-container" @click.stop.prevent="console.log('click')">
    <div>
      {{ dateText }}
    </div>
    <div>
      {{ epochText }}
    </div>
    <div v-for="(entry, index) in data" :key="index">
      <div class="line-container">
        {{ entry.name }}: {{ entry.value }}
      </div>
      <div v-for="group in entry.groups" :key="group.id">
        <div class="line-container">
          {{ group.name }}: {{ group.value }}
        </div>
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
  max-height: 400px;
  overflow-y: auto;
  pointer-events: all;

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
