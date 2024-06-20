<script lang="ts" setup>
type Channel = 0 | 1 | 2
const R: Channel = 0
const G: Channel = 1
const B: Channel = 2
type Order = Channel[]
const contributionToI = [0.3, 0.6, 0.1]

enum CS { RGBlinear, RGBgamma, EyePercI, EyeNormI }

/** Classic color space in two variants (depending on the parameter given to the constructor) :
 *  - either the values are between 0-1 and linear with respect to light intensity,
 *  - or the values are between 0-255 and include a gamma (the standard way to store images). */
class RGB {
  readonly space: CS
  readonly chans: number[]

  /** @param valuesAreLinear if `true` is given, the values will be between 0-1 and linear with respect to light intensity,
   *  otherwise the values will be between 0-255 and include a gamma (the standard way to store images). */
  constructor (valuesAreLinear: boolean) {
    this.space = valuesAreLinear ? CS.RGBlinear : CS.RGBgamma
    this.chans = [0, 0, 0]
  }

  /** Copy color from another RGB or Eye object, or even from a regular array of 3 values.
   * Color spaces are automatically converted if they are different.
   * However, when an array of numbers is given, its values are expected to be compatible with this instance of RGB (either linear or gamma-shaped). */
  import (from: number[] | RGB | Eye) : void {
    if (Array.isArray(from) || from.space === this.space) {
      if (!Array.isArray(from)) {
        from = (from as RGB).chans
      }
      for (let i = R; i < B; i++) {
        this.chans[i] = from[i]
      }
    } else {
      switch (from.space) {
        case CS.RGBlinear :
          for (let i = R; i < B; i++) {
            this.chans[i] = Math.round(((from as RGB).chans[i] ** 0.454545) * 255)
          }
          break
        case CS.RGBgamma :
          for (let i = R; i < B; i++) {
            this.chans[i] = ((from as RGB).chans[i] / 255) ** 2.2
          }
          break
        case CS.EyePercI :
        case CS.EyeNormI :
          from.export(this)
          break
      }
    }
  }

  /** Copy the color of this object to another RGB or Eye object.
   * Color spaces are automatically converted if they are different. */
  export (to: RGB | Eye) : void {
    to.import(this)
  }

  limit () : void {
    const max = (this.space === CS.RGBlinear) ? 1 : 255
    for (let i = R; i < B; i++) {
      if (this.chans[i] < 0) { this.chans[i] = 0 }
      if (this.chans[i] > max) { this.chans[i] = max }
    }
  }
}

/** Color space supposedly close to human perception, in two variants (depending on the parameter given to the constructor) :
 * - either the intensity of the light is stored in `i` and follows what a human eye perceives,
 * - or it is stored in `j` and is normalized so it can take any value between 0 and 1. */
class Eye {
  readonly space: CS
  /** Perceived wavelength indicating where the color is on the rainbow. Key values: 0 is pure red. 1/3 is pure green. 2/3 is pure blue. 1 is pure red again. */
  w: number
  /** Perceived purity indicating how much light not contributing to the perceived wavelength is present. */
  p: number
  /** Perceived intensity of the light, so not normalized (given `w` and `p`, the maximum perceived intensity that can be reached with `w` and `p` is often less than 1).
   * Property `iMax` tells the maximum value that `i` can take for the current values of `w` and `p`.
   * `i` is considered only if `true` has been passed to the constructor, otherwise its value is undefined and giving it a value has no effect. */
  i: number
  /** Normalized intensity of the light, any value between 0 and 1 is possible.
   * `j` is considered only if `false` has been passed to the constructor, otherwise its value is undefined and giving it a value has no effect. */
  j: number
  /** Maximum value that `i` can have under the constraint set by `w` and `p`.
   * This value is kept up-to-date automatically. */
  get iMax () : number {
    if (this.Imax.wOfValue !== this.w || this.Imax.pOfValue !== this.p) {
      this.snapshotImax(this.i / rgb[z])
    }
    return this.Imax.value
  }

  private Imax = {
    value: 0,
    wOfValue: 0,
    pOfValue: 0
  }

  protected static rgbLinear = new RGB(true)

