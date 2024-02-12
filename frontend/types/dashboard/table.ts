export type TableValue = {
  percent?: number,
  validatorList?: string[],
  efficiency?: {
    successCount: number,
    failedCount: number,
    rewards?: string,
    penalties?: string
  },
  inclusionDistance?: number,
  className?: string,
}
