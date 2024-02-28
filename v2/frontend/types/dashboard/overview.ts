import type { ExtendedLabel } from '~/types/value'

export type OverviewTableData = {
  label: string,
  value?: ExtendedLabel,
  additonalValues?: ExtendedLabel[][],
}
