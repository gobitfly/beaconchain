<script lang="ts" setup>

import { h, render } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { type ECharts } from 'echarts'
import { get } from 'lodash-es'
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

const { t: $t } = useI18n()
const colorMode = useColorMode()
const { fetch } = useCustomFetch()
const { tsToEpoch, slotToTs, secondsPerEpoch } = useNetworkStore()
const { dashboardKey } = useDashboardKey()
const { overview } = useValidatorDashboardOverviewStore()
const { groups } = useValidatorDashboardGroups()
const { latestState } = useLatestStateStore()
const latestSlot = ref(latestState.value?.current_slot || 0)
const { value: timeFrames, temp: tempTimeFrames, bounce: bounceTimeFrames, instant: instantTimeFrames } = useDebounceValue<{from:number, to:number}>({ from: 0, to: 0 }, 1000)
const currentZoom = { start: 80, end: 100 }
const MAX_DATA_POINTS = 10

const { value: filter, bounce: bounceFilter } = useDebounceValue(props.filter, 1000)
const aggregation = ref<AggregationTimeframe>('hourly')
const isLoading = ref(false)

interface SeriesObject {
    data: number[];
    type: string;
    smooth: boolean;
    symbol: string,
    name: string;
  }
// we don't want the series to be responsive to not trigger an auto update of the option computed
let series: (SeriesObject | undefined)[] = []

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
      step = secondsPerEpoch()
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
  const minTs = Math.max(slotToTs(0) || 0, latestTs - maxSeconds)
  while (latestTs >= minTs) {
    list.splice(0, 0, latestTs)

    latestTs -= step
  }

  return list
})

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

const loadData = async () => {
  series = []
  let liveCategories: number[] = []
  if (!dashboardKey.value || !timeFrames.value.to) {
    chart.value?.setOption({ series })
    return
  }
  isLoading.value = true
  try {
    const res = await fetch<InternalGetValidatorDashboardSummaryChartResponse>(API_PATH.DASHBOARD_SUMMARY_CHART, { query: { after_ts: timeFrames.value.from, before_ts: timeFrames.value.to, group_ids: props.filter?.groupIds.join(','), efficiency_type: props.filter?.efficiency, aggregation: aggregation.value } }, { dashboardKey: dashboardKey.value })
    if (res.data) {
      liveCategories = res.data.categories
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
      series.push()
    }
  } catch (e) {
    // TODO: Maybe we want to show an error here (either a toast or inline centred in the chart space)
  }
  isLoading.value = false
  const axis0 = get(chart.value, 'xAxis[0]') || {}
  const axis1 = get(chart.value, 'xAxis[1]')
  const xAxis = [{ ...axis0, data: liveCategories }, axis1]
  chart.value?.setOption({ series, xAxis })
}

watch([dashboardKey, filter, aggregation, timeFrames], () => {
  loadData()
}, { immediate: true })

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

const option = computed(() => {
  return {
    grid: {
      containLabel: true,
      top: 10,
      left: '5%',
      right: '5%'
    },
    xAxis: [
      {
        type: 'category',
        data: categories.value,
        boundaryGap: false,
        axisLabel: {
          fontSize: textSize,
          lineHeight: 20,
          formatter: formatTimestamp
        }
      },
      {
        type: 'category',
        data: categories.value,
        show: false,
        boundaryGap: false
      }
    ],
    series,
    yAxis: {
      name: $t(`dashboard.validator.summary.chart.efficiency.${props.filter?.efficiency}`),
      nameLocation: 'center',
      nameTextStyle: {
        padding: [0, 0, 30, 0]
      },
      type: 'value',
      minInterval: 10,
      maxInterval: 20,
      min: (range: any) => 10 * Math.ceil(range.min / 10 - 1),
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
      ...currentZoom,
      labelFormatter: (_value: number, valueStr: string) => {
        return formatToDateOrEpoch(valueStr)
      },
      xAxisIndex: [1],
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

const getDataZoomValues = () => {
  const chartOptions = chart.value?.getOption()
  const start: number = get(chartOptions, 'dataZoom[0].start', 80) as number
  const end: number = get(chartOptions, 'dataZoom[0].end', 100) as number
  return {
    start,
    end
  }
}

const getZoomTimestamps = () => {
  const max = categories.value.length - 1
  if (max <= 0) {
    return
  }
  const zoomValues = getDataZoomValues()
  const toIndex = Math.ceil(max / 100 * zoomValues.end)
  const fromIndex = Math.floor(max / 100 * zoomValues.start)
  return {
    ...zoomValues,
    toIndex,
    toTs: categories.value[toIndex],
    fromIndex,
    fromTs: categories.value[fromIndex]
  }
}

const validateDataZoom = (instant?: boolean) => {
  if (!chart.value) {
    return
  }
  const timestamps = getZoomTimestamps()
  if (!timestamps) {
    return
  }

  if (timestamps.toIndex - timestamps.fromIndex > MAX_DATA_POINTS) {
    const max = categories.value.length - 1
    if (timestamps.start !== currentZoom.start) {
      timestamps.toIndex = Math.min(timestamps.fromIndex + MAX_DATA_POINTS, max)
      timestamps.end = timestamps.toIndex * 100 / max
      timestamps.toTs = categories.value[timestamps.toIndex]
    } else {
      timestamps.fromIndex = Math.max(0, timestamps.toIndex - MAX_DATA_POINTS)
      timestamps.start = timestamps.fromIndex * 100 / max
      timestamps.fromTs = categories.value[timestamps.fromIndex]
    }
  }
  const newTimeFrames = {
    from: timestamps.fromTs,
    to: timestamps.toTs
  }
  if (tempTimeFrames.value.to !== newTimeFrames.to || tempTimeFrames.value.from !== newTimeFrames.from) {
    if (instant) {
      instantTimeFrames(newTimeFrames)
    } else {
      bounceTimeFrames(newTimeFrames, false, true)
    }
  }
  if (timestamps.start !== currentZoom.start || timestamps.end !== currentZoom.end) {
    currentZoom.end = timestamps.end
    currentZoom.start = timestamps.start

    // check if dataZoom is ready for the action otherwise use set options
    if (get(chart.value?.getOption(), 'dataZoom[0]')) {
      chart.value.dispatchAction({
        type: 'dataZoom',
        ...currentZoom
      })
    } else {
      chart.value?.setOption({
        dataZoom: {
          ...(get(chart.value, 'xAxis[1]') || {}),
          ...currentZoom
        }
      })
    }
  }
}

watch([categories, option, chart], () => {
  validateDataZoom()
}, { immediate: true })

const onDatazoom = () => {
  validateDataZoom()
}

const onMouseMove = (e: MouseEvent) => {
  lastMouseYPos = e.offsetY
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
