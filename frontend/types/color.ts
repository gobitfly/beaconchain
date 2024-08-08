export type ColorBlindMode = 'none' | 'red-green' | 'blue-yellow' | 'complete'
export type ColorDefinition = { color: string, identifier: string }

export type ColorBlindModeProvider = {
  colorBlindMode: ComputedRef<ColorBlindMode>
  convertColors: (colors: ColorDefinition[], mode?: ColorBlindMode) => ColorDefinition[]
  setColorBlindMode: (value: ColorBlindMode) => void
}
