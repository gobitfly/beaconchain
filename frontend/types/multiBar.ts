import type { Icon } from '~/components/bc/icon/BcIcon.vue'

export type MultiBarItem = {
  className?: string,
  component?: Component,
  componentClass?: string,
  componentProps?: any,
  disabled?: boolean,
  icon?: Icon,
  tooltip?: string,
  value: string,
}
