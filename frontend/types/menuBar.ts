import type { Icon } from '~/components/bc/icon/BcIcon.vue'

export interface MenuBarButton {
  active?: boolean,
  class?: string,
  command?: () => void,
  component?: Component,
  disabledTooltip?: string,
  faIcon?: Icon,
  highlight?: boolean,
  label?: string,
  route?: string,
}

export interface MenuBarEntry extends MenuBarButton {
  dropdown: boolean,
  items?: MenuBarButton[],
}
