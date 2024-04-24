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
import { formatEpochToDate } from '~/utils/format'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { getSummaryChartGroupColors, getSummaryChartTextColor, getSummaryChartTooltipBackgroundColor } from '~/utils/colors'
import { type InternalGetValidatorDashboardSummaryChartResponse } from '~/types/api/validator_dashboard'
import { type ChartData } from '~/types/api/common'
import { getGroupLabel } from '~/utils/dashboard/group'

use([
  CanvasRenderer,
  LineChart,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  GridComponent
])

const { fetch } = useCustomFetch()

const { dashboardKey } = useDashboardKey()

const data = ref<ChartData<number, number> | undefined >()
await useAsyncData('validator_overview', async () => {
  if (!dashboardKey.value) {
    data.value = undefined
    return
  }
  const res = await fetch<InternalGetValidatorDashboardSummaryChartResponse>(API_PATH.DASHBOARD_SUMMARY_CHART, undefined, { dashboardKey: dashboardKey.value })
  data.value = res.data
}, { watch: [dashboardKey], server: false })

const { overview } = useValidatorDashboardOverviewStore()

const { t: $t } = useI18n()
const colorMode = useColorMode()

const colors = computed(() => {
  return {
    groups: getSummaryChartGroupColors(colorMode.value),
    label: getSummaryChartTextColor(colorMode.value),
    background: getSummaryChartTooltipBackgroundColor(colorMode.value)
  }
})

const styles = window.getComputedStyle(document.documentElement)
const fontFamily = styles.getPropertyValue('--roboto-family')
const textSize = parseInt(styles.getPropertyValue('--standard_text_font_size'))
const fontWeightLight = parseInt(styles.getPropertyValue('--roboto-light'))
const fontWeightMedium = parseInt(styles.getPropertyValue('--roboto-medium'))

const option = computed(() => {
  if (data === undefined) {
    return undefined
  }

  interface SeriesObject {
    data: number[];
    type: string;
    name: string;
  }

  const series: SeriesObject[] = []
  if (data.value?.series) {
    const allGroups = $t('dashboard.validator.summary.chart.all_groups')
    data.value.series.forEach((element) => {
      const name = getGroupLabel($t, element.id, overview.value?.groups, allGroups)
      const newObj: SeriesObject = {
        data: element.data,
        type: 'line',
        name
      }
      series.push(newObj)
    })
  }

  return {
    grid: {
      containLabel: true,
      top: 10,
      left: '5%',
      right: '5%'
    },
    xAxis: {
      type: 'category',
      data: data.value?.categories,
      boundaryGap: false,
      axisLabel: {
        fontSize: textSize,
        lineHeight: 20,
        formatter: (value: number) => {
          const date = formatEpochToDate(value, $t('locales.date'))
          if (date === undefined) {
            return ''
          }

          return `${date}\n${$t('common.epoch')} ${value}`
        }
      }
    },
    yAxis: {
      name: $t('dashboard.validator.summary.chart.efficiency'),
      nameLocation: 'center',
      nameTextStyle: {
        padding: [0, 0, 30, 0]
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
      bottom: 40,
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
      borderColor: colors.value.background,
      valueFormatter: (value: number) => {
        return `${value}% ${$t('dashboard.validator.summary.chart.efficiency')}`
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
  <ClientOnly>
    <VChart class="chart" :option="option" autoresize />
  </ClientOnly>
</template>

<style lang="scss">
</style>
