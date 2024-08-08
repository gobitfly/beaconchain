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
  DataZoomComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import SummaryChartTooltip from './SummaryChartTooltip.vue'
import {
  getSummaryChartGroupColors,
  getChartTextColor,
  getChartTooltipBackgroundColor,
} from '~/utils/colors'
import { type InternalGetValidatorDashboardSummaryChartResponse } from '~/types/api/validator_dashboard'
import { getGroupLabel } from '~/utils/dashboard/group'
import { formatTsToTime } from '~/utils/format'
import { API_PATH } from '~/types/customFetch'
import {
  SUMMARY_CHART_GROUP_NETWORK_AVERAGE,
  SUMMARY_CHART_GROUP_TOTAL,
  type AggregationTimeframe,
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
  filter?: SummaryChartFilter
}

const props = defineProps<Props>()
const chart = ref<ECharts | undefined>()

const { t: $t } = useTranslation()
const colorMode = useColorMode()
const { fetch } = useCustomFetch()
const { tsToEpoch, slotToTs, secondsPerEpoch } = useNetworkStore()
const { dashboardKey } = useDashboardKey()
const { overview } = useValidatorDashboardOverviewStore()
const { groups } = useValidatorDashboardGroups()
const { latestState } = useLatestStateStore()
const { convertColors } = useColorBlindMode()
const latestSlot = ref(latestState.value?.current_slot || 0)
const {
  value: timeFrames,
  temp: tempTimeFrames,
  bounce: bounceTimeFrames,
  instant: instantTimeFrames,
} = useDebounceValue<{ from?: number, to: number }>({ from: undefined, to: 0 }, 1000)
const currentZoom = { start: 80, end: 100 }
const MAX_DATA_POINTS = 200

const { value: filter, bounce: bounceFilter } = useDebounceValue(
  props.filter,
  1000,
)
const aggregation = ref<AggregationTimeframe>('hourly')
const isLoading = ref(false)
let reloadCounter = 0

interface SeriesObject {
  data: number[]
  type: string
  smooth: boolean
  symbol: string
  name: string
}
// we don't want the series to be responsive to not trigger an auto update of the option computed
const series = ref<SeriesObject[]>([])
const chartCategories = ref<number[]>([])

const categories = computed<number[]>(() => {
  // charts have 5 slots delay
  if (latestSlot.value <= 5 || !aggregation.value) {
    return []
  }
  const maxSeconds
    = overview.value?.chart_history_seconds?.[aggregation.value] ?? 0
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
  while (latestTs > minTs) {
    list.splice(0, 0, latestTs)

    latestTs -= step
  }
  return list
})

const updateTimestamp = () => {
  latestSlot.value = latestState.value?.current_slot || 0
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
  updateTimestamp()
  aggregation.value = agg
}, { immediate: true })

const loadData = async () => {
  reloadCounter++
  const currentCounter = reloadCounter
  let newCategories: number[] = []
  if (!dashboardKey.value || !timeFrames.value.to) {
    series.value = []
    return
  }
  isLoading.value = true
  const newSeries: SeriesObject[] = []
  try {
    const res = await fetch<InternalGetValidatorDashboardSummaryChartResponse>(
      API_PATH.DASHBOARD_SUMMARY_CHART,
      {
        query: {
          after_ts: timeFrames.value.from,
          before_ts: timeFrames.value.to,
          group_ids: props.filter?.groupIds.join(','),
          efficiency_type: props.filter?.efficiency,
          aggregation: aggregation.value,
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
          type: 'line',
          smooth: false,
          symbol: 'none',
          name,
        }
        newSeries.push(newObj)
      })
    }
  }
  catch (e) {
    // TODO: Maybe we want to show an error here (either a toast or inline centred in the chart space)
  }
  isLoading.value = false
  chartCategories.value = newCategories
  series.value = newSeries
}

watch(
  [dashboardKey, filter, aggregation, timeFrames],
  () => {
    loadData()
  },
  { immediate: true },
)

