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
            this.chan[i] = Math.round(((from as RGB).chan[i] ** (1 / gamma)) * 255)
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

/** Color space supposedly close to human perception, in two variants (depending on the parameter given to the constructor) :
 * - either the intensity of the light is stored in `i` and follows what a human eye perceives (so a blue thing and a green thing having the same `i` will feel as luminous as each other),
 * - or it is stored in `j` and is normalized so it can take any value between 0 and 1. */
class Eye {
  readonly space: CS
  /** Read and write the perceived rainbow color here. It indicates where the color is on the rainbow. */
  r: number
  /** Read and write the perceived purity here. It indicates how much light not contributing to the perceived rainbow color is present. */
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
      this.convertToRGB(Eye.rgbLinear, Eye.lowestImax, this.j) // calculates the RGB values for the current r and p, with a standardized intensity value (the smallest iMax possible) so the RGB cannot be black
      const h = Eye.RGBtoHighestChannel(Eye.rgbLinear.chan)
      this.Imax.value = (Eye.lowestImaxIotaInv / Eye.rgbLinear.chan[h]) ** (1 / gamma)
      this.Imax.rOfValue = this.r
      this.Imax.pOfValue = this.p
    }
    return this.Imax.value
  }

  // constants of our perception model
  protected static readonly relLum = [30, 55, 15] // coefficients to calculate the perceived intensity when channels add (obtained empiricially by tweaking the definition of "relative luminance" in the ITU-R Recommendation BT.601)
  protected static readonly mix = [35, 45, 20] // This coeffients are correlated with the relative strengh of the primaries when they are mixed by pair to obtain a pure intermediate color (this coeffs are obtained empiricially). Their effect is to shift the position of the 3 secondaries (and all intermediaries) in the rainbow.
  protected static readonly phi = 0.55 // This coeffient and its brother just below are correlated with the strengh of the green and blue primaries relative to white when they are mixed to white light thus altering their purity saturation (this coeffs are obtained empiricially).
  protected static readonly phi = 0.15 //

  // the following constants will be filled by the constructor
  protected static relLumNorm = [0, 0, 0]
  protected static phi = [0, 0, 0]
  protected static rKey = [0, 0, 0, 0, 0, 0, 0] // (will be filled by the constructor) coordinates of the primaries and secondaries on the rainbow
  protected static readonly lowestImaxIotaInv = Eye.relLum[B]

  /** Among all possible colors, this is the lowest `iMax` than can be met. In other words, the intensity `i` can be set to `lowestImax` for any `r` and `p`. Greater values of `i` will be impossible for some colors. */
  static readonly lowestImax = Eye.relLum[B] ** (1 / gamma)

  /**
   * @param space if `CS.EyePercI` is given, the intensity of the light will be stored in `i` and follow what a human eye perceives;
   * if `CS.EyeNormJ` is given, the intensity will be stored in `j` and normalized so it can take any value between 0 and 1. */
  constructor (space: CS) {
    if (space !== CS.EyePercI && space !== CS.EyeNormJ) {
      throw new Error('an Eye object can carry RPI/J information only')
    }
    if (!Eye.rKey[1]) {
      for (let k = R; k <= B; k++) {
        Eye.relLumNorm[k] = Eye.relLum[k] / (Eye.relLum[R] + Eye.relLum[G] + Eye.relLum[B])
      }
      const lSequence = [B, R, G]
      for (let k = 0; k <= 6; k++) {
        if (k % 2) {
          const [h1, h2] = Eye.lowestChanToAnchors(lSequence[Math.floor(k / 2)])
          Eye.rKey[k] = (h1 + 1 * Eye.mix[h2] / (Eye.mix[h1] + Eye.mix[h2])) / 3
        } else {
          Eye.rKey[k] = k / 6
        }
      }
    }
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
      const anchor2Contribution = (rgb[h2] - rgb[l]) * Eye.mix[h2]
      const anchors12Contributions = (rgb[h1] - rgb[l]) * Eye.mix[h1] + anchor2Contribution
      if (anchors12Contributions <= 0) { // the 3 channels are equal, so we have black, grey or white
        this.r = 0.5
        this.p = 0
      } else {
        this.r = (h1 + anchor2Contribution / anchors12Contributions) / 3
        this.p = anchors12Contributions / (rgb[l] * Eye.mix[l] + rgb[h1] * Eye.mix[h1] + rgb[h2] * Eye.mix[h2])
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
    const [l, m, h] = Eye.rToChannelOrder(this.r)
    const [h1, h2] = Eye.lowestChanToAnchors(l)
    const q = this.p * ((3 * this.r - h1) * ((Eye.mix[l] + Eye.mix[h1]) / Eye.mix[h2] + 1) - 1) + 1
    if (q <= 0) { // `q == 0` is possible only if `3r-h1 == 0` and `p == 1`, which implies that `rgb[h2] == rgb[l]` (see the definition of `r` in the `import` method) with the highest purity (primary color), so `rgb[h2]=0` and `rgb[l]=0` and only `rgb[h1]` has a value
      to.chan[l] = to.chan[m] = 0
      to.chan[h] = (this.space === CS.EyePercI) ? i ** gamma / Eye.relLumNorm[h] : j ** gamma
    } else {
      const ratio = [0, 0, 0]
      ratio[l] = (1 - this.p) / q
      ratio[h1] = (Eye.mix[l] * this.p + Eye.mix[h1] + Eye.mix[h2] * (1 - q)) / (Eye.mix[h1] * q)
      ratio[h2] = 1
      if (this.space === CS.EyePercI) {
        const Ir = [Eye.relLumNorm[R] * ratio[R], Eye.relLumNorm[G] * ratio[G], Eye.relLumNorm[B] * ratio[B]]
        const [y, z] = Eye.RGBtoAnchorOrder(Ir)
        to.chan[h2] = i ** gamma / (Ir[y] + Ir[z])
        to.chan[h1] = to.chan[h2] * ratio[h1]
      } else { // (this.space === CS.EyeNormJ)
        to.chan[h] = j ** gamma
        if (h === h2) {
          to.chan[h1] = to.chan[h2] * ratio[h1]
        } else {
          to.chan[h2] = to.chan[h1] / ratio[h1] // If ratio[h1]=0, we cannot reach this point. Proof: (2+p)/q-1=0 => q=2+p => r=(h1+(2+1/p)/3)/3 and noticing that 2+1/p is at least 3 we obtain r>=(h1+1)/3, so h1=R:r>=1/3 , h1=G:r>=2/3, h1=B:r=1. The first two cases are impossible because lowestChanToAnchors(rToChannelOrder(1/3 or 2/3)[0]) returns [G,B] or [B,R] whose h1 is respectively G or B, thus contradicting the first two assumptions. Regarding the third assumption: lowestChanToAnchors(rToChannelOrder(1)[0]) returns [B,R] (so the assumption holds) whose h2 is R, and the h in rToChannelOrder(1) is R, so h2=h.
        }
      }
      to.chan[l] = to.chan[h2] * ratio[l]
    }
  }

  protected fillIntensityFromRGB (rgb : number[]) : void {
    if (this.space === CS.EyePercI) {
      const I = [Eye.relLumNorm[R] * rgb[R], Eye.relLumNorm[G] * rgb[G], Eye.relLumNorm[B] * rgb[B]]
      const [y, z] = Eye.RGBtoAnchorOrder(I)
      this.i = (I[y] + I[z]) ** (1 / gamma)
    } else {
      // due to our definitions of r and p, this value turns out to be the ratio between i (no matter how i is calculated) and the maximum i that could be reached with r and p constant
      const h = Eye.RGBtoHighestChannel(rgb)
      this.j = rgb[h] ** (1 / gamma)
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
  protected static RGBtoAnchorOrder (rgb : number[]) : Channel[] {
    if (rgb[R] < rgb[G]) {
      if (rgb[G] < rgb[B]) { return [G, B] }
      return (rgb[R] < rgb[B]) ? [B, G] : [R, G]
    }
    if (rgb[R] < rgb[B]) { return [R, B] }
    return (rgb[G] < rgb[B]) ? [B, R] : [G, R] // note that white, grey and black return [G, R], which is what we want (they are the primaries bringing the most brightness to the human eye)
  }

  /** @returns the order of the channels from the lowest value to the highest-value */
  protected static rToChannelOrder (r : number) : Channel[] {
    if (r < Eye.rKey[2]) {
      return (r < Eye.rKey[1]) ? [B, G, R] : [B, R, G]
    }
    if (r < Eye.rKey[4]) {
      return (r < Eye.rKey[3]) ? [R, B, G] : [R, G, B]
    }
    return (r < Eye.rKey[5]) ? [G, R, B] : [G, B, R]
  }
}

//
// TESTS AND ADJUSTEMENTS
//

const cons = console
const colors : Array<Array<Array<{rgb: RGB, eye: Eye}>>> = []
const numR = 6
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
  color.i = Eye.lowestImax
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
  <div style="background-color: rgb(126,126,126)">
    <br><br>
    <span v-for="(c,i) of rainbowSameI" :key="i" style="display: inline-block; width: 6px; height: 40px;" :style="'background-color: rgb(' + c.chan[R] + ',' + c.chan[G] + ',' + c.chan[B] + ')'" />
    <br><br>
    <span v-for="(c,i) of rainbowSameJ" :key="i" style="display: inline-block; width: 6px; height: 40px;" :style="'background-color: rgb(' + c.chan[R] + ',' + c.chan[G] + ',' + c.chan[B] + ')'" />
    <br><br>

    <div v-for="(rRow,r) of colors" :key="r" style="border: 0px;">
      <span v-for="(pRow,p) of rRow" :key="p">
        <div style="display: inline-block; width: 20px; height: 40px;" :style="'background-color: rgb(' + pRow[maxIntensityMinIndex].rgb.chan[R] + ',' + pRow[maxIntensityMinIndex].rgb.chan[G] + ',' + pRow[maxIntensityMinIndex].rgb.chan[B] + ')'" />
      </span>
      <span style="display: inline-block; width: 20px; height: 40px; border: 0px;" :style="'background-color: rgb(' + rRow[0][maxIntensityMinIndex].rgb.chan[R] + ',' + rRow[0][maxIntensityMinIndex].rgb.chan[G] + ',' + rRow[0][maxIntensityMinIndex].rgb.chan[B] + ')'" />
    </div>
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

    <h1>Adjustement of rhoG and rhoB</h1>
    1. Each middle square must feel as different from its left square as from its right square.
    The better this criterion is approched, the more linear in `p` the perceived purity is.
    <div style="background-color: rgb(160,160,160)">
      <div v-for="(rRow,r) of colors" :key="r" style="text-align: center;">
        <br>
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + rRow[0][maxIntensityMinIndex].rgb.chan[R] + ',' + rRow[0][maxIntensityMinIndex].rgb.chan[G] + ',' + rRow[0][maxIntensityMinIndex].rgb.chan[B] + ')'" />
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + rRow[rRow.length/2][maxIntensityMinIndex].rgb.chan[R] + ',' + rRow[rRow.length/2][maxIntensityMinIndex].rgb.chan[G] + ',' + rRow[rRow.length/2][maxIntensityMinIndex].rgb.chan[B] + ')'" />
        <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + rRow[rRow.length-1][maxIntensityMinIndex].rgb.chan[R] + ',' + rRow[rRow.length-1][maxIntensityMinIndex].rgb.chan[G] + ',' + rRow[rRow.length-1][maxIntensityMinIndex].rgb.chan[B] + ')'" />
        <br>
      </div>
    </div>
    <br>
    // test separé pour ajuster la linearité de r
    2. At the same time, try to balance the widths of the color domains on the rainbow.
    The better this criterion is approched, the more linear in `r` the perceived hue is.
    <br><br>
    <span v-for="(c,i) of rainbowSameJ" :key="i" style="display: inline-block; width: 6px; height: 40px;" :style="'background-color: rgb(' + c.chan[R] + ',' + c.chan[G] + ',' + c.chan[B] + ')'" />
    <br><br>

    // TODO: test de balance des primaires (les secondaires doivent etre au milieu)

    <h1>Screen calibration of extremes greys</h1>
    Your screen has a good linearity in extreme greys (near black and white) if each middle square feels as different from its left square as from its right square.
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

    <h1>Screen calibration of middle greys.</h1>
    Your screen is perfectly linear in mid-greys if the following squares look plain (the center parts must not look brighter or dimmer).<br>
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
  </div>
</template>

<style lang="scss" scoped>
</style>
