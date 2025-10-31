import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import type { Component } from 'vue'

export type HashTab = {
  component?: Component,
  disabled?: boolean,
  icon?: IconDefinition,
  key: string,
  placeholder?: string,
  title?: string,
}
export type HashTabs = HashTab[]
