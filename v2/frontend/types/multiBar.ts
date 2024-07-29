import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'

export type MultiBarItem = {
  icon?: IconDefinition
  component?: Component,
  componentProps?: any,
  componentClass?: string,
  value: string,
  tooltip?: string,
  className?: string,
  disabled?: boolean
}
