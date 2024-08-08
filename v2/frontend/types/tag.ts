export const TagColors = [
  'success',
  'failed',
  'orphaned',
  'partial',
  'light',
  'dark',
] as const
export type TagColor = (typeof TagColors)[number]

export type TagSize = 'default' | 'compact' | 'circle'
