export interface MenuBarButton {
  label: string;
  command?: () => void;
  route?: string;
  class?: string;
  component?: Component;
  active?: boolean;
  disabledTooltip?: string;
}

export interface MenuBarEntry extends MenuBarButton {
  dropdown: boolean;
  items?: MenuBarButton[];
}
