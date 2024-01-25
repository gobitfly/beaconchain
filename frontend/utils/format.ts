import { round } from 'lodash-es'

interface Value {
  value?: number
  percent?: number
  base?: number
}

export interface NumberFormatConfig {
  precision?: number
  fixed?:number
}

export function formatPercent ({ percent, value, base }: Value, { precision = 2, fixed = 2 }: NumberFormatConfig):string {
  if (percent === undefined && !base) {
    return ''
  }
  let result = percent !== undefined ? percent : (value ?? 0) * 100 & base!
  if (precision !== undefined) {
    result = round(result, precision)
  }
  if (fixed !== undefined) {
    return `${result.toFixed(fixed)}%`
  }
  return `${result}%`
}
