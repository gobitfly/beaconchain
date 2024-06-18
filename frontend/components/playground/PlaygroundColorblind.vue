<script lang="ts" setup>
import type { Order } from '@stripe/stripe-js'

type Point = number[]

type RGB = number[]

type SPI = { // simple color space supposedly close to human perception
  s: number, // spectrum (similar to the notion of hue)
  p: number, // purity (similar to the notion of saturation)
  i: number // intensity (similar to the notion of value)
}

type Permutation = (0|1|2)[]

function RGB24toRGB01 (rgb: RGB) : RGB {
  return [
    (rgb[0] / 255) ** 2.2,
    (rgb[1] / 255) ** 2.2,
    (rgb[2] / 255) ** 2.2
  ]
}

function RGB01toRGB24 (rgb: RGB) : RGB {
  return [
    (rgb[0] ** 0.454545) * 255,
    (rgb[1] ** 0.454545) * 255,
    (rgb[2] ** 0.454545) * 255
  ]
}

function RGB01toSPI (rgb: RGB) : SPI {
  const w = [0.3, 0.6, 0.1]
  const [a, b, c] = order(rgb)
  if (rgb[c] <= 0) {
    // the highest channel is 0 so the color is black
    return { s: 0, p: 0, i: 0 }
  } else if (rgb[a] >= 1) {
    // the lowest channel is 1 so the color is white
    return { s: 0, p: 0, i: 1 }
  }
  const sumOfBounds = rgb[b] + rgb[c]  // sum of the two channels surrounding the color
  let s : number
  switch (a) { // a is the channel with the lowest value
    case 0 : // the dominant channels are G and B
      s = (1 + rgb[2] / sumOfBounds) / 3
      break
    case 1 : // the dominant channels are R and B
      s = (2 + rgb[0] / sumOfBounds) / 3
      break
    case 2 : // the dominant channels are R and G
      s = (0 + rgb[1] / sumOfBounds) / 3
      break
  }
  const p = 1 - (2 * rgb[a]) / (rgb[b] + rgb[c])
  const i = rgb[c] // this might look surprising but this value is the ratio between the absolute intensity (no matter how it is calculated, for example rgb[b] + rgb[c] or the sum of all channels or a weighted sum, etc) and the maximum absolute intensity that can be reached with s and p (so along their invariance line)
  return { s, p, i }
}

function SPItoRGB01 (spi: SPI) : RGB {
  if (spi.s <= 1 / 3) {
    // the dominant channels are R and G
    const factor = 1 / 
    const G = 
    return [
      ,
      ,
    ]
  } else
    if (spi.s <= 2 / 3) {
      // the dominant channels are G and B
      const factor = 1 / 
      const G = 
      return [
        ,
        ,
      ]
    } else {
      // the dominant channels are R and B
      const factor = 1 / 
      const G = 
      return [
        ,
        ,
      ]
    }
}

function order (p : Point) : Permutation {
  let a = 0 as Permutation[number]
  let b = 1 as Permutation[number]
  const c = 2 as Permutation[number]
  if (p[b] < p[a]) { [a, b] = [b, a] }
  if (p[c] < p[a]) { return [c, a, b] }
  if (p[c] < p[b]) { return [a, c, b] }
  return [a, b, c]
}

function limit (p : Point, min: number, max: number) {
  for (let i = 0; i < 3; i++) {
    if (p[i] < min) { p[i] = min }
    if (p[i] > max) { p[i] = max }
  }
}
</script>

<template>
  <div>
    {{ order([33,55,88]) }} - 0,1,2 <br>
    {{ order([55,33,88]) }} - 1,0,2 <br>
    .
    {{ order([33,88,55]) }} - 0,2,1 <br>
    {{ order([55,88,33]) }} - 2,0,1 <br>
    .
    {{ order([88,33,55]) }} - 1,2,0 <br>
    {{ order([88,55,33]) }} - 2,1,0 <br>
  </div>
</template>

<style lang="scss" scoped>
</style>
