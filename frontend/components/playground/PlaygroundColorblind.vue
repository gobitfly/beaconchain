<script lang="ts" setup>
type Channel = 0 | 1 | 2
const R: Channel = 0
const G: Channel = 1
const B: Channel = 2
type Order = Channel[]
const contributionToI = [0.3, 0.6, 0.1] // rounded coefficients from the defintion of "relative luminance" in the ITU-R Recommendation BT.601

enum CS { RGBlinear, RGBgamma, EyePercI, EyeNormJ }

/** Classical color space in two variants (depending on the parameter given to the constructor) :
 *  - either the values are between 0-1 and linear with respect to light intensity,
 *  - or the values are between 0-255 and include a gamma (the standard way to store images). */
class RGB {
  readonly space: CS
  /** read and write the values of the color channels here */
  readonly chans: number[]

  /** @param space if `CS.RGBlinear` is given, the values will be between 0-1 and linear with respect to light intensity;
   *  if `CS.RGBgamma` is given, the values will be between 0-255 and include a gamma (the standard way to store images). */
  constructor (space: CS) {
    if (space !== CS.RGBlinear && space !== CS.RGBgamma) {
      throw new Error('a RGB object can carry RGB information only')
    }
    this.space = space
    this.chans = [0, 0, 0]
  }

  /** Copy a RGB or Eye instance into the current RGB instance. A regular array of 3 values can also be given (in this order: R,G,B).
   * For RGB and Eye objects, color spaces are automatically converted into the color space of the current RGB instance if they differ.
   * If a regular array of numbers is given, its values are expected to be compatible with the color space of the current instance (either linear RGB or gamma-shaped RGB). */
  import (from: number[] | RGB | Eye) : RGB {
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
  /** Read and write the perceived wavelength here. It indicates where the color is on the rainbow. Key values: 0 is red. 1/3 is green. 2/3 is blue. 1 is red again. */
  w: number
  /** Read and write the perceived purity here. It indicates how much light not contributing to the perceived wavelength is present. */
  p: number
  /** Read and write the perceived intensity of the light here. It is not normalized because `w` and `p` constrain the maximum intensity that can be perceived and it is often less than 1.
   * If needed, property `iMax` tells the maximum value that `i` can hold for the current values of `w` and `p`.
   * `i` makes sense only if `CS.EyePercI` has been passed to the constructor, otherwise its value is undefined and setting a value has no effect. */
  i: number
  /** Read and write the normalized intensity of the light here. Any value between 0 and 1 is possible.
   * `j` makes sense only if `CS.EyeNormJ` has been passed to the constructor, otherwise its value is undefined and setting a value has no effect. */
  j: number
  /** Maximum value that `i` can have under the constraint of `w` and `p`.
   * This value is kept up-to-date automatically. */
  get iMax () : number {
    if (this.Imax.wOfValue !== this.w || this.Imax.pOfValue !== this.p) {
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
      throw new Error('an Eye object can carry WPI/J information only')
    }
    this.space = space
    this.w = this.p = this.i = this.j = 0
  }

  private Imax = {
    value: 0,
    wOfValue: 0,
    pOfValue: 0
  }

  protected static rgbLinear = new RGB(CS.RGBlinear)

  /** Copy a RGB or Eye instance into the current Eye instance.
   * Color spaces are automatically converted into the color space of the current Eye instance if they differ. */
  import (from: RGB | Eye) : Eye {
    if (from.space === CS.RGBgamma) {
      Eye.rgbLinear.import(from)
      from = Eye.rgbLinear
    }
    if (from.space === CS.RGBlinear) {
      // conversion of RGB into Eye
      const rgb = (from as RGB).chans
      const l = Eye.RGBtoLowestChannel(rgb)
      if (rgb[R] + rgb[G] + rgb[B] <= 0.002) {
        // the color is black
        this.w = this.p = this.i = this.j = 0
        this.snapshotImax(0)
        return this
      } else if (rgb[l] >= 1) {
        // the lowest channel has a value of 1 so the color is white
        this.w = this.p = 0
        this.i = this.j = 1
        this.snapshotImax(1)
        return this
      }
      const [h1, h2] = Eye.RGBtoAnchors(l)
      const sumOfAnchors = rgb[h1] + rgb[h2]
      this.w = (rgb[h2] / sumOfAnchors + h1) / 3
      this.p = 1 - (2 * rgb[l]) / sumOfAnchors
      this.fillIntensityFromRGB(rgb)
      return this
    }
    from = from as Eye // for the static checker
    this.w = from.w
    this.p = from.p
    if (from.space === this.space) {
      // we copy an Eye into our Eye of the same variant
      this.i = from.i
      this.j = from.j
      this.snapshotImax(from.Imax.value)
    } else {
      // we must convert an Eye variant into our variant (EyePercI <-> EyeNormJ)
      this.export(Eye.rgbLinear)
      this.fillIntensityFromRGB(Eye.rgbLinear.chans)
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
        this.export(Eye.rgbLinear)
        to.import(Eye.rgbLinear)
      } else {
        // convert Eye (EyePercI or EyeNormJ) into RGBlinear
        let sumOfAnchors: number
        const [l, m, h] = Eye.wToChannelOrder(this.w)
        const [h1, h2] = Eye.RGBtoAnchors(l)
        const ratio = [0, 0, 0]
        ratio[l] = (1 - this.p) / 2
        ratio[h2] = 3 * this.w - h1
        if (this.space === CS.EyePercI) {
          ratio[h1] = 1 - ratio[h2]
          const pIr = [contributionToI[R] * ratio[R], contributionToI[G] * ratio[G], contributionToI[B] * ratio[B]]
          const [y, z] = Eye.RGBtoAnchorOrder(pIr)
          sumOfAnchors = this.i / (pIr[y] + pIr[z])
          for (let i = R; i < B; i++) {
            to.chans[i] = ratio[i] * sumOfAnchors
          }
        } else { // (this.space === CS.EyeNormJ)
          to.chans[h] = this.j
          if (h2 === h) {
            sumOfAnchors = this.j / ratio[h2]
            to.chans[m] = sumOfAnchors - this.j
          } else {
            sumOfAnchors = this.j + to.chans[h2]
            to.chans[m] = this.j / (1 / ratio[h2] - 1)
          }
          to.chans[l] = sumOfAnchors * ratio[l]
        }
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
      // due to our definitions of w and p, this value turns out to be the ratio between i (no matter how i is calculated) and the maximum i that could be reached with w and p constant
      const h = Eye.RGBtoHighestChannel(rgb)
      this.j = rgb[h]
    }
  }

  protected snapshotImax (iMax: number) : void {
    this.Imax.value = iMax
    this.Imax.pOfValue = this.p
    this.Imax.wOfValue = this.w
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
  protected static wToChannelOrder (w : number) : Order {
    if (w < 1 / 3) {
      return (w < 1 / 3 - 1 / 6) ? [B, G, R] : [B, R, G]
    }
    if (w < 2 / 3) {
      return (w < 2 / 3 - 1 / 6) ? [R, B, G] : [R, G, B]
    }
    return (w < 3 / 3 - 1 / 6) ? [G, R, B] : [G, B, R]
  }
}
</script>

<template>
  <div />
</template>

<style lang="scss" scoped>
</style>
