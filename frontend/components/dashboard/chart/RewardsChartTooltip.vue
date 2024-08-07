<script lang="ts" setup>
import { type ComposerTranslation } from 'vue-i18n'
import { BigNumber } from '@ethersproject/bignumber'
import type { RewardChartGroupData, RewardChartSeries } from '~/types/dashboard/rewards'
import type { WeiToValue } from '~/types/value'

interface Props {
  t: ComposerTranslation // required as dynamically created components via render do not have the proper app context,
  weiToValue: WeiToValue
  startEpoch: number
  dataIndex: number
  series: RewardChartSeries[]
}

const props = defineProps<Props>()

interface GroupValue {
  id: number
  name: string
  value: string
}

interface Series {
  name: string
  value: string
  className?: string
  groups: GroupValue[]
}

const mapData = (groups: RewardChartGroupData[]): GroupValue[] => {
  const sort = [...groups].sort((g1, g2) => {
    const v1 = g1.bigData[props.dataIndex] || BigNumber.from('0')
    const v2 = g2.bigData[props.dataIndex] || BigNumber.from('0')
    return v1.gt(v2) ? -1 : 1
  })
  return sort.map(g => ({
    name: g.name,
    id: g.id,
    value: `${props.weiToValue(g.bigData[props.dataIndex]).label}`,
  }))
}

const data = computed<Series[]>(() => {
  const el: Series = {
    name: props.series[1].name,
    className: 'cl',
    value: props.series[1].formatedData[props.dataIndex].label as string,
    groups: mapData(props.series[1].groups),
  }
  const cl: Series = {
    name: props.series[0].name,
    className: 'el',
    value: props.series[0].formatedData[props.dataIndex].label as string,
    groups: mapData(props.series[0].groups),
  }

  const totalGroups = props.series[0].groups.map((g) => {
    const elValue = props.series[1].groups.find(elG => elG.id === g.id)?.bigData?.[props.dataIndex] ?? BigNumber.from(0)
    const bigData = [...g.bigData]
    bigData[props.dataIndex] = bigData[props.dataIndex].add(elValue)
    return {
      ...g,
      bigData,
    }
  })
  props.series[1].groups.forEach((g) => {
    if (!totalGroups.find(tG => tG.id === g.id)) {
      totalGroups.push(g)
    }
  })

  const total: Series = {
    name: props.t('dashboard.validator.rewards.chart.total'),
    value: `${props.weiToValue(props.series[1].bigData[props.dataIndex].add(props.series[0].bigData[props.dataIndex])).label}`,
    groups: mapData(totalGroups),
  }
  return [el, cl, total]
})
</script>

<template>
  <div class="tooltip-container">
    <DashboardChartTooltipHeader
      :t="t"
      :start-epoch="startEpoch"
    />
    <div
      v-for="(entry, index) in data"
      :key="index"
    >
      <div class="header">
        <span
          class="circle"
          :class="entry.className"
        /><b>{{ entry.name }}: {{ entry.value }}</b>
      </div>
      <ol>
        <li
          v-for="group in entry.groups"
          :key="group.id"
        >
          {{ group.name }}: {{ group.value }}
        </li>
      </ol>
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

  .header {
    display: flex;
    align-items: center;
    margin-top: var(--padding);
    gap: 3px;

    .circle {
      width: 9px;
      height: 9px;
      border-radius: 50%;
      margin-bottom: 2px;

      &.el {
        background-color: var(--primary-orange);
      }

      &.cl {
        background-color: var(--melllow-blue);
      }
    }
  }

  ol {
    margin-block-start: 0;
    margin-block-end: 0;
    margin-inline-start: 0px;
    margin-inline-end: 0px;
    padding-inline-start: 26px;
  }
}
</style>
