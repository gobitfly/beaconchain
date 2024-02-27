<script lang="ts" setup>

import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart from 'vue-echarts'

use([
  CanvasRenderer,
  LineChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

const dataSetAll = [85, 85, 90, 67.5, 80, 45, 95]
const dataSetHetzner = [80, 75, 85, 85, 80, 0, 90]
const dataSetOVH = [90, 95, 95, 50, 80, 90, 100]

const labels = ['22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393', '22.Aug\nEpoch 141393']

const option = ref({
  xAxis: {
    data: labels
  },
  yAxis: {},
  series: [
    {
      data: dataSetAll,
      type: 'line'
    },
    {
      data: dataSetHetzner,
      type: 'line'
    },
    {
      data: dataSetOVH,
      type: 'line'
    }
  ]
})
</script>

<template>
  <div class="chart-container">
    <VChart class="chart" :option="option" />
  </div>
</template>

<style lang="scss">
.chart-container {
  height: 640px;
}
</style>
