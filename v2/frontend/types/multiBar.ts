import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'

export type MultiBarItem = {
  className?: string
  component?: Component
  componentClass?: string
  componentProps?: any
  disabled?: boolean
  icon?: IconDefinition
  tooltip?: string
  value: string
}
