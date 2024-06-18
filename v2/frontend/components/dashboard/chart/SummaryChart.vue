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
import { min } from 'lodash-es'
import SummaryChartTooltip from './SummaryChartTooltip.vue'
import { useFormat } from '~/composables/useFormat'
import { getSummaryChartGroupColors, getChartTextColor, getChartTooltipBackgroundColor } from '~/utils/colors'
import { type InternalGetValidatorDashboardSummaryChartResponse } from '~/types/api/validator_dashboard'
import { type ChartData } from '~/types/api/common'
import { getGroupLabel } from '~/utils/dashboard/group'
import { API_PATH } from '~/types/customFetch'

use([
  CanvasRenderer,
  LineChart,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  GridComponent
])

const { fetch } = useCustomFetch()
const { formatEpochToDate } = useFormat()
const { dashboardKey } = useDashboardKey()

const data = ref<ChartData<number, number> | undefined >()
const isLoading = ref(false)
await useAsyncData('validator_dashboard_summary_chart', async () => {
  if (!dashboardKey.value) {
    data.value = undefined
    return
  }
  isLoading.value = true
  const res = await fetch<InternalGetValidatorDashboardSummaryChartResponse>(API_PATH.DASHBOARD_SUMMARY_CHART, undefined, { dashboardKey: dashboardKey.value })

  isLoading.value = false
  data.value = res.data
}, { watch: [dashboardKey], server: false })

const { groups } = useValidatorDashboardGroups()

const { t: $t } = useI18n()
const colorMode = useColorMode()

const colors = computed(() => {
  return {
    groups: getSummaryChartGroupColors(colorMode.value),
    label: getChartTextColor(colorMode.value),
    background: getChartTooltipBackgroundColor(colorMode.value)
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
    smooth: boolean;
    symbol: string,
    name: string;
  }

  const series: SeriesObject[] = []
  if (data.value?.series) {
    const allGroups = $t('dashboard.validator.summary.chart.all_groups')
    data.value.series.forEach((element) => {
      const name = getGroupLabel($t, element.id, groups.value, allGroups)
      const newObj: SeriesObject = {
        data: element.data,
        type: 'line',
        smooth: true,
        symbol: 'none',
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
      minInterval: 10,
      maxInterval: 20,
      min: (range: any) => Math.max(0, 10 * Math.ceil(range.min / 10 - 1)),
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
    <BcLoadingSpinner v-if="isLoading" :loading="true" alignment="center" />
    <VChart v-else class="chart" :option="option" autoresize />
  </ClientOnly>
</template>

<style lang="scss">
</style>
