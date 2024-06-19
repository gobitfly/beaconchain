<script lang="ts" setup>
type Permutation = (0|1|2)[]

class RGB {
  chans: number[]

  protected format: 1 | 8

  constructor (from: RGB | undefined) {
    if (!from) {
      this.chans = [0, 0, 0]
      this.format = 1
      return
    }
    this.chans = [...from.chans]
    this.format = from.format
  }

  fillRGB1fromRGB8 (rgb8: number[] | RGB) {
    if (!Array.isArray(rgb8)) {
      rgb8 = rgb8.chans
    }
    this.chans[0] = (rgb8[0] / 255) ** 2.2
    this.chans[1] = (rgb8[1] / 255) ** 2.2
    this.chans[2] = (rgb8[2] / 255) ** 2.2
    this.format = 1
  }

  fillRGB8fromRGB1 (rgb1: number[] | RGB) {
    if (!Array.isArray(rgb1)) {
      rgb1 = rgb1.chans
    }
    this.chans[0] = (rgb1[0] ** 0.454545) * 255
    this.chans[1] = (rgb1[1] ** 0.454545) * 255
    this.chans[2] = (rgb1[2] ** 0.454545) * 255
    this.format = 8
  }

  limit () {
    limit(this.chans, 0, (this.format === 1) ? 1 : 255)
  }
}

/** color space supposedly close to human perception */
class Eye {
  /** Perceived wavelength indicating where the color is on the rainbow. Key values: 0 is pure red. 1/3 is pure green. 2/3 is pure blue. 1 is pure red again.
   * Important: After changing the value of `w` or `p`, you must always set `i` or `j`. */
  w: number
  /** Perceived purity indicating how much light not contributing to the perceived wavelength is present.
   * Important: After changing the value of `w` or `p`, you must always set `i` or `j`. */
  p: number
  /** Perceived intensity of the light, so not normalized (given `w` and `p`, the maximum perceived intensity that can be reached with `w` and `p` is often less than 1).
   * Changing this value updates `j` accordingly. */
  get i () : number { return this.I }
  /** Normalized intensity of the light, any value between 0 and 1 is possible.
   * Changing this value updates `i` accordingly. */
  get j () : number { return this.J }
  /** Maximum value that `i` can have under the constraint set by `w` and `p`.
   * Updated when they change. */
  get iMax () : number { return this.Imax }

  /** Read `iMax` to know the maximum value that `i` can take for the current values of `w` and `p`. */
  set i (val: number) {
    this.I = val
  }

  set j (val: number) {
    this.J = val
  }

  static channelIcontribution = [0.3, 0.6, 0.1]
  private I: number
  private J: number
  private Imax: number

  fillFromRGB1 (rgb1: RGB) {
    const rgb = rgb1.chans
    const [a, b, c] = order(rgb)
    if (rgb[c] <= 0) {
      // the highest channel is 0 so the color is black
      this.w = this.p = this.I = this.J = this.Imax = 0
      return
    } else if (rgb[a] >= 1) {
      // the lowest channel is 1 so the color is white
      this.w = this.p = 0
      this.I = this.J = this.Imax = 1
      return
    }
    const sumOfDominants = rgb[b] + rgb[c] // sum of the two channels surrounding the color
    switch (a) { // a is the channel with the lowest value
      case 0 : // the dominant channels are G and B
        this.w = (1 + rgb[2] / sumOfDominants) / 3
        break
      case 1 : // the dominant channels are R and B
        this.w = (2 + rgb[0] / sumOfDominants) / 3
        break
      case 2 : // the dominant channels are R and G
        this.w = (0 + rgb[1] / sumOfDominants) / 3
        break
    }
    this.p = 1 - (2 * rgb[a]) / sumOfDominants
    this.J = rgb[c] // due to our definitions of w and p, this value turns out to be the ratio between I (no matter how I is calculated) and the maximum I that could be reached with w and p constant
    const [x, y, z] = order([Eye.channelIcontribution[0] * rgb[0], Eye.channelIcontribution[1] * rgb[1], Eye.channelIcontribution[2] * rgb[2]])
    this.I = Eye.channelIcontribution[y] * rgb[y] + Eye.channelIcontribution[z] * rgb[z]
    this.Imax = Eye.channelIcontribution[y] * rgb[y] / rgb[z] + Eye.channelIcontribution[z]
  }

  exportToRGB1 () : RGB {
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

  constructor () {
    this.w = this.p = this.I = this.J = this.Imax = 0
  }
}

function order (p : number[]) : Permutation {
  let a = 0 as Permutation[number]
  let b = 1 as Permutation[number]
  const c = 2 as Permutation[number]
  if (p[b] < p[a]) { [a, b] = [b, a] }
  if (p[c] < p[a]) { return [c, a, b] }
  if (p[c] < p[b]) { return [a, c, b] }
  return [a, b, c]
}

function limit (p : number[], min: number, max: number) {
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
