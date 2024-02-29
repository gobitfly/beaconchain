export function getSummaryChartGroupColors (theme: string) {
  const colorsLight = ['#E7416A', '#6CF0F0', '#B2DF27', '#5D78DC', '#FFDB58', '#F067E9', '#57BD64', '#A448C0', '#DC2A7F', '#F58E45', '#87CEEB', '#438D61', '#E6BEFF', '#6BE4D8', '#FABEBE', '#90D9A5', '#FF6A00', '#FFBE7C', '#BCB997', '#DEB244', '#DDA0DD', '#FA8072', '#D2B48C', '#6B8E23', '#0E8686', '#9A6324', '#932929', '#808000', '#30308E', '#708090']
  const colorsDark = ['#E7416A', '#6CF0F0', '#C3F529', '#5D78DC', '#FFDB58', '#F067E9', '#57BD64', '#A448C0', '#DC2A7F', '#F58E45', '#87CEEB', '#438D61', '#E6BEFF', '#6BE4D8', '#FABEBE', '#AAFFC3', '#FF6A00', '#FFD8B1', '#FFFAC8', '#DEB244', '#DDA0DD', '#FA8072', '#D2B48C', '#6B8E23', '#0E8686', '#9A6324', '#932929', '#808000', '#30308E', '#708090']

  return theme === 'light' ? colorsLight : colorsDark
}

export function getSummaryChartTextColor (theme: string) {
  const styles = window.getComputedStyle(document.documentElement)

  if (theme === 'light') {
    return styles.getPropertyValue('--light-black')
  } else {
    return styles.getPropertyValue('--light-grey')
  }
}
