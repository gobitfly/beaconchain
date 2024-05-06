export const HeatmapTimeFrames = ['24h', '7d', '30d', '365d'] as const
export type HeatmapTimeFrame = typeof HeatmapTimeFrames[number]
