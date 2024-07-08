<script lang="ts" setup>
enum CS {
  /** RGB values between 0-1 and linear with respect to light intensity */
  RGBlinear,
  /** RGB values between 0-255 and including a gamma exponent of 2.2 (the standard way to store images) */
  RGBgamma,
  /** RPI values (rainbow color, purity, intensity) whose `i` follows the intensity of the light that a human eye perceive (for a given `i`, all colors that you get with various `r` and `p` produce the same intensity for the human eye). Caution: some values of `i` are out-of-range for certain `r` and `p` values (meaning that they would correspond to RGB values greater than 255). */
  EyePercI,
  /** RPJ values (rainbow color, purity, intensity) whose `j` is equivalent to `i/iMax` in the `CS.EyePercI` variant, so `j` is free to take any value between 0 and 1 (so, for a given `j`, the colors that you get with various `r` and `p` produce different intensities for the human eye, whereas the `CS.EyePercI` variant ensures that `i` represents a constant perceived intensity whatever the color is) */
  EyeNormJ
}
type Channel = 0 | 1 | 2
const R: Channel = 0
const G: Channel = 1
const B: Channel = 2
const gamma = 2.2
const gammaInv = 1 / gamma

/** Classical color space in two variants (depending on the parameter given to the constructor) :
 *  - either the values are between 0-1 and linear with respect to light intensity,
 *  - or the values are between 0-255 and include a gamma exponent of 2.2 (the standard way to store images). */
class RGB {
  readonly space: CS
  /** Read and write individually the values of the primary channels here. You cannot assign an array directly. To assign a whole array, use method `import()`. */
  readonly chan: number[]

  /** @param space if `CS.RGBlinear` is given, the values will be between 0-1 and linear with respect to light intensity;
   *  if `CS.RGBgamma` is given, the values will be between 0-255 and include a gamma exponent of 2.2 (the standard way to store images). */
  constructor (space: CS) {
    if (space !== CS.RGBlinear && space !== CS.RGBgamma) {
      throw new Error('a RGB object can carry RGB information only')
    }
    this.space = space
    this.chan = [0, 0, 0]
  }

