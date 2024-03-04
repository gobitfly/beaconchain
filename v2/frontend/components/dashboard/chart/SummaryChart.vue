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
import { useI18n } from 'vue-i18n'
import SummaryChartTooltip from './SummaryChartTooltip.vue'
import { formatTs } from '~/utils/format'
import { useValidatorDashboardOverview } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { getSummaryChartGroupColors, getSummaryChartTextColor } from '~/utils/colors'

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
const colorMode = useColorMode()

const chartData = ref<ChartData<number> | null>(null)

onMounted(async () => {
  const res = await $fetch<ChartData<number>>('./mock/dashboard/summaryChart.json')
  chartData.value = res
})

const groupColors = ref<string[]>()
const labelColor = ref<string>()

watch(colorMode, (newColorMode) => {
  groupColors.value = getSummaryChartGroupColors(newColorMode.value)
  labelColor.value = getSummaryChartTextColor(newColorMode.value)
}, { immediate: true })

const styles = window.getComputedStyle(document.documentElement)
const fontFamily = styles.getPropertyValue('--roboto-family')
const textSize = parseInt(styles.getPropertyValue('--standard_text_font_size'))
const fontWeightLight = parseInt(styles.getPropertyValue('--roboto-light'))
const fontWeightMedium = parseInt(styles.getPropertyValue('--roboto-medium'))

const option = computed(() => {
  interface SeriesObject {
    data: number[];
    type: string;
    name: string;
  }

  const series: SeriesObject[] = []
  if (chartData.value?.series) {
    const allGroups = $t('dashboard.validator.summary.chart.all_groups')
    chartData.value.series.forEach((element) => {
      let name = allGroups
      if (element.id !== -1) {
        const group = overview.value?.groups.find(group => group.id === element.id)
        name = group?.name || element.id.toString()
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
    xAxis: {
      type: 'category',
      data: chartData.value?.categories,
      boundaryGap: false,
      axisLabel: {
        fontSize: textSize,
        lineHeight: 20,
        formatter: (value: number) => {
          const ts = epochToTs(value)
          if (ts === undefined) {
            return ''
          }

          const date = formatTs(ts)
          return `${date}\nEpoch ${value}`
        }
      }
    },
    yAxis: {
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
        fontSize: textSize
      },
      splitLine: {
        lineStyle: {
          color: labelColor.value
        }
      }
    },
    series,
    textStyle: {
      fontFamily,
      fontSize: textSize,
      fontWeight: fontWeightLight,
      color: labelColor.value
    },
    color: groupColors.value,
    legend: {
      type: 'scroll',
      orient: 'horizontal',
      bottom: 65,
      textStyle: {
        color: labelColor.value,
        fontSize: textSize,
        fontWeight: fontWeightMedium
      }
    },
    tooltip: {
      order: 'seriesAsc',
      trigger: 'axis',
      padding: 0,
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
        render(h(SummaryChartTooltip, { t: $t, startEpoch, groupInfos }), d)
        return d
      }
    },
    dataZoom: {
      type: 'slider',
      start: 80,
      end: 100,
      dataBackground: {
        lineStyle: {
          color: labelColor.value
        },
        areaStyle: {
          color: labelColor.value
        }
      },
      borderColor: labelColor.value
    }
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
