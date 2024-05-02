import { get } from 'lodash-es'
import { getHeatmapContentColors } from '../colors'
import { drawEqualPieChart } from '~/utils/canvas'

interface BackgroundOption {
  width: number,
  height: number,
  backgroundColor: {
    image: HTMLCanvasElement
  }
}

export function getBackgroundFormat (data: {
  proposal: boolean,
  sync: boolean,
  slashing: boolean
}) {
  const key = Object.entries(data).filter(([_key, value]) => value).map(([key]) => key).join('_')

  return `{${key}|}`
}

export function getRichBackgroundOptions (theme: string) {
  const colors = getHeatmapContentColors(theme)
  const keys = Object.keys(colors)

  const combinations: string[][] = []

  for (let i = 0; i < keys.length; i++) {
    combinations.push([keys[i]])
    if (keys[i + 1]) {
      combinations.push([keys[i], keys[i + 1]])
    }
    if (keys[i + 2]) {
      combinations.push([keys[i], keys[i + 1], keys[i + 2]])
      combinations.push([keys[i], keys[i + 2]])
    }
  }

  const options: Record<string, BackgroundOption> = {}

  return combinations.reduce((result, combo) => {
    result[combo.join('_')] = {
      width: 14,
      height: 14,
      backgroundColor: {
        image: drawEqualPieChart(combo.map(key => get(colors, key)), 14)
      }
    }
    return result
  }, options)
}
