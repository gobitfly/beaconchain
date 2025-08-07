<script lang="ts" setup>
import {
  h, render,
} from 'vue'
import {
  type ElementEvent, use,
} from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart } from 'echarts/charts'
import {
  DatasetComponent,
  DataZoomComponent,
  GridComponent,
  LegendComponent,
  TooltipComponent,
  TransformComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'

import type {
  BarSeriesOption,
  EChartsOption,
  EChartsType,
} from 'echarts'
import {
  getChartTextColor,
  getChartTooltipBackgroundColor,
  getRewardChartColors,
  getRewardsChartLineColor,
} from '~/utils/colors'
import type { GetValidatorDashboardRewardsChartResponse } from '~/types/api/validator_dashboard'
import type {
  ChartData, ChartSeries,
} from '~/types/api/common'
import { DashboardChartRewardsTooltip } from '#components'

const {
  getTimestampFromEpoch,
} = useNetwork()

use([
  GridComponent,
  DatasetComponent,
  LegendComponent,
  TooltipComponent,
  DataZoomComponent,
  TransformComponent,
  BarChart,
  CanvasRenderer,
])

const { fetch } = useCustomFetch()

const {
  key,
} = useDashboard()

const data = ref<ChartData<number, string> | undefined>()

const { status } = useAsyncData(
  'validator_dashboard_rewards_chart',
  async () => {
    if (key.value === undefined) {
      data.value = undefined
      return
    }
    const res = await fetch<GetValidatorDashboardRewardsChartResponse>(
      'DASHBOARD_VALIDATOR_REWARDS_CHART',
      undefined,
      { dashboardKey: key.value },
    )
    data.value = res.data
  },
  {
    immediate: true,
    server: false,
    watch: [ key ],
  },
)

const { groups } = useValidatorDashboardGroups()

const { t: $t } = useTranslation()
const colorMode = useColorMode()

const colors = computed(() => {
  return {
    background: getChartTooltipBackgroundColor(colorMode.value),
    data: getRewardChartColors(),
    label: getChartTextColor(colorMode.value),
    line: getRewardsChartLineColor(colorMode.value),
  }
})

const styles = window.getComputedStyle(document.documentElement)
const fontFamily = `${styles.getPropertyValue('--roboto-family')}, ${styles.getPropertyValue('--roboto-family')}, Roboto`
const textSize = parseInt(styles.getPropertyValue('--standard_text_font_size'))
const fontWeightLight = parseInt(styles.getPropertyValue('--roboto-light'))
const fontWeightMedium = parseInt(styles.getPropertyValue('--roboto-medium'))

const {
  addCurrencies,
  clCurrency,
  elCurrency,
  formatAmount,
  selectedCurrencyMain,
} = useCurrency()

const categoryCount = computed(() => data.value?.categories?.length ?? 0)

const isGwei = ref(false)

/**
 *
 * @returns string[] of formatted values
 * if one value is smaller than `0.000000` all values will be formatted in `GWei`
 */
const autoFormatAmount = (values: string[], {
  sourceCurrency,
  targetUnit = 'base',
}: {
  sourceCurrency?: CurrencyCode,
  targetUnit?: 'base' | 'gwei',
} = {},
) => {
  let result: string[] = []
  for (const value of values) {
    const formattedValue = formatAmount(value, {
      hasCurrencyDisplay: false,
      maximumFractionDigits: 6,
      minimumFractionDigits: 6,
      sourceCurrency,
      targetUnit,
      useGrouping: false,
    })
    if (
      targetUnit === 'base'
      && isCrypto(selectedCurrencyMain.value)
      && value !== '0'
      && Number(formattedValue) === 0
    ) {
      isGwei.value = true
      break
    }
    result.push(formattedValue)
  }
  if (result.length < values.length) {
    result = autoFormatAmount(values, {
      sourceCurrency,
      targetUnit: 'gwei',
    })
  }
  return result
}

const clSeries = computed(() => data.value?.series?.filter(series => series.property === 'cl') ?? [])
const clSeriesGroupTotal = computed(() => {
  let total = Array(categoryCount.value).fill('0')
  clSeries.value.forEach((group) => {
    total = total.map((value, index) => {
      return addCurrencies({
        currencyItems: [
          { value },
          { value: group.data[index] },
        ],
      })
    })
  })
  return total
})
const clSeriesGroupTotalFormatted = computed(() => autoFormatAmount(clSeriesGroupTotal.value))

const elSeries = computed(() => data.value?.series?.filter(series => series.property === 'el') ?? [])
const elSeriesGroupTotal = computed(() => {
  let total = Array(categoryCount.value).fill('0')
  elSeries.value.forEach((group) => {
    total = total.map((value, index) => {
      return addCurrencies({
        currencyItems: [
          { value },
          {
            sourceCurrency: elCurrency,
            value: group.data[index] ?? 0,
          },
        ],
      })
    })
  })
  return total
})
const elSeriesGroupTotalFormatted = computed(
  () => autoFormatAmount(elSeriesGroupTotal.value, { sourceCurrency: elCurrency }),
)

const formatYAxisLabel = (value: string) => {
  const unit = isGwei.value ? ` (${$t('common.units.gwei')})` : ''
  return `${value}${unit} ${selectedCurrencyMain.value}`
}

const seriesId = {
  cl: 'cl',
  el: 'el',
} as const
const series = computed<BarSeriesOption[]>(() => {
  const stackId = 'stack1'
  return [
    {
      barMaxWidth: 33,
      color: colors.value.data.cl,
      data: clSeriesGroupTotalFormatted.value,
      id: seriesId.cl,
      name: $t('dashboard.validator.rewards.chart.cl'),
      stack: stackId,
      type: 'bar',
    },
    {
      barMaxWidth: 33,
      color: colors.value.data.el,
      data: elSeriesGroupTotalFormatted.value,
      id: seriesId.el,
      name: $t('dashboard.validator.rewards.chart.el'),
      stack: stackId,
      type: 'bar',
    },
  ]
})

type DataZoomEvent = {
  end: number,
  start: number,
  type: 'datazoom',
}
const dataZoomStart = ref(60)
const dataZoomEnd = ref(100)
// position of tooltip get's lost due to rerendering of `options` (> computed > currency recalculations > latest-state)
const onDatazoom = ({
  end,
  start,
}: DataZoomEvent) => {
  hideTooltip()
  dataZoomStart.value = start
  dataZoomEnd.value = end
}
const chart = useTemplateRef<EChartsType>('chart')
const tooltipPosition = ref({
  x: 0,
  y: 0,
})
const setTooltipPosition = (event: ElementEvent) => {
  tooltipPosition.value.x = event.offsetX
  tooltipPosition.value.y = event.offsetY
}
const restoreTooltip = () => {
  nextTick(() => {
    if (!chart.value) return
    chart.value.dispatchAction({
      type: 'showTip',
      x: tooltipPosition.value.x,
      y: tooltipPosition.value.y,
    })
  })
}
const hideTooltip = () => {
  tooltipPosition.value.x = 0
  tooltipPosition.value.y = 0
  if (!chart.value) return
  chart.value.dispatchAction({
    type: 'hideTip',
  })
}

const getGroupInfo = (series: ChartSeries<number, string>[], currentIndex: number) =>
  series.map(({
    data,
    id,
    property,
  }) => {
    if (property === 'cl') {
      return {
        id,
        name: groups.value.find(group => group.id === id)?.name ?? '',
        value: data[currentIndex] === '0' || data[currentIndex] === null
          ? '-'
          : formatAmount(data[currentIndex], {
              hasCurrencyDisplay: true,
              hasHigherPrecision: true,
              hasUnitDisplay: true,
              targetUnit: isGwei.value ? 'gwei' : 'base',
            }),
      }
    }
    return {
      id,
      name: groups.value.find(group => group.id === id)?.name ?? '',
      value: data[currentIndex] === '0' || data[currentIndex] === null
        ? '-'
        : formatAmount(data[currentIndex], {
            hasCurrencyDisplay: true,
            hasHigherPrecision: true,
            hasUnitDisplay: true,
            sourceCurrency: elCurrency,
            targetUnit: isGwei.value ? 'gwei' : 'base',
          }),
    }
  })

const option = computed<EChartsOption>(() => {
// position of tooltip get's lost due to rerendering of `options` (> computed > currency recalculations > latest-state)
  if (tooltipPosition.value.x !== 0 && tooltipPosition.value.y !== 0) {
    restoreTooltip()
  }
  return {
    dataZoom: {
      borderColor: colors.value.label,
      dataBackground: {
        areaStyle: { color: colors.value.label },
        lineStyle: { color: colors.value.label },
      },
      end: dataZoomEnd.value,
      labelFormatter: (_value: number, valueStr: string) => {
        const unixTimestamp = getTimestampFromEpoch(Number(valueStr))
        return getDateTime(unixTimestamp, { hasTime: false })
      },
      start: dataZoomStart.value,
      type: 'slider',
    },
    grid: {
      bottom: 80,
      containLabel: true,
      left: '5%',
      right: '5%',
      top: 20,
    },
    legend: {
      bottom: 50,
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
      alwaysShowContent: true,
      borderColor: colors.value.background,
      confine: true,
      enterable: true,
      formatter(params) {
        if (!Array.isArray(params)) return ''
        if (params.length === 0) return ''

        const paramsConsensusLayer = params.find(param => param.seriesId === seriesId.cl)
        const paramsExecutionLayer = params.find(param => param.seriesId === seriesId.el)
        const currentIndex = params[0].dataIndex
        const currentTimestamp = getTimestampFromEpoch(Number(params[0].name))
        const currentEpoch = {
          index: params[0].name,
          timestamp: currentTimestamp,
        }
        const currentGroupTotalCl = clSeriesGroupTotal.value[currentIndex]
        const currentGroupTotalEl = elSeriesGroupTotal.value[currentIndex]

        const consensusLayerRewardSum = currentGroupTotalCl === '0'
          ? '-'
          : formatAmount(currentGroupTotalCl, {
              hasHigherPrecision: true,
              hasUnitDisplay: true,
              sourceCurrency: clCurrency,
              targetUnit: isGwei.value ? 'gwei' : 'base',
            })

        const executionLayerRewardSum = currentGroupTotalEl === '0'
          ? '-'
          : formatAmount(currentGroupTotalEl, {
              hasHigherPrecision: true,
              hasUnitDisplay: true,
              sourceCurrency: elCurrency,
              targetUnit: isGwei.value ? 'gwei' : 'base',
            })

        if (typeof executionLayerRewardSum !== 'string') return ''
        const consesnsusLayerRewardSumLabel = paramsConsensusLayer?.seriesName ?? ''
        const executionLayerRewardSumLabel = paramsExecutionLayer?.seriesName ?? ''

        const groupInfo = {
          cl: getGroupInfo(clSeries.value, currentIndex),
          el: getGroupInfo(elSeries.value, currentIndex),
        }

        const d = document.createElement('div')
        render(
          h(DashboardChartRewardsTooltip, {
            consensusLayerRewardSum,
            consesnsusLayerRewardSumLabel,
            currentEpoch,
            executionLayerRewardSum,
            executionLayerRewardSumLabel,
            groupInfo,
          }),
          d,
        )
        return d
      },
      order: 'seriesAsc',
      padding: 0,
      trigger: 'axis',
      triggerOn: 'click',
    },
    xAxis: {
      axisLabel: {
        fontSize: textSize,
        fontWeight: fontWeightMedium,
        formatter: (epoch: number) => {
          const unixTimestamp = getTimestampFromEpoch(epoch)
          const date = getDateTime(unixTimestamp, { hasTime: false })
          return `${date}\n${$t('common.epoch')} ${epoch}`
        },
        lineHeight: 20,
      },
      data: data.value?.categories,
      type: 'category',
    },
    yAxis: {
      axisLabel: {
        fontSize: textSize,
        fontWeight: fontWeightMedium,
        formatter: formatYAxisLabel,
        padding: [
          0,
          10,
          0,
          0,
        ],
      },
      silent: true,
      splitLine: { lineStyle: { color: colors.value.line } },
      type: 'value',
    },
  }
})
</script>

<template>
  <div class="rewards-chart-container">
    <ClientOnly
      v-if="data"
    >
      <VChart
        ref="chart"
        class="chart"
        :option
        autoresize
        @datazoom="onDatazoom"
        @zr:click="setTooltipPosition"
      />
    </ClientOnly>
    <BcLoadingSpinner
      v-if="status === 'pending'"
      class="loading-spinner"
      :loading="true"
      alignment="center"
    />
    <div
      v-if="status === 'error'"
      class="no-data"
      alignment="center"
    >
      {{ $t("dashboard.validator.summary.chart.error") }}
    </div>
    <div
      v-if="status === 'success' && !clSeries.length"
      class="no-data"
      alignment="center"
    >
      {{ $t("dashboard.validator.summary.chart.no_data") }}
    </div>
  </div>
</template>

<style lang="scss" scoped>
.rewards-chart-container {
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
