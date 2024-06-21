<script lang="ts" setup>
enum CS {
  /** RGB values between 0-1 and linear with respect to light intensity */
  RGBlinear,
  /** RGB values between 0-255 and including a gamma exponent of 2.2 (the standard way to store images) */
  RGBgamma,
  /** RPI values whose `i` follows the intensity of the light that a human eye perceives (for a given `j`, different `r` and `p` values produce the same intensity for the human eye, but some values of `j` do not exist for certain values of r` and `p`) */
  EyePercI,
  /** RPJ values whose `j` normalizes the light intensity (for a given `j`, different `r` and `p` values produce different intensities for the human eye, but this format is easier to handle because `j` is free to take any value between 0 and 1) */
  EyeNormJ
}

type Channel = 0 | 1 | 2
const R: Channel = 0
const G: Channel = 1
const B: Channel = 2
type Order = Channel[]
const contributionToI = [0.3, 0.6, 0.1] // rounded coefficients from the defintion of "relative luminance" in the ITU-R Recommendation BT.601

/** Classical color space in two variants (depending on the parameter given to the constructor) :
 *  - either the values are between 0-1 and linear with respect to light intensity,
 *  - or the values are between 0-255 and include a gamma exponent of 2.2 (the standard way to store images). */
class RGB {
  readonly space: CS
  /** Read and write individually the values of the color channels here. You cannot assign an array directly. To assign a whole array, use method `import()`. */
  readonly chans: number[]

  /** @param space if `CS.RGBlinear` is given, the values will be between 0-1 and linear with respect to light intensity;
   *  if `CS.RGBgamma` is given, the values will be between 0-255 and include a gamma exponent of 2.2 (the standard way to store images). */
  constructor (space: CS) {
    if (space !== CS.RGBlinear && space !== CS.RGBgamma) {
      throw new Error('a RGB object can carry RGB information only')
    }
    this.space = space
    this.chans = [0, 0, 0]
  }

  /** Copy a RGB or Eye instance into the current RGB instance. A regular array of 3 values can also be given (in this order: R,G,B).
   * For RGB and Eye objects, color spaces are automatically converted into the color space of the current RGB instance if they differ.
   * If a regular array of numbers is given, its values are expected to be compatible with the color space of the current instance (either linear RGB or gamma-shaped RGB).
   * @returns the instance that you import into (so not the parameter) */
  import (from: number[] | RGB | Eye) : RGB {
    if (Array.isArray(from) || from.space === this.space) {
      if (!Array.isArray(from)) {
        from = (from as RGB).chans
      }
      for (let i = R; i <= B; i++) {
        this.chans[i] = from[i]
      }
    } else {
      switch (from.space) {
        case CS.RGBlinear :
          for (let i = R; i <= B; i++) {
            this.chans[i] = Math.round(((from as RGB).chans[i] ** 0.454545) * 255)
          }
          break
        case CS.RGBgamma :
          for (let i = R; i <= B; i++) {
            this.chans[i] = ((from as RGB).chans[i] / 255) ** 2.2
          }
          break
        case CS.EyePercI :
        case CS.EyeNormJ :
          from.export(this)
          break
      }
    }
    return this
  }

  /** Copy this RGB instance into another RGB or Eye instance.
   * The color space of the current instance is automatically converted into the color space of the target instance if needed.
   * @param to existing instance to fill, or if the identifier of a color space is given instead of an object, `export` creates an instance for you, fills it and returns it.
   * @returns target instance (same as `to` if `to` was an instance) */
  export (to: RGB | Eye | CS) : RGB | Eye {
    if (typeof to !== 'object') {
      to = (to === CS.RGBlinear || to === CS.RGBgamma) ? new RGB(to) : new Eye(to)
    }
    to.import(this)
    return to
  }

