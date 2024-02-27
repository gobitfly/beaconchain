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

// TODO: Extend data to way more data points
const dataSetAll = [85, 85, 90, 67.5, 80, 45, 95]
const dataSetHetzner = [80, 75, 85, 85, 80, 0, 90]
const dataSetOVH = [90, 95, 95, 50, 80, 90, 100]

// TODO: Replace with numeric values and provide a simple formatter
const labels = ['22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393']

const option = ref({
  height: 400,

  textStyle: {
    fontFamily: 'Roboto',
    fontSize: 14,
    fontWeight: 300,
    color: '#f0f0f0'
  },

  color: ['#f0f0f0', '#e6194b', '#46f0f0', '#bcf60c', '#4363d8', '#ffe119', '#f032e6', '#3cb44b', '#911eb4', '#f58231', '#87ceeb', '#e6beff', '#40e0d0', '#fabebe', '#aaffc3', '#ffd8b1', '#fffac8', '#daa520', '#dda0dd', '#fa8072', '#d2b48c', '#6b8e23', '#a0522d', '#008080', '#9a6324', '#800000', '#808000', '#000075', '#808080', '#708090', '#ffdb58'],

  title: {
    text: $t('dashboard.validator.summary.chart.title'),
    left: 'center',
    textAlign: 'center',
    textStyle: {
      fontSize: 24,
      fontWeight: 500,
      color: '#f0f0f0'
    }
  },

  xAxis: {
    type: 'category',
    data: labels,
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
  },

  yAxis: {
    name: $t('dashboard.validator.summary.chart.yAxis'),
    nameLocation: 'center',
    nameTextStyle: { // TODO: retrieve from css?
      padding: [0, 0, 35, 0]
    },

    type: 'value',
    minInterval: 50,
    silent: true,

    axisLabel: {
      formatter: '{value} %',
      fontSize: 14
    }
  },

  // TODO: Styling
  legend: {
    orient: 'horizontal',
    bottom: 50,
    textStyle: {
      color: '#f0f0f0',
      fontSize: 14,
      fontWeight: 'bold'
    }
  },

  // TODO: Styling
  tooltip: {
    order: 'valueDesc',
    trigger: 'axis'
  },

  // TODO: Styling
  dataZoom: [
    {
      type: 'inside',
      start: 0,
      end: 20
    },
    {
      start: 0,
      end: 20
    }
  ],

  series: [
    {
      data: dataSetAll,
      type: 'line',
      name: 'All Groups' // TODO: Translation
    },
    {
      data: dataSetHetzner,
      type: 'line',
      name: 'Hetzner' // TODO: Use cached group names on dashboard
    },
    {
      data: dataSetOVH,
      type: 'line',
      name: 'OVH' // TODO: Use cached group names on dashboard
    }
  ]
})
</script>

<template>
  <div class="chart-container">
    <!-- TODO: Somehow this is not reactive yet -->
    <VChart class="chart" :option="option" />
  </div>
</template>

<style lang="scss">
.chart-container {
  padding-top: var(--padding-large);
  width: 100%;
  height: 625px;
}
</style>
