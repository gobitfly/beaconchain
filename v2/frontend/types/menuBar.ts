import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'

export interface MenuBarButton {
  active?: boolean,
  class?: string,
  command?: () => void,
  component?: Component,
  disabledTooltip?: string,
  faIcon?: IconDefinition,
  highlight?: boolean,
  label?: string,
  route?: string,
}

export interface MenuBarEntry extends MenuBarButton {
  dropdown: boolean,
  items?: MenuBarButton[],
}
