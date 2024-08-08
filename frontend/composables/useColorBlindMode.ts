import { inject } from 'vue'
import type { ColorBlindMode, ColorDefinition, ColorBlindModeProvider } from '~/types/color'

export function useColorBlindMode() {
  const provider = inject<ColorBlindModeProvider>('color-blind-mode')

  if (!provider) {
    throw new Error('useColorBlindMode must be in a child of useColorBlindModeProvider')
  }

  const convertColors = (colors: ColorDefinition[], mode?: ColorBlindMode): ColorDefinition[] => {
    return provider.convertColors(colors, mode)
  }

  const setColorBlindMode = (mode: ColorBlindMode): void => {
    return provider.setColorBlindMode(mode)
  }

  const colorBlindMode = computed<ColorBlindMode>(() => provider.colorBlindMode.value ?? 'none')

  return { colorBlindMode, convertColors, setColorBlindMode }
}
