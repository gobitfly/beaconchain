import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'

export type MultiBarItem = {
  icon?: IconDefinition
  component?: Component,
  value: string,
  tooltip?: string,
  className?: string
}
