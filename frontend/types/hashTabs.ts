import type { Component } from 'vue'
import type { Icon } from '~/components/bc/icon/BcIcon.vue'

export type HashTab = {
  component?: Component,
  disabled?: boolean,
  icon?: Icon,
  key: string,
  placeholder?: string,
  title?: string,
}
export type HashTabs = HashTab[]
