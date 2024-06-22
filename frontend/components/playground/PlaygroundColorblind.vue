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

  /** Copy a RGB or Eye instance into the current RGB instance. A regular array of 3 values can also be given (in this order: R,G,B).
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
            this.chan[i] = Math.round(((from as RGB).chan[i] ** 0.454545) * 255)
          }
          break
        case CS.RGBgamma :
          for (let i = R; i <= B; i++) {
            this.chan[i] = ((from as RGB).chan[i] / 255) ** 2.2
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
      if (this.chan[i] < 0) { this.chan[i] = 0 }
      if (this.chan[i] > max) { this.chan[i] = max }
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
    if (this.Imax.rOfValue !== this.r || this.Imax.pOfValue !== this.p) {
      if (this.space !== CS.EyePercI) {
        return 1
      }
      this.convertToRGB(Eye.rgbLinear, contributionToI[B]) // calculates the RGB values for the current r and p, with a standardized intensity value (the smallest iMax possible, which is 0.1 when this comment was written) so the RGB cannot be black
      const h = Eye.RGBtoHighestChannel(Eye.rgbLinear.chan)
      this.Imax.value = contributionToI[B] / Eye.rgbLinear.chan[h]
      this.Imax.rOfValue = this.r
      this.Imax.pOfValue = this.p
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
    rOfValue: -1,
    pOfValue: -1
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
      const rgb = (from as RGB).chan
      const l = Eye.RGBtoLowestChannel(rgb)
      const [h1, h2] = Eye.lowestChanToAnchors(l)
      const anchorSum = rgb[h1] + rgb[h2]
      const anchorContributions = anchorSum - 2 * rgb[l]
      if (anchorContributions <= 0) { // the 3 channels are equal, so we have black, grey or white
        this.r = 0.5
        this.p = 0
      } else {
        this.r = ((rgb[h2] - rgb[l]) / anchorContributions + h1) / 3
        this.p = anchorContributions / (anchorSum + rgb[l])
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
    } else
      if (to.space === CS.RGBgamma) {
        this.export(Eye.rgbLinear).export(to)
      } else {
        this.convertToRGB(to as RGB, this.i)
      }
    return to
  }

  protected convertToRGB (to: RGB, i: number) {
    const [l, m, h] = Eye.rToChannelOrder(this.r)
    const [h1, h2] = Eye.lowestChanToAnchors(l)
    const q = 1 + (3 * (3 * this.r - h1) - 1) * this.p
    if (q <= 0) { // `q == 0` is possible only if `3r-h1 == 0` and `p == 1`, which implies that `rgb[h2] == rgb[l]` (see the definition of `r` in the `import` method) with the highest purity (primary color), so `rgb[h2]=0` and `rgb[l]=0` and only `rgb[h1]` has a value
      to.chan[l] = to.chan[m] = 0
      to.chan[h] = (this.space === CS.EyePercI) ? i / contributionToI[h] : this.j
    } else {
      const ratio = [0, 0, 0]
      ratio[l] = (1 - this.p) / q
      ratio[h1] = (2 + this.p) / q - 1
      ratio[h2] = 1
      if (this.space === CS.EyePercI) {
        const Ir = [contributionToI[R] * ratio[R], contributionToI[G] * ratio[G], contributionToI[B] * ratio[B]]
        const [y, z] = Eye.RGBtoAnchorOrder(Ir)
        to.chan[h2] = i / (Ir[y] + Ir[z])
        to.chan[h1] = to.chan[h2] * ratio[h1]
      } else { // (this.space === CS.EyeNormJ)
        to.chan[h] = this.j
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
      const I = [contributionToI[R] * rgb[R], contributionToI[G] * rgb[G], contributionToI[B] * rgb[B]]
      const [y, z] = Eye.RGBtoAnchorOrder(I)
      this.i = I[y] + I[z]
    } else {
      // due to our definitions of r and p, this value turns out to be the ratio between i (no matter how i is calculated) and the maximum i that could be reached with r and p constant
      const h = Eye.RGBtoHighestChannel(rgb)
      this.j = rgb[h]
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
  protected static lowestChanToAnchors (lowestChan: Channel) : Order {
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
    return (rgb[G] < rgb[B]) ? [B, R] : [G, R] // note that white, grey and black return [G, R], which is what we want (they are the primaries bringing the most brightness to the human eye)
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

const colors : Array<Array<Array<{rgb: RGB, eye: Eye}>>> = []
const numR = 10
const numP = 20
const numI = 40
const color = new Eye(CS.EyePercI)

for (let r = 0; r <= numR; r++) {
  colors.push([])
  for (let p = 0; p <= numP; p++) {
    colors[r].push([])
    for (let i = 0; i <= numI; i++) {
      if (i > 0 && (i / numI) > color.iMax) { break }
      color.r = r / numR
      color.p = p / numP
      color.i = i / numI
      colors[r][p].push({ rgb: color.export(CS.RGBgamma) as RGB, eye: color.export(CS.EyePercI) as Eye })
    }
  }
}

const values: number[] = []

for (let i = 0; i < 1; i += 0.05) {
  values.push(Math.round((i ** 0.4545) * 255))
}

</script>

<template>
  <div style="background-color: rgb(186,186,186)">
    <div v-for="(rRow,r) of colors" :key="r">
      <div v-for="(pRow,p) of rRow" :key="p">
        <div v-for="(c,i) of pRow" :key="i" style="display: inline-block; width: 20px; height: 20px;" :style="'background-color: rgb(' + c.rgb.chan[R] + ',' + c.rgb.chan[G] + ',' + c.rgb.chan[B] + ')'" />
      </div>
      <br>
    </div>
    <br>
    <span style="border: 1px solid black">
      <div v-for="i of values" :key="i" style="display: inline-block; width: 20px; height: 20px;" :style="'background-color: rgb(' + i + ',' + i + ',' + i + ')'" />
    </span>
  </div>
</template>

<style lang="scss" scoped>
</style>
