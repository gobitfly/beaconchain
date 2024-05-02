<script lang="ts" setup>

import { h, render } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { HeatmapChart } from 'echarts/charts'
import {
  TooltipComponent,
  GridComponent,
  DataZoomComponent,
  DatasetComponent,
  TransformComponent,
  VisualMapComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import type { ECBasicOption } from 'echarts/types/dist/shared'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { getChartTextColor, getChartTooltipBackgroundColor, getHeatmapColors, getRewardChartColors, getRewardsChartLineColor } from '~/utils/colors'
import { getGroupLabel } from '~/utils/dashboard/group'
import { useValidatorDashboardHeatmapStore } from '~/stores/dashboard/useValidatorDashboardHeatmapStore'
import { getRichBackgroundOptions, getBackgroundFormat } from '~/utils/dashboard/heatmap'
import { BcLoadingSpinner } from '#components'

use([
  GridComponent,
  DatasetComponent,
  TooltipComponent,
  DataZoomComponent,
  TransformComponent,
  VisualMapComponent,
  HeatmapChart,
  CanvasRenderer
])

const { dashboardKey } = useDashboardKey()

const { heatmap, isLoading, getHeatmap, getHeatmapTooltip } = useValidatorDashboardHeatmapStore()

const { overview } = useValidatorDashboardOverviewStore()

watch([dashboardKey, overview], () => {
  getHeatmap(dashboardKey.value)
}, { immediate: true })

const { t: $t } = useI18n()
const colorMode = useColorMode()

const colors = computed(() => {
  return {
    data: getRewardChartColors(),
    label: getChartTextColor(colorMode.value),
    line: getRewardsChartLineColor(colorMode.value),
    background: getChartTooltipBackgroundColor(colorMode.value),
    heatmap: getHeatmapColors(colorMode.value),
    richOptions: getRichBackgroundOptions(colorMode.value)
  }
})

const styles = window.getComputedStyle(document.documentElement)
const fontFamily = styles.getPropertyValue('--roboto-family')
const textSize = parseInt(styles.getPropertyValue('--standard_text_font_size'))
const fontWeightLight = parseInt(styles.getPropertyValue('--roboto-light'))

const groupNameLabel = (groupId?: number) => {
  return getGroupLabel($t, groupId, overview.value?.groups)
}

const ttFormatter = ({ data }: { data: number[] }): HTMLElement => {
  const d = document.createElement('div')
  d.style.width = '100px'
  d.style.height = '100px'
  render(h(BcLoadingSpinner, { loading: true, alignment: 'center' }), d)

  getHeatmapTooltip(dashboardKey.value, data[0], data[1]).then((tt) => {
    d.innerHTML = tt ? JSON.stringify(tt) : 'no data'
  })
  return d
}

const option = computed<ECBasicOption | undefined>(() => {
  if (heatmap.value === undefined) {
    return undefined
  }
  return {
    grid: {
      containLabel: true,
      top: 20,
      left: '5%',
      right: '5%',
      height: '75%'
    },
    xAxis: {
      type: 'category',
      data: heatmap.value.epochs,
      splitArea: {
        show: true
      }
    },
    yAxis: {
      type: 'category',
      data: heatmap.value.group_ids.map(groupNameLabel),
      splitArea: {
        show: true
      }
    },
    visualMap: {
      min: 0,
      max: 100,
      calculable: true,
      orient: 'horizontal',
      right: '5%',
      bottom: '0',
      left: '5%',
      itemHeight: '500px',
      inRange: {
        color: colors.value.heatmap
      }
    },
    series: [
      {
        name: 'Attestations',
        type: 'heatmap',
        data: heatmap.value.data.map(d => [d.x, d.y, d.value]),
        label: {
          show: true,
          formatter: ({ data }: { data: number[] }) => {
            const event = heatmap.value?.events?.find(e => e.x === data[0] && e.x === data[1])
            if (!event) {
              return ''
            }
            const f = getBackgroundFormat({
              proposal: event.proposal,
              sync: event.sync,
              slashing: event.slash
            })
            return f
          },
          rich: colors.value.richOptions
        }
      }
    ],
    textStyle: {
      fontFamily,
      fontSize: textSize,
      fontWeight: fontWeightLight,
      color: colors.value.label
    },
    tooltip: {
      trigger: 'item',
      triggerOn: 'click',
      padding: 0,
      borderColor: colors.value.background,
      formatter: ttFormatter
    },
    dataZoom: {
      type: 'slider',
      start: 60,
      end: 100,
      bottom: 50,
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
  <div class="heatmap">
    <ClientOnly>
      <BcLoadingSpinner v-if="isLoading" :loading="true" alignment="center" />
      <VChart v-else class="chart" :option="option" autoresize />
    </ClientOnly>
  </div>
</template>

<style lang="scss">
.heatmap {
  height: 870px;
  width: 100%;
}
</style>