  /** corrects channel values that are not within the limits of the format (0-1 or 0-255) */
  limit () : void {
    const max = (this.space === CS.RGBlinear) ? 1 : 255
    for (let i = R; i <= B; i++) {
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
  /** Read and write the perceived rainbow color here. It indicates where the color is on the rainbow. Key values: 0 is red. 1/3 is green. 2/3 is blue. 1 is red again. */
  r: number
  /** Read and write the perceived purity here. It indicates how much light not contributing to the perceived rainbow color is present. */
  p: number
  /** Read and write the perceived intensity of the light here. It is not normalized because `r` and `p` constrain the maximum intensity that can be perceived and it is often less than 1.
   * If needed, property `iMax` tells the maximum value that `i` can hold for the current values of `r` and `p`.
   * `i` makes sense only if `CS.EyePercI` has been passed to the constructor, otherwise its value is undefined and setting a value has no effect. */
  i: number
  /** Read and write the normalized intensity of the light here. Any value between 0 and 1 is possible.
   * `j` makes sense only if `CS.EyeNormJ` has been passed to the constructor, otherwise its value is undefined and setting a value has no effect. */
  j: number
  /** Maximum value that `i` can have under the constraint of `r` and `p`.
   * This value is kept up-to-date automatically. */
  get iMax () : number {
    if (this.Imax.wOfValue !== this.r || this.Imax.pOfValue !== this.p) {
      if (this.space !== CS.EyePercI) {
        return 0
      }
      this.export(Eye.rgbLinear)
      const z = Eye.RGBtoHighestChannel(Eye.rgbLinear.chans)
      this.snapshotImax(this.i / Eye.rgbLinear.chans[z])
    }
    return this.Imax.value
  }

  /** @param space if `CS.EyePercI` is given, the intensity of the light will be stored in `i` and follow what a human eye perceives;
 * if `CS.EyeNormJ` is given, the intensity will be stored in `j` and normalized so it can take any value between 0 and 1. */
  constructor (space: CS) {
    if (space !== CS.EyePercI && space !== CS.EyeNormJ) {
      throw new Error('an Eye object can carry RPI/J information only')
    }
    this.space = space
    this.r = this.p = this.i = this.j = 0
  }

  private Imax = {
    value: 0,
    wOfValue: 0,
    pOfValue: 0
  }

  protected static rgbLinear = new RGB(CS.RGBlinear)

  /** Copy a RGB or Eye instance into the current Eye instance.
   * Color spaces are automatically converted into the color space of the current Eye instance if they differ.
   * @returns the instance that you import into (so not the parameter) */
  import (from: RGB | Eye) : Eye {
    if (from.space === CS.RGBgamma) {
      from = Eye.rgbLinear.import(from)
    }
    if (from.space === CS.RGBlinear) {
      // conversion of RGB into Eye
      const rgb = (from as RGB).chans
      const l = Eye.RGBtoLowestChannel(rgb)
      if (rgb[R] + rgb[G] + rgb[B] <= 0.002) {
        // the color is black
        this.r = 0.5
        this.p = this.i = this.j = 0
        this.snapshotImax(0)
      } else if (rgb[l] >= 1) {
        // the lowest channel has a value of 1 so the color is white
        this.r = 0.5
        this.p = 0
        this.i = this.j = 1
        this.snapshotImax(1)
      } else {
        const [h1, h2] = Eye.RGBtoAnchors(l)
        const sumOfAnchors = rgb[h1] + rgb[h2]
        this.r = ((rgb[h2] - rgb[l]) / (sumOfAnchors - 2 * rgb[l]) + h1) / 3
        this.p = 1 - (2 * rgb[l]) / sumOfAnchors
        this.fillIntensityFromRGB(rgb)
      }
    } else {
      from = from as Eye // for the static checker
      this.r = from.r
      this.p = from.p
      if (from.space === this.space) {
      // we copy an Eye into our Eye of the same variant
        this.i = from.i
        this.j = from.j
        this.snapshotImax(from.Imax.value)
      } else {
      // we must convert an Eye variant into our variant (EyePercI <-> EyeNormJ)
        from.export(Eye.rgbLinear)
        this.fillIntensityFromRGB(Eye.rgbLinear.chans)
      }
    }
    return this
  }

  /** Copy this Eye instance into another RGB or Eye instance.
   * The color space of the current instance is automatically converted into the color space of the target instance if needed.
   * @param to existing instance to fill, or if the identifier of a color space is given instead of an object, `export` creates an instance for you, fills it and returns it.
   * @returns target instance (same as `to` if `to` was an instance) */
  export (to: RGB | Eye | CS) : RGB | Eye {
    if (typeof to !== 'object') {
      to = (to === CS.RGBlinear || to === CS.RGBgamma) ? new RGB(to) : new Eye(to)
    }
    if (to.space === CS.EyePercI || to.space === CS.EyeNormJ) {
      to.import(this)
    } else {
      to = to as RGB // for the static checker
      if (to.space === CS.RGBgamma) {
        this.export(Eye.rgbLinear).export(to)
      } else {
        if (this.r < 0.001) { this.r = 1 }
        const [l, m, h] = Eye.rToChannelOrder(this.r)
        const [h1, h2] = Eye.RGBtoAnchors(l)
        const ratio = [0, 0, 0]
        const q = this.p / (1 - this.p)
        ratio[l] = 1
        ratio[h2] = 2 * (3 * this.r - h1) * q + 1
        ratio[h1] = 2 * q - ratio[h2] + 2
        if (this.space === CS.EyePercI) {
          const pIr = [contributionToI[R] * ratio[R], contributionToI[G] * ratio[G], contributionToI[B] * ratio[B]]
          const [y, z] = Eye.RGBtoAnchorOrder(pIr)
          to.chans[l] = this.i / (pIr[y] + pIr[z])
          to.chans[h] = to.chans[l] * ratio[h]
        } else { // (this.space === CS.EyeNormJ)
          to.chans[h] = this.j
          to.chans[l] = to.chans[h] / ratio[h]
        }
        to.chans[m] = to.chans[l] * ratio[m]
      }
    }
    return to
  }

  protected fillIntensityFromRGB (rgb : number[]) : void {
    if (this.space === CS.EyePercI) {
      const pI = [contributionToI[R] * rgb[R], contributionToI[G] * rgb[G], contributionToI[B] * rgb[B]]
      const [m, h] = Eye.RGBtoAnchorOrder(pI)
      this.i = pI[m] + pI[h]
      this.snapshotImax(this.i / rgb[h])
    } else {
      // due to our definitions of r and p, this value turns out to be the ratio between i (no matter how i is calculated) and the maximum i that could be reached with r and p constant
      const h = Eye.RGBtoHighestChannel(rgb)
      this.j = rgb[h]
    }
  }

  protected snapshotImax (iMax: number) : void {
    this.Imax.value = iMax
    this.Imax.pOfValue = this.p
    this.Imax.wOfValue = this.r
  }

  /** @returns the channel carrying the lowest value */
  protected static RGBtoLowestChannel (rgb : number[]) : Channel {
    if (rgb[R] < rgb[G]) {
      if (rgb[R] < rgb[B]) { return R }
    } else
      if (rgb[G] < rgb[B]) { return G }
    return B
  }

  /** @returns the channel carrying the highest value */
  protected static RGBtoHighestChannel (rgb : number[]) : Channel {
    if (rgb[R] > rgb[G]) {
      if (rgb[R] > rgb[B]) { return R }
    } else
      if (rgb[G] > rgb[B]) { return G }
    return B
  }

  /** @returns the anchors in the same order as on the rainbow (note that R is both before G and after B) */
  protected static RGBtoAnchors (lowestChan: Channel) : Order {
    switch (lowestChan) {
      case R : return [G, B]
      case G : return [B, R]
      case B : return [R, G]
    }
    return [] // impossible but the static checker believes it can happen
  }

  /** @returns the anchor having the lowest value followed by the highest-value anchor */
  protected static RGBtoAnchorOrder (rgb : number[]) : Order {
    if (rgb[R] < rgb[G]) {
      if (rgb[G] < rgb[B]) { return [G, B] }
      return (rgb[R] < rgb[B]) ? [B, G] : [R, G]
    }
    if (rgb[R] < rgb[B]) { return [R, B] }
    return (rgb[G] < rgb[B]) ? [B, R] : [G, R]
  }

  /** @returns the order of the channels from the lowest value to the highest-value */
  protected static rToChannelOrder (r : number) : Order {
    if (r < 1 / 3) {
      return (r < 1 / 3 - 1 / 6) ? [B, G, R] : [B, R, G]
    }
    if (r < 2 / 3) {
      return (r < 2 / 3 - 1 / 6) ? [R, B, G] : [R, G, B]
    }
    return (r < 3 / 3 - 1 / 6) ? [G, R, B] : [G, B, R]
  }
}

const cons = console

const x = new RGB(CS.RGBgamma)
x.import([255, 255, 255])

const X = x.export(CS.EyeNormJ)
const Y = x.export(CS.EyePercI)

cons.log(X)
cons.log(X.export(CS.RGBgamma))
cons.log(Y)
cons.log(Y.export(CS.RGBgamma))
</script>

<template>
  <div />
</template>

<style lang="scss" scoped>
</style>
