<script lang="ts" setup>

import { h, render } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import {
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import SummaryChartTooltip from './SummaryChartTooltip.vue'
import { formatTs } from '~/utils/format'
import { useValidatorDashboardOverview } from '~/stores/dashboard/useValidatorDashboardOverviewStore'

import { type ChartData } from '~/types/api/common'

use([
  CanvasRenderer,
  LineChart,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  GridComponent
])

const { overview } = storeToRefs(useValidatorDashboardOverview())

const { t: $t } = useI18n()

const chartData = ref<ChartData<number> | null>(null)

onMounted(async () => {
  const res = await $fetch<ChartData<number>>('./mock/dashboard/summaryChart.json')
  chartData.value = res
})

function xToLabel (x: number): string | undefined {
  const ts = epochToTs(x)
  if (ts === undefined) {
    return undefined
  }

  const date = formatTs(ts)
  return `${date}\nEpoch ${x}`
}

// TODO: retrieve from css?
const textStyle = {
  fontFamily: 'Roboto',
  fontSize: 14,
  fontWeight: 300,
  color: '#f0f0f0'
}

// TODO: Replace with colors coming from designer
const color = ['#f0f0f0', '#e6194b', '#46f0f0', '#bcf60c', '#4363d8', '#ffe119', '#f032e6', '#3cb44b', '#911eb4', '#f58231', '#87ceeb', '#e6beff', '#40e0d0', '#fabebe', '#aaffc3', '#ffd8b1', '#fffac8', '#daa520', '#dda0dd', '#fa8072', '#d2b48c', '#6b8e23', '#a0522d', '#008080', '#9a6324', '#800000', '#808000', '#000075', '#808080', '#708090', '#ffdb58']

const legend = {
  orient: 'horizontal',
  bottom: 50,
  textStyle: {
    color: '#f0f0f0',
    fontSize: 14,
    fontWeight: 500
  }
}

const tooltip = {
  order: 'seriesAsc',
  trigger: 'axis',
  valueFormatter: (value: number) => {
    return `${value}% ${$t('dashboard.validator.summary.chart.yAxis')}`
  },
  formatter (params : any) : HTMLElement {
    const startEpoch = parseInt(params[0].axisValue)
    const groupInfos = params.map((param: any) => {
      return {
        name: param.seriesName,
        efficiency: param.value,
        color: param.color
      }
    })

    const d = document.createElement('div')
    render(h(SummaryChartTooltip, { startEpoch, groupInfos }), d)
    return d
  }
}

const dataZoom = {
  type: 'slider',
  start: 80,
  end: 100
}

// yAxis does not need to be computed as it will always be the same
const yAxis = {
  name: $t('dashboard.validator.summary.chart.yAxis'),
  nameLocation: 'center',
  nameTextStyle: {
    padding: [0, 0, 35, 0]
  },

  type: 'value',
  minInterval: 50,
  silent: true,

  axisLabel: {
    formatter: '{value} %',
    fontSize: 14
  }
}

interface SeriesObject {
  data: number[];
  type: string;
  name: string;
}

const option = computed(() => {
  const xAxis = {
    type: 'category',
    data: chartData.value?.categories,
    boundaryGap: false,

    axisLabel: {
      fontSize: 14, // TODO: Why is this needed? It should use the global textStyle
      lineHeight: 20,
      formatter: (value: number) => {
        return xToLabel(value) || ''
      }
    },

    axisLine: {
      lineStyle: {
        color: '#f0f0f0'
      }
    }
  }

  const series: SeriesObject[] = []
  if (chartData.value?.series) {
    chartData.value.series.forEach((element) => {
      let name = $t('dashboard.validator.summary.chart.all_groups')

      if (element.id !== -1) {
        const group = overview.value?.groups.find(group => group.id === element.id)
        name = group !== undefined ? group.name : 'Group Id ' + element.id
      }

      const newObj: SeriesObject = {
        data: element.data,
        type: 'line',
        name
      }
      series.push(newObj)
    })
  }

  return {
    height: 400,

    textStyle,
    color,
    legend,
    tooltip,
    dataZoom,

    xAxis,
    yAxis,
    series
  }
})
</script>

<template>
  <VChart class="chart" :option="option" autoresize />
</template>

<style lang="scss">
  .chart-container {
    background-color: var(--container-background);
  }
</style>