  /** Copies a RGB or Eye instance into the current RGB instance. A regular array of 3 values can also be given (in this order: R,G,B).
   * For RGB and Eye objects, color spaces are automatically converted into the color space of the current RGB instance if they differ.
   * If a regular array of numbers is given, its values are expected to be compatible with the color space of the current instance (either linear RGB or gamma-shaped RGB).
   * @returns the instance that you import into (so not the parameter) */
  import (from: number[] | RGB | Eye) : RGB {
    if (Array.isArray(from) || from.space === this.space) {
      if (!Array.isArray(from)) {
        from = (from as RGB).chan
      }
      for (let i = R; i <= B; i++) {
        this.chan[i] = from[i]
      }
    } else {
      switch (from.space) {
        case CS.RGBlinear :
          for (let i = R; i <= B; i++) {
            this.chan[i] = Math.round(((from as RGB).chan[i] ** gammaInv) * 255)
          }
          break
        case CS.RGBgamma :
          for (let i = R; i <= B; i++) {
            this.chan[i] = ((from as RGB).chan[i] / 255) ** gamma
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

  /** Copies this RGB instance into another RGB or Eye instance.
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
      if (this.chan[i] < 0) { this.chan[i] = 0 }
      if (this.chan[i] > max) { this.chan[i] = max }
    }
  }
}

/** Color space close to human perception, in two variants (depending on the parameter given to the constructor) :
 * - either the intensity of the light is stored in `i` and follows what a human eye perceives (so a blue and a green having the same `i` will feel as luminous as each other),
 * - or it is stored in `j` and is normalized so it can take any value between 0 and 1. */
class Eye {
  readonly space: CS
  /** Read and write the perceived rainbow color here. It indicates where the color is on the rainbow. */
  r: number
  /** Read and write the perceived purity here. It indicates how much of the pure color `r` is added to white. */
  p: number
  /** Read and write the perceived intensity of the light here. It is not normalized because `r` and `p` determine the maximum intensity that can be perceived and it is often less than 1.
   * If needed, property `iMax` tells the maximum value that `i` can hold for the current values of `r` and `p`.
   * `i` makes sense only if `CS.EyePercI` has been passed to the constructor, otherwise its value is undefined and setting a value has no effect. */
  i: number
  /** Read and write the normalized intensity of the light here. Any value between 0 and 1 is possible.
   * `j` makes sense only if `CS.EyeNormJ` has been passed to the constructor, otherwise its value is undefined and setting a value has no effect. */
  j: number
  /** Maximum value that `i` can have for the current values of `r` and `p`. The intensity that a human perceives depends on the color. For example, in RGB (so on all monitors) a 255 blue looks less bright than a 255 green so the `iMax` of blue is lower than the `iMax` of green.
   *  If you set `i` to a value higher than `iMax`, the color will correspond to RGB values greater than 255.
   * `iMax` is kept up-to-date automatically when you change `r` or `p`. */
  get iMax () : number {
    if (this.Imax.rOfValue !== this.r || this.Imax.pOfValue !== this.p) {
      if (this.space !== CS.EyePercI) {
        return 1
      }
      this.convertToRGB(Eye.rgbLinear, this.lowestImax, this.j) // calculates the RGB values for the current r and p, with a standardized intensity value (the smallest iMax possible) so the RGB cannot be black
      const h = Eye.RGBtoHighestChannel(Eye.rgbLinear.chan)
      this.Imax.value = (Eye.sensicolNorm[B] / Eye.rgbLinear.chan[h]) ** gammaInv
      this.Imax.rOfValue = this.r
      this.Imax.pOfValue = this.p
    }
    return this.Imax.value
  }

  /** Among all possible colors, this is the lowest `iMax` than can be met. In other words, the intensity `i` can be set to `lowestImax` for any `r` and `p`. Greater values of `i` will be impossible for some colors. */
  readonly lowestImax: number

  // constants of our perception model, all obtained empirically
  protected static readonly sensicol = [15, 20, 6] // sensitivity of the human eye to primaries, used to calculate the perceived intensity when channels add
  protected static readonly intercol = [10, 20, 5] // relative strengh of the primaries when they are mixed by pairs to obtain a pure intermediate color
  protected static readonly overwhite = [7, 10, 2] // perceived ability of the primaries to tint a white light when added to it (controls the width of the the grey part in a row where the purity goes from 0 to 1)
  protected static readonly phi = 1.7 // power law on the purity to make it feel linear to the human eye
  protected static readonly iotaD = 0.2 // when the primaries of a given color are ordered by perceived intensities (so by `value * sensicol`), tells how much the perceived dimmest contribute to the perceived intensity of the mix of the three
  protected static readonly iotaM = 0.2 // when the primaries of a given color are ordered by perceived intensities (so by `value * sensicol`), tells how much the second perceived dimmest contribute to the perceived intensity of the mix of the three
  // the following constants will be filled by the constructor
  protected static readonly sensicolNorm = [0, 0, 0]
  protected static readonly rPowers = [0, 0, 0] // accessed by the index of the dimmest primary
  protected static readonly rPowersInv = [0, 0, 0]
  protected static readonly overwhiteNorm = [0, 0, 0]
  protected static readonly phiInv = 1 / Eye.phi

  /**
   * @param space if `CS.EyePercI` is given, the intensity of the light will be stored in `i` and follow what a human eye perceives;
   * if `CS.EyeNormJ` is given, the intensity will be stored in `j` and normalized so it can take any value between 0 and 1. */
  constructor (space: CS) {
    if (space !== CS.EyePercI && space !== CS.EyeNormJ) {
      throw new Error('an Eye object can carry RPI/J information only')
    }
    if (!Eye.sensicolNorm[0]) {
      for (let k = R; k <= B; k++) {
        Eye.sensicolNorm[k] = Eye.sensicol[k] / (Eye.sensicol[R] + Eye.sensicol[G] + Eye.sensicol[B])
        Eye.overwhiteNorm[k] = Eye.overwhite[k] / (Eye.overwhite[R] + Eye.overwhite[G] + Eye.overwhite[B])
        const [h1, h2] = Eye.lowestChanToAnchors(k)
        Eye.rPowers[k] = 1 / Math.log2(Eye.intercol[h1] / Eye.intercol[h2] + 1)
        Eye.rPowersInv[k] = 1 / Eye.rPowers[k]
      }
    }
    this.lowestImax = Eye.sensicolNorm[B] ** gammaInv
    this.space = space
    this.r = this.p = this.i = this.j = 0
  }

  private Imax = {
    value: 0,
    rOfValue: -1,
    pOfValue: -1
  }

  protected static rgbLinear = new RGB(CS.RGBlinear)

  /** Copies a RGB or Eye instance into the current Eye instance.
   * Color spaces are automatically converted into the color space of the current Eye instance if they differ.
   * @returns the instance that you import into (so not the parameter) */
  import (from: RGB | Eye) : Eye {
    if (from.space === CS.RGBgamma) {
      from = Eye.rgbLinear.import(from)
    }
    if (from.space === CS.RGBlinear) {
      // conversion of RGB into Eye
      const rgb = (from as RGB).chan
      const l = Eye.RGBtoLowestChannel(rgb)
      const [h1, h2] = Eye.lowestChanToAnchors(l)
      const anchor2Contribution = (rgb[h2] - rgb[l]) * Eye.intercol[h2]
      const anchors12Contributions = (rgb[h1] - rgb[l]) * Eye.intercol[h1] + anchor2Contribution
      if (anchors12Contributions < 0.001) { // the 3 channels are equal, so we have black, grey or white
        this.r = 0.5
        this.p = 0
      } else {
        this.r = (h1 + (anchor2Contribution / anchors12Contributions) ** Eye.rPowers[l]) / 3
        this.p = ((rgb[h1] - rgb[l]) * Eye.overwhiteNorm[h1] + (rgb[h2] - rgb[l]) * Eye.overwhiteNorm[h2]) / (rgb[l] * Eye.overwhiteNorm[l] + rgb[h1] * Eye.overwhiteNorm[h1] + rgb[h2] * Eye.overwhiteNorm[h2])
        this.p **= Eye.phiInv
      }
      this.fillIntensityFromRGB(rgb)
    } else {
      from = from as Eye // for the static checker
      this.r = from.r
      this.p = from.p
      if (from.space === this.space) {
        // we copy an Eye into our Eye of the same variant
        this.i = from.i
        this.j = from.j
        this.Imax.value = from.Imax.value
        this.Imax.rOfValue = from.Imax.rOfValue
        this.Imax.pOfValue = from.Imax.pOfValue
      } else {
        // we must convert an Eye variant into our variant (EyePercI <-> EyeNormJ)
        from.export(Eye.rgbLinear)
        this.fillIntensityFromRGB(Eye.rgbLinear.chan)
      }
    }
    return this
  }

  /** Copies this Eye instance into another RGB or Eye instance.
   * The color space of the current instance is automatically converted into the color space of the target instance if needed.
   * @param to existing instance to fill, or if the identifier of a color space is given instead of an object, `export` creates an instance for you, fills it and returns it.
   * @returns target instance (same as `to` if `to` was an instance) */
  export (to: RGB | Eye | CS) : RGB | Eye {
    if (typeof to !== 'object') {
      to = (to === CS.RGBlinear || to === CS.RGBgamma) ? new RGB(to) : new Eye(to)
    }
    if (to.space === CS.EyePercI || to.space === CS.EyeNormJ) {
      to.import(this)
    } else
      if (to.space === CS.RGBgamma) {
        this.export(Eye.rgbLinear).export(to)
      } else {
        this.convertToRGB(to as RGB, this.i, this.j)
      }
    return to
  }

  protected convertToRGB (to: RGB, i: number, j: number) {
    const [l, , h] = Eye.rToChannelOrder(this.r)
    const [h1, h2] = Eye.lowestChanToAnchors(l)
    const ratio = [0, 0, 0]
    const D = (3 * this.r - h1) ** Eye.rPowersInv[l]
    const p = this.p ** Eye.phi
    if (D < 0.001 && p > 0.999) {
      to.chan[l] = 0
      to.chan[h1] = (this.space === CS.EyePercI) ? i ** gamma / Eye.sensicolNorm[h1] : j ** gamma
      to.chan[h2] = 0
    } else {
      if (D > 0.999 || p < 0.001) {
        ratio[l] = (1 - p) / (1 + p * (Eye.overwhiteNorm[l] + Eye.overwhiteNorm[h1]) / Eye.overwhiteNorm[h2])
        ratio[h1] = ratio[l]
      } else {
        const A = D * Eye.intercol[h1] / ((1 - D) * Eye.intercol[h2])
        const B = (1 - p) / p * (Eye.overwhiteNorm[h1] + A * Eye.overwhiteNorm[h2])
        const Q = 1 / (A + B)
        ratio[l] = B * Q
        ratio[h1] = Q + ratio[l]
      }
      ratio[h2] = 1
      if (this.space === CS.EyePercI) {
        const Ir = [Eye.sensicolNorm[R] * ratio[R], Eye.sensicolNorm[G] * ratio[G], Eye.sensicolNorm[B] * ratio[B]]
        const [x, y, z] = Eye.RGBtoOrder(Ir)
        to.chan[h2] = i ** gamma / (Ir[x] * Eye.iotaD + Ir[y] * Eye.iotaM + Ir[z])
        to.chan[h1] = to.chan[h2] * ratio[h1]
      } else { // (this.space === CS.EyeNormJ)
        to.chan[h] = j ** gamma
        if (h === h2) {
          to.chan[h1] = to.chan[h2] * ratio[h1]
        } else {
          to.chan[h2] = to.chan[h1] / ratio[h1]
        }
      }
      to.chan[l] = to.chan[h2] * ratio[l]
    }
  }

  protected fillIntensityFromRGB (rgb : number[]) : void {
    if (this.space === CS.EyePercI) {
      const I = [Eye.sensicolNorm[R] * rgb[R], Eye.sensicolNorm[G] * rgb[G], Eye.sensicolNorm[B] * rgb[B]]
      const [x, y, z] = Eye.RGBtoOrder(I)
      this.i = (I[x] * Eye.iotaD + I[y] * Eye.iotaM + I[z]) ** gammaInv
    } else {
      const h = Eye.RGBtoHighestChannel(rgb)
      // due to our definitions of r and p, rgb[h] turns out to be the ratio between i (no matter how i is calculated) and the maximum i that could be reached with r and p constant
      this.j = rgb[h] ** gammaInv
    }
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

  protected static RGBtoOrder (rgb : number[]) : Channel[] {
    if (rgb[R] < rgb[G]) {
      if (rgb[G] < rgb[B]) { return [R, G, B] }
      return (rgb[R] < rgb[B]) ? [R, B, G] : [B, R, G]
    }
    if (rgb[R] < rgb[B]) { return [G, R, B] }
    return (rgb[G] < rgb[B]) ? [G, B, R] : [B, G, R] // note that white, grey and black return [G, R], which is what we want (they are the primaries bringing the most brightness to the human eye)
  }

  /** @returns the anchors in the same order as on the rainbow (note that R is both before G and after B) */
  protected static lowestChanToAnchors (lowestChan: Channel) : Channel[] {
    switch (lowestChan) {
      case R : return [G, B]
      case G : return [B, R]
      case B : return [R, G]
    }
    return [] // impossible but the static checker believes it can happen
  }

  /** @returns the anchor having the lowest value followed by the highest-value anchor */
  /* protected static RGBtoAnchorOrder (rgb : number[]) : Channel[] {
    if (rgb[R] < rgb[G]) {
      if (rgb[G] < rgb[B]) { return [G, B] }
      return (rgb[R] < rgb[B]) ? [B, G] : [R, G]
    }
    if (rgb[R] < rgb[B]) { return [R, B] }
    return (rgb[G] < rgb[B]) ? [B, R] : [G, R] // note that white, grey and black return [G, R], which is what we want (they are the primaries bringing the most brightness to the human eye)
  } */

  /** @returns the order of the channels from the lowest value to the highest-value */
  protected static rToChannelOrder (r : number) : Channel[] {
    if (r < 2 / 6) {
      return (r < 1 / 6) ? [B, G, R] : [B, R, G]
    }
    if (r < 4 / 6) {
      return (r < 3 / 6) ? [R, B, G] : [R, G, B]
    }
    return (r < 5 / 6) ? [G, R, B] : [G, B, R]
  }
}

//
// TESTS AND ADJUSTEMENTS
//

const cons = console
const colors : Array<Array<Array<{rgb: RGB, eye: Eye}>>> = []
const numR = 3
const numP = 45
const numI = 40
const color = new Eye(CS.EyePercI)
let maxIntensityMinIndex = 1000
for (let r = 0; r <= numR; r++) {
  colors.push([])
  for (let p = 0; p <= numP; p++) {
    colors[r].push([])
    for (let i = 0; i <= numI; i++) {
      if (i > 0 && (i / numI) > color.iMax) {
        if (i - 1 < maxIntensityMinIndex) { maxIntensityMinIndex = i - 1 }
        break
      }
      color.r = r / numR
      color.p = p / numP
      color.i = i / numI
      colors[r][p].push({ rgb: color.export(CS.RGBgamma) as RGB, eye: color.export(CS.EyePercI) as Eye })
      const rgb = (colors[r][p][i].eye.export(CS.EyeNormJ).export(CS.RGBgamma) as RGB).chan
      if (colors[r][p][i].rgb.chan[0] !== rgb[0] || colors[r][p][i].rgb.chan[1] !== rgb[1] || colors[r][p][i].rgb.chan[2] !== rgb[2]) {
        cons.log(colors[r][p][i].rgb.chan, rgb)
        cons.log(colors[r][p][i].eye, colors[r][p][i].eye.export(CS.EyeNormJ))
      }
      const rgb2 = ((new Eye(CS.EyePercI)).import(colors[r][p][i].rgb)).export(CS.RGBgamma) as RGB
      if (colors[r][p][i].rgb.chan[0] !== rgb2.chan[0] || colors[r][p][i].rgb.chan[1] !== rgb2.chan[1] || colors[r][p][i].rgb.chan[2] !== rgb2.chan[2]) {
        cons.log(colors[r][p][i].rgb, rgb)
        cons.log(colors[r][p][i].eye, colors[r][p][i].eye.export(CS.EyeNormJ))
      }
    }
  }
}

const rainbowSameI: Array<RGB> = []
const rainbowSameJ: Array<RGB> = []
for (let r = 0; r <= 200; r++) {
  color.r = r / 200
  color.p = 1
  color.i = color.lowestImax
  rainbowSameI.push(color.export(CS.RGBgamma) as RGB)
  color.i = color.iMax
  rainbowSameJ.push(color.export(CS.RGBgamma) as RGB)
  if (rainbowSameJ[r].chan[R] > 255 || rainbowSameJ[r].chan[G] > 255 || rainbowSameJ[r].chan[B] > 255) { cons.log('ERROR', rainbowSameJ[r].chan, color) }
}

const linearCount80: number[] = []
for (let k = 0; k < 80; k++) {
  linearCount80.push(k)
}

const linearCount14: number[] = []
for (let k = 0; k <= 14; k++) {
  linearCount14.push(k)
}

const iotaIadjuster: Array<RGB> = []
color.p = 0
for (let k = 0; k <= 16; k++) {
  color.i = color.iMax * k / 16
  iotaIadjuster.push(color.export(CS.RGBgamma) as RGB)
}

const iotaJadjuster: Array<RGB> = []
const colorJ = new Eye(CS.EyeNormJ)
for (let k = 0; k <= 16; k++) {
  colorJ.p = 0
  colorJ.j = k / 16
  iotaJadjuster.push(colorJ.export(CS.RGBgamma) as RGB)
}
</script>

<template>
  <div style="background-color: rgb(128,128,128)">
    <br>
    <h1>Adjustement of sensicol</h1>
    These pure colors must all have the same perceived intensity: <br><br>
    <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameI[0].chan[R] + ',' + rainbowSameI[0].chan[G] + ',' + rainbowSameI[0].chan[B] + ')'" />
    <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameI[Math.round(rainbowSameI.length*2/6)].chan[R] + ',' + rainbowSameI[Math.round(rainbowSameI.length*2/6)].chan[G] + ',' + rainbowSameI[Math.round(rainbowSameI.length*2/6)].chan[B] + ')'" />
    <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameI[Math.round(rainbowSameI.length*4/6)].chan[R] + ',' + rainbowSameI[Math.round(rainbowSameI.length*4/6)].chan[G] + ',' + rainbowSameI[Math.round(rainbowSameI.length*4/6)].chan[B] + ')'" />
    <br><br>

    <h1>Adjustement of iotaM</h1>
    The secondaries colors must all have the same perceived intensity as their surrounding primaries: <br><br>
    <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameI[0].chan[R] + ',' + rainbowSameI[0].chan[G] + ',' + rainbowSameI[0].chan[B] + ')'" />
    <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameI[Math.round(rainbowSameI.length*1/6)].chan[R] + ',' + rainbowSameI[Math.round(rainbowSameI.length*1/6)].chan[G] + ',' + rainbowSameI[Math.round(rainbowSameI.length*1/6)].chan[B] + ')'" />
    <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameI[Math.round(rainbowSameI.length*2/6)].chan[R] + ',' + rainbowSameI[Math.round(rainbowSameI.length*2/6)].chan[G] + ',' + rainbowSameI[Math.round(rainbowSameI.length*2/6)].chan[B] + ')'" />
    <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameI[Math.round(rainbowSameI.length*3/6)].chan[R] + ',' + rainbowSameI[Math.round(rainbowSameI.length*3/6)].chan[G] + ',' + rainbowSameI[Math.round(rainbowSameI.length*3/6)].chan[B] + ')'" />
    <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameI[Math.round(rainbowSameI.length*4/6)].chan[R] + ',' + rainbowSameI[Math.round(rainbowSameI.length*4/6)].chan[G] + ',' + rainbowSameI[Math.round(rainbowSameI.length*4/6)].chan[B] + ')'" />
    <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameI[Math.round(rainbowSameI.length*5/6)].chan[R] + ',' + rainbowSameI[Math.round(rainbowSameI.length*5/6)].chan[G] + ',' + rainbowSameI[Math.round(rainbowSameI.length*5/6)].chan[B] + ')'" />
    <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameI[rainbowSameI.length-1].chan[R] + ',' + rainbowSameI[rainbowSameI.length-1].chan[G] + ',' + rainbowSameI[rainbowSameI.length-1].chan[B] + ')'" />
    <br><br>

    <h1>Adjustement of iotaD </h1>
    <div style="background-color: rgb(160,160,160)">
      iotaD adjust the perceived intensity of the grey, that should feel as bright as the primaries
      <div v-for="(rRow,r) of colors" :key="r" style="text-align: center;">
        <br>
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + rRow[rRow.length-1][maxIntensityMinIndex].rgb.chan[R] + ',' + rRow[rRow.length-1][maxIntensityMinIndex].rgb.chan[G] + ',' + rRow[rRow.length-1][maxIntensityMinIndex].rgb.chan[B] + ')'" />
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + rRow[0][maxIntensityMinIndex].rgb.chan[R] + ',' + rRow[0][maxIntensityMinIndex].rgb.chan[G] + ',' + rRow[0][maxIntensityMinIndex].rgb.chan[B] + ')'" />
        <br>
      </div>
    </div>

    <h1>Adjustement of overwhite and phi </h1>
    overwhite adjusts the width of the the grey part in each row. <br>
    phi adjusts the linearity of the transition from grey to pure. <br> <br>
    <div v-for="(rRow,r) of colors" :key="r" style="border: 0px;">
      <span v-for="(pRow,p) of rRow" :key="p">
        <div style="display: inline-block; width: 10px; height: 40px;" :style="'background-color: rgb(' + pRow[maxIntensityMinIndex].rgb.chan[R] + ',' + pRow[maxIntensityMinIndex].rgb.chan[G] + ',' + pRow[maxIntensityMinIndex].rgb.chan[B] + ')'" />
      </span>
    </div>
    <br>

    <div style="background-color: rgb(160,160,160)">
      Each middle square must feel as different from its left square as from its right square.
      The better this criterion is approched, the more linear in `p` the perceived purity is.
      <div v-for="(rRow,r) of colors" :key="r" style="text-align: center;">
        <br>
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + rRow[0][maxIntensityMinIndex].rgb.chan[R] + ',' + rRow[0][maxIntensityMinIndex].rgb.chan[G] + ',' + rRow[0][maxIntensityMinIndex].rgb.chan[B] + ')'" />
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + rRow[rRow.length/2][maxIntensityMinIndex].rgb.chan[R] + ',' + rRow[rRow.length/2][maxIntensityMinIndex].rgb.chan[G] + ',' + rRow[rRow.length/2][maxIntensityMinIndex].rgb.chan[B] + ')'" />
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + rRow[rRow.length-1][maxIntensityMinIndex].rgb.chan[R] + ',' + rRow[rRow.length-1][maxIntensityMinIndex].rgb.chan[G] + ',' + rRow[rRow.length-1][maxIntensityMinIndex].rgb.chan[B] + ')'" />
        <br>
      </div>
    </div>
    <br>

    <h1>Adjustement of intercol</h1>

    TODO: test pour ajuster la linearité de r: couper le rainbow en 6
    The better this criterion is approched, the more linear in `r` the perceived hue is.
    <br><br>
    <span v-for="(c,i) of rainbowSameI" :key="i" style="display: inline-block; width: 6px; height: 40px;" :style="'background-color: rgb(' + c.chan[R] + ',' + c.chan[G] + ',' + c.chan[B] + ')'" />
    <br><br>
    <span v-for="(c,i) of rainbowSameJ" :key="i" style="display: inline-block; width: 6px; height: 40px;" :style="'background-color: rgb(' + c.chan[R] + ',' + c.chan[G] + ',' + c.chan[B] + ')'" />
    <br><br>

    <h1>Screen calibration: color balance</h1>
    TODO: test de balance des primaires (les secondaires doivent etre au milieu) : dessiner trois colonnes sur le rainbow

    <h1>Screen calibration: linearity of extremes greys</h1>
    The perceived linearity of your screen in extreme greys (near black and white) is good if each middle square feels as different from its left square as from its right square.
    <br><br>
    <div v-for="k of linearCount14" :key="k" style="text-align: center; background-color: #7030f0">
      <span v-if="k == 0 || k==1 || k==2 || k==12 || k==13 || k==14">
        <br>
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + iotaIadjuster[0+k].chan[R] + ',' + iotaIadjuster[0+k].chan[G] + ',' + iotaIadjuster[0+k].chan[B] + ')'">
          I
        </div>
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + iotaIadjuster[1+k].chan[R] + ',' + iotaIadjuster[1+k].chan[G] + ',' + iotaIadjuster[1+k].chan[B] + ')'">
          I
        </div>
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + iotaIadjuster[2+k].chan[R] + ',' + iotaIadjuster[2+k].chan[G] + ',' + iotaIadjuster[2+k].chan[B] + ')'">
          I
        </div>
        <br>
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + iotaJadjuster[0+k].chan[R] + ',' + iotaJadjuster[0+k].chan[G] + ',' + iotaJadjuster[0+k].chan[B] + ')'">
          J
        </div>
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + iotaJadjuster[1+k].chan[R] + ',' + iotaJadjuster[1+k].chan[G] + ',' + iotaJadjuster[1+k].chan[B] + ')'">
          J
        </div>
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + iotaJadjuster[2+k].chan[R] + ',' + iotaJadjuster[2+k].chan[G] + ',' + iotaJadjuster[2+k].chan[B] + ')'">
          J
        </div>
        <br>
      </span>
    </div>

    <h1>Screen calibration: gamma in medium intensities.</h1>
    Your screen gamma is 2.2 on the three channels if the following squares look plain (the center parts must not look brighter or dimmer).<br>
    For the test to work properly: the zoom of your browser must be 100% and you should look from far enough (or without glasses)
    <br><br>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(135,135,135)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (k%2)*186 + ',' + (k%2)*186 + ',' + (k%2)*186 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + ((k+1)%2)*186 + ',' + ((k+1)%2)*186 + ',' + ((k+1)%2)*186 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (k%2)*186 + ',' + (k%2)*186 + ',' + (k%2)*186 + ')'" />
    </div>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(186,186,186)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (k%2)*255 + ',' + (k%2)*255 + ',' + (k%2)*255 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + ((k+1)%2)*255 + ',' + ((k+1)%2)*255 + ',' + ((k+1)%2)*255 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (k%2)*255 + ',' + (k%2)*255 + ',' + (k%2)*255 + ')'" />
    </div>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(223,223,223)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (255-(k%2)*69) + ',' + (255-(k%2)*69) + ',' + (255-(k%2)*69) + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (255-((k+1)%2)*69) + ',' + (255-((k+1)%2)*69) + ',' + (255-((k+1)%2)*69) + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (255-(k%2)*69) + ',' + (255-(k%2)*69) + ',' + (255-(k%2)*69) + ')'" />
    </div>
    <br><br>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(135,0,0)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (k%2)*186 + ',' + 0 + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + ((k+1)%2)*186 + ',' + 0 + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (k%2)*186 + ',' + 0 + ',' + 0 + ')'" />
    </div>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(186,0,0)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (k%2)*255 + ',' + 0 + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + ((k+1)%2)*255 + ',' + 0 + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (k%2)*255 + ',' + 0 + ',' + 0 + ')'" />
    </div>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(223,0,0)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (255-(k%2)*69) + ',' + 0 + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (255-((k+1)%2)*69) + ',' + 0 + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + (255-(k%2)*69) + ',' + 0 + ',' + 0 + ')'" />
    </div>
    <br><br>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,135,0)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + (k%2)*186 + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + ((k+1)%2)*186 + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + (k%2)*186 + ',' + 0 + ')'" />
    </div>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,186,0)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + (k%2)*255 + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + ((k+1)%2)*255 + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + (k%2)*255 + ',' + 0 + ')'" />
    </div>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,223,0)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + (255-(k%2)*69) + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + (255-((k+1)%2)*69) + ',' + 0 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + (255-(k%2)*69) + ',' + 0 + ')'" />
    </div>
    <br><br>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,0,135)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (k%2)*186 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + ((k+1)%2)*186 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (k%2)*186 + ')'" />
    </div>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,0,186)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (k%2)*255 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + ((k+1)%2)*255 + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (k%2)*255 + ')'" />
    </div>
    <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,0,223)">
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (255-(k%2)*69) + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (255-((k+1)%2)*69) + ')'" /><br>
      <div v-for="k of linearCount80" :key="k" style="display: inline-block; width: 1px; height: 20px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (255-(k%2)*69) + ')'" />
    </div>
    <br><br>
    <h1>All wavelenghts, purities and perceived intensities :</h1>
    <br>
    <div v-for="(rRow,r) of colors" :key="r">
      <div v-for="(pRow,p) of rRow" :key="p">
        <div v-if="!(p%3)">
          <div v-for="(c,i) of pRow" :key="i" style="display: inline-block; width: 20px; height: 20px;" :style="'background-color: rgb(' + c.rgb.chan[R] + ',' + c.rgb.chan[G] + ',' + c.rgb.chan[B] + ')'">
            {{ (c.rgb.chan[R] > 255 || c.rgb.chan[G] > 255 || c.rgb.chan[B] > 255) ? 'E+' + (cons.log(c.rgb.chan, c.eye)! || '') : '' }}
            {{ (c.rgb.chan[R] < 0 || c.rgb.chan[G] < 0 || c.rgb.chan[B] < 0) ? 'E-' + (cons.log(c.rgb.chan, c.eye)! || '') : '' }}
          </div>
        </div>
      </div>
      <br>
    </div>
  </div>
</template>

<style lang="scss" scoped>
</style>
