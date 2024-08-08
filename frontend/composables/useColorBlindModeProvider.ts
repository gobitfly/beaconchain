import { provide } from 'vue'
import type { ColorBlindMode, ColorDefinition, ColorBlindModeProvider } from '~/types/color'

export function useColorBlindModeProvider() {
  const { setting, changeSetting } = useGlobalSetting<ColorBlindMode>('color-blind-mode')

  const colorBlindMode = computed(() => setting.value || 'none')

  function getRandomColor() {
    const letters = '0123456789ABCDEF'
    let color = '#'
    for (let i = 0; i < 6; i++) {
      color += letters[Math.floor(Math.random() * 16)]
    }
    return color
  }

  const convertColors = (colors: ColorDefinition[], mode?: ColorBlindMode): ColorDefinition[] => {
    mode = mode ?? colorBlindMode.value
    if (mode === 'none') {
      return colors
    }
    // TODO: Replace with real color conversion
    return colors.map(c => ({
      ...c,
      color: getRandomColor(),
    }))
  }

  watch(setting, (mode) => {
    if (!isServer) {
      const appColors = convertColors(APP_COLORS, mode)
      console.log('appColors', appColors)
      appColors.forEach((c) => {
        document.documentElement.style.setProperty(c.identifier, c.color)
      })
    }
  }, { immediate: true })

  provide<ColorBlindModeProvider>('color-blind-mode', { colorBlindMode, convertColors, setColorBlindMode: changeSetting })
}
