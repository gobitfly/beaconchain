<script lang="ts" setup>

import { h, render } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart } from 'echarts/charts'
import {
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent,
  DatasetComponent,
  TransformComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import type { ECBasicOption } from 'echarts/types/dist/shared'
import { BigNumber } from '@ethersproject/bignumber'
import { formatEpochToDate } from '~/utils/format'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { getChartTextColor, getChartTooltipBackgroundColor, getRewardChartColors, getRewardsChartLineColor } from '~/utils/colors'
import { type InternalGetValidatorDashboardRewardsChartResponse } from '~/types/api/validator_dashboard'
import { type ChartData } from '~/types/api/common'
import { type RewardChartSeries, type RewardChartGroupData } from '~/types/dashboard/rewards'
import { getGroupLabel } from '~/utils/dashboard/group'
import { DashboardChartRewardsChartTooltip } from '#components'
import { API_PATH } from '~/types/customFetch'

const { currency } = useCurrency()
// TODO: once we have different chains we migh need to change the default from 'ETH' to the dashboard currency
const currencyLabel = computed(() => !currency.value || currency.value === 'NAT' ? 'ETH' : currency.value)

use([
  GridComponent,
  DatasetComponent,
  LegendComponent,
  TooltipComponent,
  DataZoomComponent,
  TransformComponent,
  BarChart,
  CanvasRenderer
])

const { fetch } = useCustomFetch()

const { dashboardKey, isPrivate: groupsEnabled } = useDashboardKey()

const data = ref<ChartData<number, string> | undefined >()
const isLoading = ref(false)

await useAsyncData('validator_dashboard_rewards_chart', async () => {
  if (dashboardKey.value === undefined) {
    data.value = undefined
    return
  }
  isLoading.value = true
  const res = await fetch<InternalGetValidatorDashboardRewardsChartResponse>(API_PATH.DASHBOARD_VALIDATOR_REWARDS_CHART, undefined, { dashboardKey: dashboardKey.value })

  isLoading.value = false
  data.value = res.data
}, { watch: [dashboardKey], server: false, immediate: true })

const { overview } = useValidatorDashboardOverviewStore()

const { t: $t } = useI18n()
const colorMode = useColorMode()

const { converter } = useValue()

const colors = computed(() => {
  return {
    data: getRewardChartColors(),
    label: getChartTextColor(colorMode.value),
    line: getRewardsChartLineColor(colorMode.value),
    background: getChartTooltipBackgroundColor(colorMode.value)
  }
})

const styles = window.getComputedStyle(document.documentElement)
const fontFamily = styles.getPropertyValue('--roboto-family')
const textSize = parseInt(styles.getPropertyValue('--standard_text_font_size'))
const fontWeightLight = parseInt(styles.getPropertyValue('--roboto-light'))
const fontWeightMedium = parseInt(styles.getPropertyValue('--roboto-medium'))

const valueFormatter = computed(() => {
  const decimals = isFiat(currency.value) ? 2 : 5
  return (value: number) => `${trim(value, decimals, decimals)} ${currencyLabel.value}`
})

const mapSeriesData = (data: RewardChartSeries) => {
  data.bigData.forEach((bigValue, index) => {
    if (!bigValue.isZero()) {
      const formatted = converter.value.weiToValue(bigValue, { fixedDecimalCount: 5, minUnit: 'MAIN' })
      data.formatedData[index] = formatted
      const parsedValue = parseFloat(`${formatted.label}`.split(' ')[0])
      if (!isNaN(parsedValue)) {
        data.data[index] = parsedValue
      }
    }
  })
}

const series = computed<RewardChartSeries[]>(() => {
  const list:RewardChartSeries[] = []
  if (!data.value?.series) {
    return list
  }

  const categoryCount = data.value?.categories.length ?? 0
  const clSeries:RewardChartSeries = {
    id: 1,
    name: $t('dashboard.validator.rewards.chart.cl'),
    color: colors.value.data.cl,
    property: 'cl',
    type: 'bar',
    stack: 'x',
    barMaxWidth: 33,
    groups: [],
    bigData: Array.from(Array(categoryCount)).map(() => BigNumber.from('0')),
    formatedData: Array.from(Array(categoryCount)).map(() => ({ label: `0 ${currencyLabel.value}` })),
    data: Array.from(Array(categoryCount)).map(() => 0)
  }
  const elSeries:RewardChartSeries = {
    id: 2,
    name: $t('dashboard.validator.rewards.chart.el'),
    color: colors.value.data.el,
    property: 'el',
    type: 'bar',
    stack: 'x',
    barMaxWidth: 33,
    groups: [],
    bigData: Array.from(Array(categoryCount)).map(() => BigNumber.from('0')),
    formatedData: Array.from(Array(categoryCount)).map(() => ({ label: `0 ${currencyLabel.value}` })),
    data: Array.from(Array(categoryCount)).map(() => 0)
  }
  list.push(elSeries)
  list.push(clSeries)
  data.value.series.forEach((group) => {
    let name
    if (!groupsEnabled) {
      name = $t('dashboard.validator.rewards.chart.rewards')
    } else {
      name = getGroupLabel($t, group.id, overview.value?.groups)
    }
    const newData: RewardChartGroupData = {
      id: group.id,
      bigData: [],
      name
    }
    for (let i = 0; i < categoryCount; i++) {
      const bigValue = group.data[i] ? BigNumber.from(group.data[i]) : BigNumber.from('0')

      if (!bigValue.isZero()) {
        if (group.property === 'el') {
          elSeries.bigData[i] = elSeries.bigData[i].add(bigValue)
        } else {
          clSeries.bigData[i] = clSeries.bigData[i].add(bigValue)
        }
      }
      newData.bigData.push(bigValue)
    }

    if (group.property === 'el') {
      elSeries.groups.push(newData)
    } else {
      clSeries.groups.push(newData)
    }
  })
  mapSeriesData(elSeries)
  mapSeriesData(clSeries)
  return list
})

const option = computed<ECBasicOption | undefined>(() => {
  if (series.value === undefined) {
    return undefined
  }

  return {
    grid: {
      containLabel: true,
      top: 20,
      left: '5%',
      right: '5%'
    },
    xAxis: {
      type: 'category',
      data: data.value?.categories,
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
      type: 'value',
      silent: true,
      axisLabel: {
        formatter: valueFormatter.value,
        fontSize: textSize,
        padding: [0, 10, 0, 0]
      },
      splitLine: {
        lineStyle: {
          color: colors.value.line
        }
      }
    },
    series: series.value,
    textStyle: {
      fontFamily,
      fontSize: textSize,
      fontWeight: fontWeightLight,
      color: colors.value.label
    },
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
      triggerOn: 'click',
      padding: 0,
      borderColor: colors.value.background,
      formatter (params : any) : HTMLElement {
        const startEpoch = parseInt(params[0].axisValue)
        const dataIndex = params[0].dataIndex

        const d = document.createElement('div')
        render(h(DashboardChartRewardsChartTooltip, { t: $t, startEpoch, dataIndex, series: series.value, weiToValue: converter.value.weiToValue }), d)
        return d
      }
    },
    dataZoom: {
      type: 'slider',
      start: 60,
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
