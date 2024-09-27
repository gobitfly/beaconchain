<script lang="ts" setup>
import {
  h, render,
} from 'vue'
import { use } from 'echarts/core'
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
import type { ECBasicOption } from 'echarts/types/dist/shared'
import { BigNumber } from '@ethersproject/bignumber'
import {
  getChartTextColor,
  getChartTooltipBackgroundColor,
  getRewardChartColors,
  getRewardsChartLineColor,
} from '~/utils/colors'
import { type GetValidatorDashboardRewardsChartResponse } from '~/types/api/validator_dashboard'
import { type ChartData } from '~/types/api/common'
import {
  type RewardChartGroupData,
  type RewardChartSeries,
} from '~/types/dashboard/rewards'
import { getGroupLabel } from '~/utils/dashboard/group'
import { DashboardChartRewardsChartTooltip } from '#components'
import { API_PATH } from '~/types/customFetch'
import { useNetworkStore } from '~/stores/useNetworkStore'
import { useFormat } from '~/composables/useFormat'
import type { CryptoUnits } from '~/types/currencies'

const { formatEpochToDate } = useFormat()
const { networkInfo } = useNetworkStore()
const networkNativeELcurrency = computed(() => networkInfo.value.elCurrency)
const { currency } = useCurrency()
const currencyLabel = computed(() =>
  !currency.value || currency.value === 'NAT'
    ? networkNativeELcurrency.value
    : currency.value,
)

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
  dashboardKey, isPrivate: groupsEnabled,
} = useDashboardKey()

const data = ref<ChartData<number, string> | undefined>()
const isLoading = ref(false)

useAsyncData(
  'validator_dashboard_rewards_chart',
  async () => {
    if (dashboardKey.value === undefined) {
      data.value = undefined
      return
    }
    isLoading.value = true
    const res = await fetch<GetValidatorDashboardRewardsChartResponse>(
      API_PATH.DASHBOARD_VALIDATOR_REWARDS_CHART,
      undefined,
      { dashboardKey: dashboardKey.value },
    )

    isLoading.value = false
    data.value = res.data
  },
  {
    immediate: true,
    server: false,
    watch: [ dashboardKey ],
  },
)

const { groups } = useValidatorDashboardGroups()

const { t: $t } = useTranslation()
const colorMode = useColorMode()

const { converter } = useValue()

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

const valueFormatter = computed(() => {
  const fiat = isFiat(currency.value)
  const label = fiat || series.value?.minUnit === 'MAIN' ? currencyLabel.value : series.value?.minUnit
  const decimals = isFiat(currency.value) ? 2 : 5
  return (value: number) =>
    `${trim(value, decimals, decimals)} ${label}`
})

const getMinUnit = (data: RewardChartSeries[]): CryptoUnits => {
  let unit: CryptoUnits = 'MAIN'
  for (const seriesI in data) {
    const series = data[seriesI]
    for (const bigValueI in series.bigData) {
      const bigValue = series.bigData[bigValueI]
      if (bigValue.isZero()) {
        continue
      }
      if (lessThanGwei(bigValue, 5)) {
        return 'WEI'
      }
      if (lessThanEth(bigValue, 5)) {
        unit = 'GWEI'
        break
      }
    }
    if (unit === 'GWEI') {
      break
    }
  }
  return unit
}

const mapSeriesData = (data: RewardChartSeries, minUnit: CryptoUnits) => {
  data.bigData.forEach((bigValue, index) => {
    if (!bigValue.isZero()) {
      const formatted = converter.value.weiToValue(bigValue, {
        fixedDecimalCount: 5,
        minUnit,
      })
      data.formatedData[index] = formatted
      const parsedValue = parseFloat(`${formatted.label}`.split(' ')[0])
      if (!isNaN(parsedValue)) {
        data.data[index] = parsedValue
      }
    }
  })
}

