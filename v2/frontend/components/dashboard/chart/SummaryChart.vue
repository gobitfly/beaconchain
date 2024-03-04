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

use([
  CanvasRenderer,
  LineChart,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  GridComponent
])

interface Props {
  dashboardId : number
}
const props = defineProps<Props>()

const store = useValidatorDashboardSummaryChartStore()
const { getDashboardSummaryChart } = store
const { chartData } = storeToRefs(store)

watch(props, () => {
  getDashboardSummaryChart(props.dashboardId)
}, { immediate: true })

const { overview } = storeToRefs(useValidatorDashboardOverview())

const { t: $t } = useI18n()
const colorMode = useColorMode()

const colors = computed(() => {
  return {
    groups: getSummaryChartGroupColors(colorMode.value),
    label: getSummaryChartTextColor(colorMode.value)
  }
})

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
          color: colors.value.label
        }
      }
    },
    series,
    textStyle: {
      fontFamily,
      fontSize: textSize,
      fontWeight: fontWeightLight,
      color: colors.value.label
    },
    color: colors.value.groups,
    legend: {
      type: 'scroll',
      orient: 'horizontal',
      bottom: 65,
      textStyle: {
        color: colors.value.label,
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
          color: colors.value.label
        },
        areaStyle: {
          color: colors.value.label
        }
      },
      borderColor: colors.value.label
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
