export const HeatmapTimeFrames = ['last_24h', 'last_7d', 'last_30d', 'last_365d'] as const
export type HeatmapTimeFrame = typeof HeatmapTimeFrames[number]
