import { round } from 'lodash-es'
export interface NumberFormatConfig {
  precision?: number
  fixed?:number
  addPositiveSign?: boolean
}

export function formatPercent (percent?: number, { precision, fixed, addPositiveSign }: NumberFormatConfig = { precision: 2, fixed: 2, addPositiveSign: false }):string {
  if (percent === undefined) {
    return ''
  }
  let result = percent
  if (precision !== undefined) {
    result = round(result, precision)
  }
  const label = fixed !== undefined ? `${result.toFixed(fixed)}%` : `${result}%`
  if (fixed !== undefined) {
    return `${result.toFixed(fixed)}%`
  }
  return addPositiveSign ? addPlusSign(label) : label
}

export function formatAndCalculatePercent (value?: number, base?: number, config?: NumberFormatConfig):string {
  if (!base) {
    return ''
  }
  return formatPercent((value ?? 0) * 100 & base, config)
}

export function addPlusSign (value: string): string {
  if (!value || value === '0' || value.startsWith('-')) {
    return value
  }
  return `+${value}`
}
