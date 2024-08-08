import type { ColorDefinition } from '~/types/color'

export function getSummaryChartGroupColors(theme: string) {
  const colorsLight = [
    '#ffaa31',
    '#7db5ec',
    '#b2df27',
    '#5d78dc',
    '#ffdb58',
    '#f067e9',
    '#57bd64',
    '#a448c0',
    '#dc2a7f',
    '#e6beff',
    '#87ceeb',
    '#438d61',
    '#e7416a',
    '#6be4d8',
    '#fabebe',
    '#90d9a5',
    '#ff6a00',
    '#ffbe7c',
    '#bcb997',
    '#deb244',
    '#dda0dd',
    '#fa8072',
    '#d2b48c',
    '#6b8e23',
    '#0e8686',
    '#9a6324',
    '#932929',
    '#808000',
    '#30308e',
    '#708090',
  ]
  const colorsDark = [
    '#ffaa31',
    '#7db5ec',
    '#c3f529',
    '#5d78dc',
    '#ffdb58',
    '#f067e9',
    '#57bd64',
    '#a448c0',
    '#dc2a7f',
    '#e6beff',
    '#87ceeb',
    '#438d61',
    '#e7416a',
    '#6be4d8',
    '#fabebe',
    '#aaffc3',
    '#ff6a00',
    '#ffd8b1',
    '#fffac8',
    '#deb244',
    '#dda0dd',
    '#fa8072',
    '#d2b48c',
    '#6b8e23',
    '#0e8686',
    '#9a6324',
    '#932929',
    '#808000',
    '#30308e',
    '#708090',
  ]

  return theme === 'light' ? colorsLight : colorsDark
}

export function getChartTextColor(theme: string) {
  const styles = window.getComputedStyle(document.documentElement)

  if (theme === 'light') {
    return styles.getPropertyValue('--light-black')
  }
  else {
    return styles.getPropertyValue('--light-grey')
  }
}

export function getChartTooltipBackgroundColor(theme: string) {
  const styles = window.getComputedStyle(document.documentElement)

  if (theme === 'light') {
    return styles.getPropertyValue('--light-grey-3')
  }
  else {
    return styles.getPropertyValue('--dark-grey')
  }
}

export function getRewardsChartLineColor(theme: string) {
  const styles = window.getComputedStyle(document.documentElement)

  if (theme === 'light') {
    return styles.getPropertyValue('--light-grey-3')
  }
  else {
    return styles.getPropertyValue('--dark-grey')
  }
}

export function getRewardChartColors() {
  const styles = window.getComputedStyle(document.documentElement)

  return {
    el: styles.getPropertyValue('--primary-orange'),
    cl: styles.getPropertyValue('--melllow-blue'),
  }
}

export const APP_COLORS: ColorDefinition[] = [
  { identifier: '--white', color: '#ffffff' },
  { identifier: '--grey', color: '#a5a5a5' },
  { identifier: '--medium-grey', color: '#b3b3b3' },
  { identifier: '--grey-1', color: '#e9e9e9' },
  { identifier: '--grey-4', color: '#f4f4f4' },
  { identifier: '--light-grey', color: '#f0f0f0' },
  { identifier: '--light-grey-2', color: '#dfdfdf' },
  { identifier: '--light-grey-3', color: '#d3d3d3' },
  { identifier: '--light-grey-4', color: '#dddddd' },
  { identifier: '--light-grey-5', color: '#eaeaea' },
  { identifier: '--light-grey-6', color: '#e9e9e9' },
  { identifier: '--light-grey-7', color: '#c0c0c0' },
  { identifier: '--dark-grey', color: '#5c4e4e' },
  { identifier: '--very-dark-grey', color: '#362f32' },
  { identifier: '--very-dark-grey-2', color: '#444142' },
  { identifier: '--asphalt', color: '#484f56' },
  { identifier: '--graphite', color: '#343a40' },
  { identifier: '--almost-black', color: '#232024' },
  { identifier: '--light-black', color: '#2a262c' },
  { identifier: '--dark-blue', color: '#2f2e42' },
  { identifier: '--purple-blue', color: '#3e4461' },
  { identifier: '--light-purple-blue', color: '#545C7E' },
  { identifier: '--sky-blue', color: '#cde3ee' },
  { identifier: '--light-blue', color: '#66bce9' },
  { identifier: '--melllow-blue', color: '#7db5eC' },
  { identifier: '--blue', color: '#2e82ae' },
  { identifier: '--teal-blue', color: '#116897' },
  { identifier: '--mint-green', color: '#D3E4D4' },
  { identifier: '--flashy-green', color: '#90ed7d' },
  { identifier: '--light-green', color: '#4e7451' },
  { identifier: '--green', color: '#7dc382' },
  { identifier: '--dark-green', color: '#346f39' },
  { identifier: '--flashy-red', color: '#f3454a' },
  { identifier: '--light-red', color: '#d42127' },
  { identifier: '--dark-red', color: '#ce3438' },
  { identifier: '--bold-red', color: '#b10b11' },
  { identifier: '--pastel-red', color: '#F1C5C6' },
  { identifier: '--yellow', color: '#ffd600' },
  { identifier: '--light-yellow', color: '#969100' },
  { identifier: '--purple', color: '#9747ff' },
  { identifier: '--light-orange', color: '#e77f27' },
  { identifier: '--orange', color: '#f7a35c' },
  { identifier: '--primary-color', color: '#f89811' },
  { identifier: '--primary-orange', color: '#ffaa31' },
  { identifier: '--primary-orange-hover', color: '#ff951a' },
  { identifier: '--primary-orange-pressed', color: '#ffb346' },
]
