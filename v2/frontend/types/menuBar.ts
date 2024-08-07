import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'

export interface MenuBarButton {
  label?: string;
  command?: () => void;
  route?: string;
  class?: string;
  highlight?: boolean;
  faIcon?: IconDefinition;
  component?: Component;
  active?: boolean;
  disabledTooltip?: string;
}

export interface MenuBarEntry extends MenuBarButton {
  dropdown: boolean;
  items?: MenuBarButton[];
}
