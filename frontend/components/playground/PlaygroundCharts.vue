<script lang="ts" setup>

import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent
} from 'echarts/components'
import VChart from 'vue-echarts'

import { type ChartData } from '~/types/api/common'

use([
  CanvasRenderer,
  LineChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  GridComponent
])

const { t: $t } = useI18n()

const chartData = ref<ChartData<number> | null>(null)

onMounted(async () => {
  const res = await $fetch<ChartData<number>>('./mock/dashboard/summaryChart.json')
  chartData.value = res
})

const title = {
  text: $t('dashboard.validator.summary.chart.title'),
  left: 'center',
  textAlign: 'center',
  textStyle: {
    fontSize: 24,
    fontWeight: 500,
    color: '#f0f0f0'
  }
}

// TODO: retrieve from css?
const textStyle = {
  fontFamily: 'Roboto',
  fontSize: 14,
  fontWeight: 300,
  color: '#f0f0f0'
}

const color = ['#f0f0f0', '#e6194b', '#46f0f0', '#bcf60c', '#4363d8', '#ffe119', '#f032e6', '#3cb44b', '#911eb4', '#f58231', '#87ceeb', '#e6beff', '#40e0d0', '#fabebe', '#aaffc3', '#ffd8b1', '#fffac8', '#daa520', '#dda0dd', '#fa8072', '#d2b48c', '#6b8e23', '#a0522d', '#008080', '#9a6324', '#800000', '#808000', '#000075', '#808080', '#708090', '#ffdb58']

// TODO: Styling
const legend = {
  orient: 'horizontal',
  bottom: 50,
  textStyle: {
    color: '#f0f0f0',
    fontSize: 14,
    fontWeight: 'bold'
  }
}

// TODO: Styling
const tooltip = {
  order: 'valueDesc',
  trigger: 'axis'
}

// TODO: Styling and default values
const dataZoom = {
  type: 'slider',
  start: 80,
  end: 100
}

const xAxis = {
  type: 'category',
  data: chartData.value?.categories,
  boundaryGap: false,

  axisLabel: {
    fontSize: 14, // TODO: Why is this needed? It should use the global textStyle
    lineHeight: 20
  },

  axisLine: {
    lineStyle: {
      color: '#f0f0f0'
    }
  }
}

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
  const series: SeriesObject[] = []
  if (chartData.value?.series) {
    chartData.value.series.forEach((element) => {
      const newObj: SeriesObject = {
        data: element.data,
        type: 'line',
        name: 'Group ' + element.id // TODO: Use cached group names
      }
      series.push(newObj)
    })
  }

  return {
    height: 400,

    title,
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
  <div class="chart-container">
    <VChart class="chart" :option="option" autoresize />
  </div>
</template>

<style lang="scss">
.chart-container {
  background-color: var(--container-background);
  padding-top: var(--padding-large);
  width: 100%;
  height: 625px;
}
</style>