const series = computed<{ list: RewardChartSeries[], minUnit: CryptoUnits } >(() => {
  const list: RewardChartSeries[] = []
  if (!data.value?.series) {
    return {
      list, minUnit: 'MAIN',
    }
  }

  const categoryCount = data.value?.categories.length ?? 0
  const clSeries: RewardChartSeries = {
    barMaxWidth: 33,
    bigData: Array.from(Array(categoryCount)).map(() => BigNumber.from('0')),
    color: colors.value.data.cl,
    data: Array.from(Array(categoryCount)).map(() => 0),
    formatedData: Array.from(Array(categoryCount)).map(() => ({ label: `0 ${currencyLabel.value}` })),
    groups: [],
    id: 1,
    name: $t('dashboard.validator.rewards.chart.cl'),
    property: 'cl',
    stack: 'x',
    type: 'bar',
  }
  const elSeries: RewardChartSeries = {
    barMaxWidth: 33,
    bigData: Array.from(Array(categoryCount)).map(() => BigNumber.from('0')),
    color: colors.value.data.el,
    data: Array.from(Array(categoryCount)).map(() => 0),
    formatedData: Array.from(Array(categoryCount)).map(() => ({ label: `0 ${currencyLabel.value}` })),
    groups: [],
    id: 2,
    name: $t('dashboard.validator.rewards.chart.el'),
    property: 'el',
    stack: 'x',
    type: 'bar',
  }
  list.push(elSeries)
  list.push(clSeries)
  data.value.series.forEach((group) => {
    let name
    if (!groupsEnabled) {
      name = $t('dashboard.validator.rewards.chart.rewards')
    }
    else {
      name = getGroupLabel($t, group.id, groups.value)
    }
    const newData: RewardChartGroupData = {
      bigData: [],
      id: group.id,
      name,
    }
    for (let i = 0; i < categoryCount; i++) {
      const bigValue = group.data[i]
        ? BigNumber.from(group.data[i])
        : BigNumber.from('0')

      if (!bigValue.isZero()) {
        if (group.property === 'el') {
          elSeries.bigData[i] = elSeries.bigData[i].add(bigValue)
        }
        else {
          clSeries.bigData[i] = clSeries.bigData[i].add(bigValue)
        }
      }
      newData.bigData.push(bigValue)
    }

    if (group.property === 'el') {
      elSeries.groups.push(newData)
    }
    else {
      clSeries.groups.push(newData)
    }
  })
  const min = getMinUnit([
    elSeries,
    clSeries,
  ])
  mapSeriesData(elSeries, min)
  mapSeriesData(clSeries, min)
  return {
    list, minUnit: min,
  }
})

const option = computed<ECBasicOption | undefined>(() => {
  if (series.value === undefined) {
    return undefined
  }

  return {
    dataZoom: {
      borderColor: colors.value.label,
      dataBackground: {
        areaStyle: { color: colors.value.label },
        lineStyle: { color: colors.value.label },
      },
      end: 100,
      labelFormatter: (_value: number, valueStr: string) => {
        return formatEpochToDate(parseInt(valueStr), $t('locales.date'))
      },
      start: 60,
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
    series: series.value.list,
    textStyle: {
      color: colors.value.label,
      fontFamily,
      fontSize: textSize,
      fontWeight: fontWeightLight,
    },
    tooltip: {
      borderColor: colors.value.background,
      formatter(params: any): HTMLElement {
        const startEpoch = parseInt(params[0].axisValue)
        const dataIndex = params[0].dataIndex

        const d = document.createElement('div')
        render(
          h(DashboardChartRewardsChartTooltip, {
            dataIndex,
            series: series.value.list,
            startEpoch,
            t: $t,
            weiToValue: converter.value.weiToValue,
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
        formatter: (value: number) => {
          const date = formatEpochToDate(value, $t('locales.date'))
          if (date === undefined) {
            return ''
          }

          return `${date}\n${$t('common.epoch')} ${value}`
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
        formatter: valueFormatter.value,
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
    <ClientOnly>
      <VChart ref="chart" class="chart" :option autoresize />
    </ClientOnly>
    <BcLoadingSpinner v-if="isLoading" class="loading-spinner" :loading="true" alignment="center" />
    <div v-if="!isLoading && !series?.list?.length" class="no-data" alignment="center">
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