const colors = computed(() => {
  const label = getChartTextColor(colorMode.value)
  const background = getChartTooltipBackgroundColor(colorMode.value)
  // we pass the label and background color to the color conversion so it can adjust the summary line colors better
  // but we will not use the converted label/background colors as we should use the same as the rest of the app.
  const list = convertColors([
    { identifier: 'label', color: label },
    { identifier: 'background', color: background },
    ...getSummaryChartGroupColors(colorMode.value).map((c, index) => ({ identifier: `chart_${index}`, color: c })),
  ])
  return {
    label,
    background,
    groups: list.slice(2).map(c => c.color),
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
    grid: {
      containLabel: true,
      top: 10,
      left: '5%',
      right: '5%',
    },
    xAxis: [
      {
        // xAxis of the chart
        type: 'category',
        data: chartCategories.value,
        boundaryGap: false,
        axisLabel: {
          fontSize: textSize,
          lineHeight: 20,
          formatter: formatTimestamp,
        },
      },
      {
        // xAxis of the time frame selection
        type: 'category',
        data: categories.value,
        show: false,
        boundaryGap: false,
      },
    ],
    series: series.value,
    yAxis: {
      name: $t(
        `dashboard.validator.summary.chart.efficiency.${props.filter?.efficiency}`,
      ),
      nameLocation: 'center',
      nameTextStyle: {
        padding: [0, 0, 30, 0],
      },
      type: 'value',
      minInterval: 10,
      maxInterval: 20,
      min: (range: any) =>
        range.min >= 0
          ? Math.max(0, 10 * Math.ceil(range.min / 10 - 1))
          : 10 * Math.ceil(range.min / 10 - 1),
      silent: true,
      axisLabel: {
        formatter: '{value} %',
        fontSize: textSize,
      },
      splitLine: {
        lineStyle: {
          color: colors.value.label,
        },
      },
    },
    textStyle: {
      fontFamily,
      fontSize: textSize,
      fontWeight: fontWeightLight,
      color: colors.value.label,
    },
    color: colors.value.groups,
    legend: {
      type: 'scroll',
      orient: 'horizontal',
      bottom: 40,
      textStyle: {
        color: colors.value.label,
        fontSize: textSize,
        fontWeight: fontWeightMedium,
      },
    },
    tooltip: {
      order: 'seriesAsc',
      trigger: 'axis',
      padding: 0,
      borderColor: colors.value.background,
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
            name: param.seriesName,
            efficiency: param.value,
            color: param.color,
          }
        })
        const d = document.createElement('div')
        render(
          h(SummaryChartTooltip, {
            t: $t,
            ts,
            efficiencyType: props.filter?.efficiency || 'all',
            aggregation: aggregation.value,
            groupInfos,
            highlightGroup,
          }),
          d,
        )
        return d
      },
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
          color: colors.value.label,
        },
        areaStyle: {
          color: colors.value.label,
        },
      },
      borderColor: colors.value.label,
    },
  }
})

// get the current dataZoom settings in the chart
const getDataZoomValues = () => {
  const chartOptions = chart.value?.getOption()
  const start: number = get(chartOptions, 'dataZoom[0].start', 80) as number
  const end: number = get(chartOptions, 'dataZoom[0].end', 100) as number
  return {
    start,
    end,
  }
}

// get the from to values for the selected zoom settings
const getZoomTimestamps = () => {
  const max = categories.value.length - 1
  if (max <= 0) {
    return
  }
  const zoomValues = getDataZoomValues()
  const toIndex = Math.ceil((max / 100) * zoomValues.end)
  const fromIndex = Math.floor((max / 100) * zoomValues.start)
  return {
    ...zoomValues,
    toIndex,
    toTs: categories.value[toIndex],
    fromIndex,
    fromTs: categories.value[fromIndex],
  }
}

// validate and adjust zoom settings
const validateDataZoom = (instant?: boolean) => {
  if (!chart.value) {
    return
  }
  const timestamps = getZoomTimestamps()
  if (!timestamps) {
    return
  }

  const max = categories.value.length - 1
  // check for max data points
  if (timestamps.toIndex - timestamps.fromIndex > MAX_DATA_POINTS) {
    if (timestamps.start !== currentZoom.start) {
      timestamps.toIndex = Math.min(
        timestamps.fromIndex + MAX_DATA_POINTS,
        max,
      )
      timestamps.end = (timestamps.toIndex * 100) / max
      timestamps.toTs = categories.value[timestamps.toIndex]
    }
    else {
      timestamps.fromIndex = Math.max(0, timestamps.toIndex - MAX_DATA_POINTS)
      timestamps.start = (timestamps.fromIndex * 100) / max
      timestamps.fromTs = categories.value[timestamps.fromIndex]
    }
  }
  // to index must be greater then from index
  if (timestamps.toIndex <= timestamps.fromIndex) {
    if (
      (timestamps.start !== currentZoom.start
      && timestamps.fromIndex !== max)
      || timestamps.toIndex === 0
    ) {
      timestamps.toIndex = timestamps.fromIndex + 1
      timestamps.end = (timestamps.toIndex * 100) / max
      timestamps.toTs = categories.value[timestamps.toIndex]
    }
    else {
      timestamps.fromIndex = timestamps.toIndex - 1
      timestamps.start = (timestamps.fromIndex * 100) / max
      timestamps.fromTs = categories.value[timestamps.fromIndex]
    }
  }

  let fromTs: number | undefined = timestamps.fromTs
  const bufferSteps = aggregation.value === 'epoch' ? 0 : 5
  // if we are on the far left of the time frame we omit the fromTs to avoid going to far and cause a webservice error
  // in that case the backend will go back depending on the max secons of the dashboard settings
  if (timestamps.fromIndex <= bufferSteps) {
    fromTs = undefined
  }
  const newTimeFrames = {
    from: fromTs,
    to: timestamps.toTs,
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

    // check if dataZoom is ready for the action otherwise use set options
    nextTick(() => {
      if (get(chart.value?.getOption(), 'dataZoom[0]')) {
        chart.value?.dispatchAction({
          type: 'dataZoom',
          ...currentZoom,
        })
      }
      else {
        chart.value?.setOption({
          dataZoom: {
            ...(get(chart.value, 'xAxis[1]') || {}),
            ...currentZoom,
          },
        })
      }
    })
  }
}

watch([option], () => {
  updateTimestamp()
  validateDataZoom(true)
}, { immediate: true })

watch([categories, chart], () => {
  validateDataZoom(true)
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
        :option="option"
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
