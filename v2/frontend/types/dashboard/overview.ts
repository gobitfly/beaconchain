import type { ExtendedLabel } from '~/types/value'

export type OverviewTableData = {
  label: string,
  value?: ExtendedLabel,
  additonalValues?: ExtendedLabel[][],
  infos?: {
    label: string,
    value: string | number
  }[]
}
