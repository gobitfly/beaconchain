import type { ExtendedLabel } from '~/types/value'

export type OverviewTableData = {
  additonalValues?: ExtendedLabel[][],
  addValidatorModal?: boolean,
  infos?: {
    label: string,
    value: number | string,
  }[],
  label: string,
  value?: ExtendedLabel,
}
