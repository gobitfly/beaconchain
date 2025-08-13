export type Query = {
  cursor?: string,
  group_id?: number,
  limit?: Limit,
  // duty?: 'none' | 'proposal' | 'slashed' | 'sync',
  period?:
    | 'all_time'
    | 'last_1h'
    | 'last_7d'
    | 'last_24h'
    | 'last_30d',
  search?: string,
  sort?: `${string}:${'asc' | 'desc'}`,
}

export const limits = [
  5,
  10,
  25,
  50,
  100,
] as const

export type Limit = (typeof limits)[number]

export const useDefaultQuery = (query?: Query) => {
  return ref<Query>({
    limit: 25,
    ...query,
  })
}