  import (from: RGB | Eye) : void {
    if (from.space === CS.RGBgamma) {
      Eye.rgbLinear.import(from as RGB)
      from = Eye.rgbLinear
    }
    if (from.space === CS.RGBlinear) {
      const rgb = (from as RGB).chans
      const [a, b, c] = Eye.orderRGB(rgb)
      if (rgb[c] <= 0) {
        // the highest channel has a value of 0 so the color is black
        this.w = this.p = this.i = this.j = 0
        this.snapshotImax(0)
        return
      } else if (rgb[a] >= 1) {
        // the lowest channel has a value of 1 so the color is white
        this.w = this.p = 0
        this.i = this.j = 1
        this.snapshotImax(1)
        return
      }
      const sumOfDominants = rgb[b] + rgb[c] // sum of the two channels surrounding the color
      const { anchor, left } = Eye.spectralPosition(a)
      this.w = (rgb[anchor] / sumOfDominants + left) / 3
      this.p = 1 - (2 * rgb[a]) / sumOfDominants
      this.j = rgb[c] // due to our definitions of w and p, this value turns out to be the ratio between I (no matter how I is calculated) and the maximum I that could be reached with w and p constant
      const [, y, z] = Eye.orderRGB([contributionToI[R] * rgb[R], contributionToI[G] * rgb[G], contributionToI[B] * rgb[B]])
      this.i = contributionToI[y] * rgb[y] + contributionToI[z] * rgb[z]
      this.snapshotImax(this.i / rgb[z])
      return
    }
    from = from as Eye // for the static checker
    if (from.space === this.space) {
      this.w = from.w
      this.p = from.p
      this.i = from.i
      this.j = from.j
      this.snapshotImax(from.Imax.value)
    } else {
      // either we convert  EyePercI into EyeNormI  or  EyeNormI into EyePercI
    }
  }

  export (to: RGB | Eye) : void {
    if (to.space === CS.EyePercI || to.space === CS.EyeNormI) {
      to.import(this)
    } else {
      to = to as RGB // for the static checker
      if (to.space === CS.RGBgamma) {
        this.export(Eye.rgbLinear)
        to.import(Eye.rgbLinear)
      } else {
        // convert Eye into RGBlinear
      }
    }
  }

  protected snapshotImax (iMax: number) : void {
    this.Imax.value = iMax
    this.Imax.pOfValue = this.p
    this.Imax.wOfValue = this.w
  }

  protected static orderRGB (rgb : number[]) : Order {
    if (rgb[R] < rgb[G]) {
      if (rgb[G] < rgb[B]) {
        return [R, G, B]
      }
      return (rgb[R] < rgb[B]) ? [R, B, G] : [B, R, G]
    }
    if (rgb[R] < rgb[B]) {
      return [G, R, B]
    }
    return (rgb[G] < rgb[B]) ? [G, B, R] : [B, G, R]
  }

  protected static spectralPosition (weakestChan: Channel) {
    switch (weakestChan) {
      case R : // the dominant channels are G and B
        return { anchor: B, left: 1 }
      case G : // the dominant channels are B and R
        return { anchor: R, left: 2 }
      case B : // the dominant channels are R and G
        return { anchor: G, left: 0 }
    }
    // impossible but the static checker believes it can happen:
    return { anchor: R, left: 0 }
  }

  protected static wToOrder (w : number) : Order {
    if (w < 1 / 3) {
      return (w < 1 / 3 - 1 / 6) ? [B, G, R] : [B, R, G]
    }
    if (w < 2 / 3) {
      return (w < 2 / 3 - 1 / 6) ? [R, B, G] : [R, G, B]
    }
    return (w < 3 / 3 - 1 / 6) ? [G, R, B] : [G, B, R]
  }

  /** @param intensityAsPerceived if `true` is given, the intensity of the light will be stored in `i` and follow what a human eye perceives,
   * otherwise it will be stored in `j` and normalized so it can take any value between 0 and 1. */
  constructor (intensityAsPerceived: boolean) {
    this.space = intensityAsPerceived ? CS.EyePercI : CS.EyeNormI
    this.w = this.p = this.i = this.j = 0
  }
}
</script>

<template>
  <div />
</template>

<style lang="scss" scoped>
</style>
