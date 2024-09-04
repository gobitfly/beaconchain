<script lang="ts" setup>
import {
  h, render,
} from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { type ECharts } from 'echarts'
import { get } from 'lodash-es'
import {
  DataZoomComponent,
  GridComponent,
  LegendComponent,
  TooltipComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import SummaryChartTooltip from './SummaryChartTooltip.vue'
import {
  getChartTextColor,
  getChartTooltipBackgroundColor,
  getSummaryChartGroupColors,
} from '~/utils/colors'
import { type GetValidatorDashboardSummaryChartResponse } from '~/types/api/validator_dashboard'
import { getGroupLabel } from '~/utils/dashboard/group'
import { formatTsToTime } from '~/utils/format'
import { API_PATH } from '~/types/customFetch'
import {
  type AggregationTimeframe,
  SUMMARY_CHART_GROUP_NETWORK_AVERAGE,
  SUMMARY_CHART_GROUP_TOTAL,
  type SummaryChartFilter,
} from '~/types/dashboard/summary'

use([
  CanvasRenderer,
  LineChart,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  GridComponent,
])

interface Props {
  filter?: SummaryChartFilter,
}

const props = defineProps<Props>()
const chart = ref<ECharts | undefined>()

const { t: $t } = useTranslation()
const colorMode = useColorMode()
const { fetch } = useCustomFetch()
const {
  secondsPerEpoch, slotToTs, tsToEpoch,
} = useNetworkStore()
const { dashboardKey } = useDashboardKey()
const { overview } = useValidatorDashboardOverviewStore()
const { groups } = useValidatorDashboardGroups()
const { latestState } = useLatestStateStore()
const latestSlot = ref(latestState.value?.current_slot || 0)
const {
  bounce: bounceTimeFrames,
  instant: instantTimeFrames,
  temp: tempTimeFrames,
  value: timeFrames,
} = useDebounceValue<{ from?: number,
  to?: number, }>({
  from: undefined,
  to: undefined,
}, 1000)
const currentZoom = {
  end: 100,
  start: 80,
}
const MAX_DATA_POINTS = 199

const {
  bounce: bounceFilter, value: filter,
} = useDebounceValue(
  props.filter,
  1000,
)
const aggregation = ref<AggregationTimeframe>('hourly')
const isLoading = ref(false)
let reloadCounter = 0

interface SeriesObject {
  data: number[],
  name: string,
  smooth: boolean,
  symbol: string,
  type: string,
}
// we don't want the series to be responsive to not trigger an auto update of the option computed
const series = ref<SeriesObject[]>([])
const chartCategories = ref<number[]>([])

const categories = computed<number[]>(() => {
  // charts have at least 5 slots delay, we give it 2 more to be sure
  if (latestSlot.value <= 7 || !aggregation.value) {
    return []
  }
  const maxSeconds
    = overview.value?.chart_history_seconds?.[aggregation.value] ?? 0
  if (!maxSeconds) {
    return []
  }
  const list: number[] = []
  let latestTs = slotToTs(latestSlot.value - 7) || 0
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
  while (latestTs > minTs) {
    list.splice(0, 0, latestTs)

    latestTs -= step
  }
  return list
})

const updateTimestamp = () => {
  latestSlot.value = latestState.value?.current_slot || 0
}

watch([
  () => props.filter?.efficiency,
  () => props.filter?.groupIds,
], () => {
  if (!props.filter?.initialised || !props.filter?.efficiency) {
    return
  }
  bounceFilter({
    ...props.filter,
    groupIds: [ ...props.filter.groupIds ],
  }, true, true)
}, {
  deep: true,
  immediate: true,
})

watch(() => props.filter?.aggregation, (agg) => {
  if (!agg) {
    return
  }
  updateTimestamp()
  aggregation.value = agg
}, { immediate: true })

const loadData = async () => {
  reloadCounter++
  const currentCounter = reloadCounter
  let newCategories: number[] = []
  if (!dashboardKey.value || (!timeFrames.value.to && !timeFrames.value.from)) {
    series.value = []
    return
  }
  isLoading.value = true
  const newSeries: SeriesObject[] = []
  try {
    const res = await fetch<GetValidatorDashboardSummaryChartResponse>(
      API_PATH.DASHBOARD_SUMMARY_CHART,
      {
        query: {
          after_ts: timeFrames.value.from,
          aggregation: aggregation.value,
          before_ts: timeFrames.value.to,
          efficiency_type: props.filter?.efficiency,
          group_ids: props.filter?.groupIds.join(','),
        },
      },
      { dashboardKey: dashboardKey.value },
    )
    if (currentCounter !== reloadCounter) {
      return // make sure we only use the data from the latest call
    }

    if (res.data) {
      newCategories = res.data.categories
      const allGroups = $t('dashboard.validator.summary.chart.all_groups')
      res.data.series.forEach((element) => {
        let name: string
        if (element.id === SUMMARY_CHART_GROUP_TOTAL) {
          name = $t('dashboard.validator.summary.chart.total')
        }
        else if (element.id === SUMMARY_CHART_GROUP_NETWORK_AVERAGE) {
          name = $t('dashboard.validator.summary.chart.average')
        }
        else {
          name = getGroupLabel($t, element.id, groups.value, allGroups)
        }
        const newObj: SeriesObject = {
          data: element.data,
          name,
          smooth: false,
          symbol: 'none',
          type: 'line',
        }
        newSeries.push(newObj)
      })
    }
  }
  catch (e) {
    if (currentCounter !== reloadCounter) {
      return // make sure we only use the data from the latest call
    }
    // TODO: Maybe we want to show an error here (either a toast or inline centred in the chart space)
  }
  isLoading.value = false
  chartCategories.value = newCategories
  series.value = newSeries
}

watch(
  [
    dashboardKey,
    filter,
    aggregation,
    timeFrames,
  ],
  () => {
    loadData()
  },
  { immediate: true },
)

const colors = computed(() => {
  return {
    background: getChartTooltipBackgroundColor(colorMode.value),
    groups: getSummaryChartGroupColors(colorMode.value),
    label: getChartTextColor(colorMode.value),
  }
})

const styles = window.getComputedStyle(document.documentElement)
const fontFamily = styles.getPropertyValue('--roboto-family')
const textSize = parseInt(styles.getPropertyValue('--standard_text_font_size'))
const fontWeightLight = parseInt(styles.getPropertyValue('--roboto-light'))
const fontWeightMedium = parseInt(styles.getPropertyValue('--roboto-medium'))
let lastMouseYPos = 0

const formatTSToDate = (value: string) => {
  return formatGoTimestamp(
    Number(value),
    undefined,
    'absolute',
    'narrow',
    $t('locales.date'),
    false,
  )
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
  switch (aggregation.value) {
    case 'epoch':
      return `${date}\n${formatTSToEpoch(value)}`
    case 'hourly':
      return `${date}\n${formatTsToTime(Number(value), $t('locales.date'))}`
    default:
      return date
  }
}

// chart options
const option = computed(() => {
  return {
    color: colors.value.groups,
    dataZoom: {
      type: 'slider',
      ...currentZoom,
      borderColor: colors.value.label,
      dataBackground: {
        areaStyle: { color: colors.value.label },
        lineStyle: { color: colors.value.label },
      },
      labelFormatter: (_value: number, valueStr: string) => {
        return formatToDateOrEpoch(valueStr)
      },
      xAxisIndex: [ 1 ],
    },
    grid: {
      containLabel: true,
      left: '5%',
      right: '5%',
      top: 10,
    },
    legend: {
      bottom: 40,
      orient: 'horizontal',
      textStyle: {
        color: colors.value.label,
        fontSize: textSize,
        fontWeight: fontWeightMedium,
      },
      type: 'scroll',
    },
    series: series.value,
    textStyle: {
      color: colors.value.label,
      fontFamily,
      fontSize: textSize,
      fontWeight: fontWeightLight,
    },
    tooltip: {
      borderColor: colors.value.background,
      confine: true,
      formatter(params: any): HTMLElement {
        const ts = parseInt(params[0].axisValue)
        let lastDif = 0
        let highlightGroup = ''
        const groupInfos = params.map((param: any) => {
          if (chart.value) {
            const distance = Math.abs(
              lastMouseYPos
              - chart.value.convertToPixel({ yAxisIndex: 0 }, param.value),
            )
            if (distance < lastDif || !highlightGroup) {
              lastDif = distance
              highlightGroup = param.seriesName
            }
          }
          return {
            color: param.color,
            efficiency: param.value,
            name: param.seriesName,
          }
        })
        const d = document.createElement('div')
        render(
          h(SummaryChartTooltip, {
            aggregation: aggregation.value,
            efficiencyType: props.filter?.efficiency || 'all',
            groupInfos,
            highlightGroup,
            t: $t,
            ts,
          }),
          d,
        )
        return d
      },
      order: 'seriesAsc',
      padding: 0,
      trigger: 'axis',
    },
    xAxis: [
      {
        axisLabel: {
          fontSize: textSize,
          formatter: formatTimestamp,
          lineHeight: 20,
        },
        boundaryGap: false,
        data: chartCategories.value,
        // xAxis of the chart
        type: 'category',
      },
      {
        boundaryGap: false,
        data: categories.value,
        show: false,
        // xAxis of the time frame selection
        type: 'category',
      },
    ],
    yAxis: {
      axisLabel: {
        fontSize: textSize,
        formatter: '{value} %',
      },
      maxInterval: 20,
      min: (range: any) =>
        range.min >= 0
          ? Math.max(0, 10 * Math.ceil(range.min / 10 - 1))
          : 10 * Math.ceil(range.min / 10 - 1),
      minInterval: 10,
      name: $t(
        `dashboard.validator.summary.chart.efficiency.${props.filter?.efficiency}`,
      ),
      nameLocation: 'center',
      nameTextStyle: {
        padding: [
          0,
          0,
          30,
          0,
        ],
      },
      silent: true,
      splitLine: {
        lineStyle: {
          color: colors.value.label, opacity: 0.3,
        },
      },
      type: 'value',
    },
  }
})

// get the current dataZoom settings in the chart
const getDataZoomValues = () => {
  const chartOptions = chart.value?.getOption()
  const start: number = get(chartOptions, 'dataZoom[0].start', 80) as number
  const end: number = get(chartOptions, 'dataZoom[0].end', 100) as number
  return {
    end,
    start,
  }
}

// get the from to values for the selected zoom settings
const getZoomTimestamps = () => {
  const max = categories.value.length - 1
  if (max <= 0) {
    return
  }
  const zoomValues = getDataZoomValues()
  const toIndex = Math.floor((max / 100) * zoomValues.end)
  const fromIndex = Math.floor((max / 100) * zoomValues.start)
  return {
    ...zoomValues,
    fromIndex,
    fromTs: categories.value[fromIndex],
    toIndex,
    toTs: categories.value[toIndex],
  }
}

// validate and adjust zoom settings
const validateDataZoom = (instant?: boolean, categoryChanged?: boolean) => {
  if (!chart.value) {
    return
  }
  const timestamps = getZoomTimestamps()
  if (!timestamps) {
    return
  }
  const firstTime = !tempTimeFrames.value.to && !tempTimeFrames.value.from

  const max = categories.value.length - 1

  const useDefault = categoryChanged || firstTime
  let dataPointsChanged = false
  if (useDefault) {
    dataPointsChanged = true
    let targetPoints = 6
    switch (aggregation.value) {
      case 'epoch':
        targetPoints = 12
        break
      case 'daily':
        targetPoints = 7
        break
      case 'weekly':
        targetPoints = 8
        break
    }
    // for dashboards with a large amount of time frames we show at least 3%
    targetPoints = Math.max(targetPoints, Math.ceil(max * 0.03))
    timestamps.toIndex = firstTime ? max : Math.max(Math.ceil(max / 100 * timestamps.end), targetPoints)
    timestamps.fromIndex = timestamps.toIndex - targetPoints
  }
  else if (timestamps.toIndex - timestamps.fromIndex > MAX_DATA_POINTS) {
    dataPointsChanged = true
    if (timestamps.start !== currentZoom.start) {
      timestamps.toIndex = Math.min(
        timestamps.fromIndex + MAX_DATA_POINTS,
        max,
      )
    }
    else {
      timestamps.fromIndex = Math.max(0, timestamps.toIndex - MAX_DATA_POINTS)
    }
  }
  else {
    let minDataPoints = 3
    if (aggregation.value === 'epoch') {
      minDataPoints = 6
    }
    if (timestamps.toIndex - timestamps.fromIndex < minDataPoints) {
      dataPointsChanged = true
      if (timestamps.start !== currentZoom.start) {
        timestamps.toIndex = Math.min(timestamps.fromIndex + minDataPoints, max)
        timestamps.fromIndex = timestamps.toIndex - minDataPoints
      }
      else {
        timestamps.fromIndex = Math.max(timestamps.toIndex - minDataPoints, 0)
        timestamps.toIndex = timestamps.fromIndex + minDataPoints
      }
    }
  }

  if (dataPointsChanged) {
    timestamps.end = (timestamps.toIndex * 100) / max
    timestamps.toTs = categories.value[timestamps.toIndex]
    timestamps.start = (timestamps.fromIndex * 100) / max
    timestamps.fromTs = categories.value[timestamps.fromIndex]
  }

  let fromTs: number | undefined = timestamps.fromTs
  let toTs: number | undefined = timestamps.toTs
  const bufferSteps = aggregation.value === 'epoch' ? 0 : 5
  // if we are on the far left/right we omit the from/to timestamp
  // to prevent webservice errors if we get slight over the limit
  // when we omit one of the time stamps the backend will use the max secons of the dashboard settings
  if (timestamps.fromIndex <= bufferSteps) {
    fromTs = undefined
  }
  else if (timestamps.toIndex >= max - bufferSteps) {
    toTs = undefined
  }
  const newTimeFrames = {
    from: fromTs,
    to: toTs,
  }
  // when the timeframes of the slider change we bounce the new timeframe for the chart
  if (
    tempTimeFrames.value.to !== newTimeFrames.to
    || tempTimeFrames.value.from !== newTimeFrames.from
  ) {
    if (instant) {
      instantTimeFrames(newTimeFrames)
    }
    else {
      bounceTimeFrames(newTimeFrames, false, true)
    }
  }
  // if we had to fix the slider ranges we need to update the zoom settings
  if (
    timestamps.start !== currentZoom.start
    || timestamps.end !== currentZoom.end
  ) {
    currentZoom.end = timestamps.end
    currentZoom.start = timestamps.start

    nextTick(() => {
      chart.value?.setOption({
        dataZoom: {
          ...(get(chart.value, 'xAxis[1]') || {}),
          ...currentZoom,
        },
      })
      if (get(chart.value?.getOption(), 'dataZoom[0]')) {
        chart.value?.dispatchAction({
          type: 'dataZoom',
          ...currentZoom,
        })
      }
    })
  }
}

watch([
  chart,
  categories,
], ([
  cha,
  cat,
], [
  oldCha,
  oldCat,
]) => {
  const chartChanged = cha !== oldCha
  const categoriesChanged = cat?.length !== oldCat?.length
  if (chartChanged) {
    updateTimestamp()
  }
  validateDataZoom(true, categoriesChanged)
}, { immediate: true })

const onDatazoom = () => {
  updateTimestamp()
  validateDataZoom()
}

// we store the last mouse position so we can highlight the closest entry in the tooltip
const onMouseMove = (e: MouseEvent) => {
  lastMouseYPos = e.offsetY
}
</script>

<template>
  <div
    class="summary-chart-container"
    @mousemove="onMouseMove"
  >
    <ClientOnly>
      <VChart
        ref="chart"
        class="chart"
        :option
        autoresize
        @datazoom="onDatazoom"
      />
      <BcLoadingSpinner
        v-if="isLoading"
        class="loading-spinner"
        :loading="true"
        alignment="center"
      />
      <div
        v-if="!isLoading && !series?.length"
        class="no-data"
        alignment="center"
      >
        {{ $t("dashboard.validator.summary.chart.no_data") }}
      </div>
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
  .no-data {
    position: absolute;
    display: flex;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    justify-content: center;
    align-items: center;
    pointer-events: none;
  }
}
</style>
