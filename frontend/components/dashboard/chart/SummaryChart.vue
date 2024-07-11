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
import { getSummaryChartGroupColors, getChartTextColor, getChartTooltipBackgroundColor } from '~/utils/colors'
import { type InternalGetValidatorDashboardSummaryChartResponse } from '~/types/api/validator_dashboard'
import { type ChartData } from '~/types/api/common'
import { getGroupLabel } from '~/utils/dashboard/group'
import { API_PATH } from '~/types/customFetch'
import { SUMMARY_CHART_GROUP_NETWORK_AVERAGE, SUMMARY_CHART_GROUP_TOTAL, type AggregationTimeframe, type SummaryChartFilter } from '~/types/dashboard/summary'

use([
  CanvasRenderer,
  LineChart,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  GridComponent
])

interface Props {
  filter?: SummaryChartFilter
}

const props = defineProps<Props>()

const { fetch } = useCustomFetch()
const { tsToEpoch } = useNetworkStore()
const { dashboardKey } = useDashboardKey()

const data = ref<ChartData<number, number> | undefined >()
const { value: filter, bounce: bounceFilter } = useDebounceValue(props.filter, 1000)
const aggregation = ref<AggregationTimeframe>('hourly')
const isLoading = ref(false)
const loadData = async () => {
  if (!dashboardKey.value) {
    data.value = undefined
    return
  }
  isLoading.value = true
  const requestAggregation = props.filter?.aggregation || 'hourly'
  const res = await fetch<InternalGetValidatorDashboardSummaryChartResponse>(API_PATH.DASHBOARD_SUMMARY_CHART, { query: { group_ids: props.filter?.groupIds.join(','), efficiency_type: props.filter?.efficiency, aggregation: requestAggregation } }, { dashboardKey: dashboardKey.value })
  aggregation.value = requestAggregation

  isLoading.value = false
  data.value = res.data
}

watch([dashboardKey, filter], () => {
  loadData()
}, { immediate: true })

watch(() => props.filter, (filter) => {
  if (!filter) {
    return
  }
  bounceFilter({ ...filter, groupIds: [...filter.groupIds] }, true, true)
}, { immediate: true, deep: true })

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
      let name: string
      if (element.id === SUMMARY_CHART_GROUP_TOTAL) {
        name = $t('dashboard.validator.summary.chart.total')
      } else if (element.id === SUMMARY_CHART_GROUP_NETWORK_AVERAGE) {
        name = $t('dashboard.validator.summary.chart.average')
      } else {
        name = getGroupLabel($t, element.id, groups.value, allGroups)
      }
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
          const date = formatGoTimestamp(value, undefined, 'absolute', 'narrow', $t('locales.date'), false)
          if (aggregation.value === 'epoch') {
            return `${date}\n${$t('common.epoch')} ${tsToEpoch(value)}`
          }
          return date
        }
      }
    },
    yAxis: {
      name: $t(`dashboard.validator.summary.chart.efficiency.${props.filter?.efficiency}`),
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
      formatter (params : any) : HTMLElement {
        const ts = parseInt(params[0].axisValue)
        const groupInfos = params.map((param: any) => {
          return {
            name: param.seriesName,
            efficiency: param.value,
            color: param.color
          }
        })

        const d = document.createElement('div')
        render(h(SummaryChartTooltip, { t: $t, ts, efficiencyType: props.filter?.efficiency || 'all', aggregation: aggregation.value, groupInfos }), d)
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
