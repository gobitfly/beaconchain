import type { ExtendedLabel } from '~/types/value'

export type OverviewTableData = {
  label: string
  addValidatorModal?: boolean
  value?: ExtendedLabel
  additonalValues?: ExtendedLabel[][]
  infos?: {
    label: string
    value: string | number
  }[]
}
