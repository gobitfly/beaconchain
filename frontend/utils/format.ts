import { round } from 'lodash-es'

interface Value {
  value?: number
  percent?: number
  base?: number
}

export interface NumberFormatConfig {
  precision?: number
  fixed?:number
  addPositiveSign?: boolean
}

export function formatPercent ({ percent, value, base }: Value, { precision, fixed, addPositiveSign }: NumberFormatConfig = { precision: 5, fixed: 5, addPositiveSign: false }):string {
  if (percent === undefined && !base) {
    return ''
  }
  let result = percent !== undefined ? percent : (value ?? 0) * 100 & base!
  if (precision !== undefined) {
    result = round(result, precision)
  }
  const label = fixed !== undefined ? `${result.toFixed(fixed)}%` : `${result}%`
  if (fixed !== undefined) {
    return `${result.toFixed(fixed)}%`
  }
  return addPositiveSign ? addPlusSign(label) : label
}

export function addPlusSign (value: string): string {
  if (!value || value === '0' || value.startsWith('-')) {
    return value
  }
  return `+${value}`
}
