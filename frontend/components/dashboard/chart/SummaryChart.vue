<script lang="ts" setup>

import { h, render } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { type ECharts } from 'echarts'
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
const chart = ref<ECharts | undefined>()

const { fetch } = useCustomFetch()
const { tsToEpoch, slotToTs, epochToTs } = useNetworkStore()
const { dashboardKey } = useDashboardKey()
const { overview } = useValidatorDashboardOverviewStore()
const { latestState } = useLatestStateStore()
const latestSlot = ref(latestState.value?.current_slot || 0)
const dataZoom = ref<{start:number, end:number}>({ start: 80, end: 100 })

const { value: filter, bounce: bounceFilter } = useDebounceValue(props.filter, 1000)
const aggregation = ref<AggregationTimeframe>('hourly')
const isLoading = ref(false)
const loadData = async () => {
  if (!dashboardKey.value) {
    return
  }
  isLoading.value = true

  interface SeriesObject {
    data: number[];
    type: string;
    smooth: boolean;
    symbol: string,
    name: string;
  }
  const series: SeriesObject[] = []
  try {
    const res = await fetch<InternalGetValidatorDashboardSummaryChartResponse>(API_PATH.DASHBOARD_SUMMARY_CHART, { query: { group_ids: props.filter?.groupIds.join(','), efficiency_type: props.filter?.efficiency, aggregation: aggregation.value } }, { dashboardKey: dashboardKey.value })
    if (res.data) {
      const allGroups = $t('dashboard.validator.summary.chart.all_groups')
      res.data.series.forEach((element) => {
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
          smooth: false,
          symbol: 'none',
          name
        }
        series.push(newObj)
      })
    }
  } catch (e) {
    // TODO: Maybe we want to show an error here (either a toast or inline centred in the chart space)
  }
  isLoading.value = false
  chart.value?.setOption({ series })
}

watch([() => props.filter?.efficiency, () => props.filter?.groupIds], () => {
  if (!props.filter?.initialised || !props.filter?.efficiency) {
    return
  }
  bounceFilter({ ...props.filter, groupIds: [...props.filter.groupIds] }, true, true)
}, { immediate: true, deep: true })

watch(() => props.filter?.aggregation, (agg) => {
  if (!agg) {
    return
  }
  latestSlot.value = latestState.value?.current_slot || 0
  aggregation.value = agg
}, { immediate: true })

watch([dashboardKey, filter], () => {
  loadData()
}, { immediate: true })

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
let lastMouseYPos = 0

const formatTSToDate = (value: string) => {
  return formatGoTimestamp(Number(value), undefined, 'absolute', 'narrow', $t('locales.date'), false)
}
const formatTSToEpoch = (value: string) => {
  return `${$t('common.epoch')} ${tsToEpoch(Number(value))}`
}
const formatToDateOrEpoch = (value: string) => {
  if (aggregation.value === 'epoch') {
    return formatTSToEpoch(value)
  }
  return formatTSToDate(value)
}

const formatTimestamp = (value: string) => {
  const date = formatTSToDate(value)
  if (aggregation.value === 'epoch') {
    return `${date}\n${formatTSToEpoch(value)}`
  }
  return date
}

const categories = computed<number[]>(() => {
  // charts have 5 slots delay
  if (latestSlot.value <= 5 || !aggregation.value) {
    return []
  }
  const maxSeconds = overview.value?.chart_history_seconds?.[aggregation.value] ?? 0
  if (!maxSeconds) {
    return []
  }
  const list: number[] = []
  let latestTs = slotToTs(latestSlot.value - 5) || 0
  let step = 0
  switch (aggregation.value) {
    case 'epoch':
      step = epochToTs(1) || 0
      break
    case 'daily':
      step = ONE_DAY
      break
    case 'hourly':
      step = ONE_HOUR
      break
    case 'weekly':
      step = ONE_WEEK
      break
  }
  if (!step) {
    return []
  }
  const minTS = Math.max(slotToTs(0) || 0, latestTs - maxSeconds)
  while (latestTs >= minTS) {
    list.splice(0, 0, latestTs)

    latestTs -= step
  }
  console.log('list', list)

  return list
})

const option = computed(() => {
  return {
    grid: {
      containLabel: true,
      top: 10,
      left: '5%',
      right: '5%'
    },
    xAxis: {
      type: 'category',
      data: categories.value,
      boundaryGap: false,
      axisLabel: {
        fontSize: textSize,
        lineHeight: 20,
        formatter: formatTimestamp
      }
    },
    series: [],
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
      formatter (params: any): HTMLElement {
        const ts = parseInt(params[0].axisValue)
        let lastDif = 0
        let highlightGroup = ''
        const groupInfos = params.map((param: any) => {
          if (chart.value) {
            const distance = Math.abs(lastMouseYPos - chart.value.convertToPixel({ yAxisIndex: 0 }, param.value))
            if (distance < lastDif || !highlightGroup) {
              lastDif = distance
              highlightGroup = param.seriesName
            }
          }
          return {
            name: param.seriesName,
            efficiency: param.value,
            color: param.color
          }
        })
        const d = document.createElement('div')
        render(h(SummaryChartTooltip, { t: $t, ts, efficiencyType: props.filter?.efficiency || 'all', aggregation: aggregation.value, groupInfos, highlightGroup }), d)
        return d
      }
    },
    dataZoom: {
      type: 'slider',
      ...dataZoom.value,
      labelFormatter: (_value: number, valueStr: string) => {
        return formatToDateOrEpoch(valueStr)
      },
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

/* watch(option, (o) => {
  chart.value?.setOption(o)
}) */

const onMouseMove = (e: MouseEvent) => {
  lastMouseYPos = e.offsetY
}

const onDatazoom = (e: any) => {
  console.log('onDataZoom', e)
}

</script>

<template>
  <div class="summary-chart-container" @mousemove="onMouseMove">
    <ClientOnly>
      <VChart ref="chart" class="chart" :option="option" autoresize @datazoom="onDatazoom" />
      <BcLoadingSpinner v-if="isLoading" class="loading-spinner" :loading="true" alignment="center" />
    </ClientOnly>
  </div>
</template>

<style lang="scss" scoped>
.summary-chart-container {
  position: relative;
  height: 100%;

  .loading-spinner {
    position: absolute;
    top: 0;
    left: 0;
  }
}
</style>
