export enum CS {
  /** RGB values between 0-1 and linear with respect to light intensity */
  RGBlinear,
  /** RGB values between 0-255 and including a gamma exponent of 2.2 (the standard way to store images) */
  RGBgamma,
  /** RPI values (rainbow color, purity, intensity) whose `i` follows the intensity of the light that a human eye
   * perceive (for a given `i`, all colors that you get with various `r` and `p` produce the same intensity for the
   * human eye). Caution: some values of `i` are out-of-range for certain `r` and `p` values (meaning that they would
   * correspond to RGB values greater than 255). */
  EyePercI,
  /** RPJ values (rainbow color, purity, intensity) whose `j` is equivalent to `i/iMax` in the `CS.EyePercI` variant, so
   * `j` is free to take any value between 0 and 1 (so, for a given `j`, the colors that you get with various `r` and
   * `p` produce different intensities for the human eye, whereas the `CS.EyePercI` variant ensures that `i` represents
   * a constant perceived intensity whatever the color is) */
  EyeNormJ,
}

type Channel = 0 | 1 | 2
export const R: Channel = 0
export const G: Channel = 1
export const B: Channel = 2
const gamma = 2.2
const gammaInv = 1 / gamma
const SmallestValue = 0.95 * (1 / 255) ** gamma

/** Classical color space in two variants (depending on the parameter given to the constructor) :
 *  - either the values are between 0-1 and linear with respect to light intensity,
 *  - or the values are between 0-255 and include a gamma exponent of 2.2 (the standard way to store images). */
export class RGB {
  readonly space: CS
  /** Read and write individually the values of the primary channels here. You cannot assign an array directly.
   * To assign a whole array, use method `import()`. */
  readonly chan: number[]

  /** @param space if `CS.RGBlinear` is given, the values will be between 0-1 and linear with respect to light
   * intensity;
   * if `CS.RGBgamma` is given, the values will be between 0-255 and include a gamma exponent of 2.2 (the standard way
   * to store images). */
  constructor(space: CS) {
    if (space !== CS.RGBlinear && space !== CS.RGBgamma) {
      throw new Error('a RGB object can carry RGB information only')
    }
    this.space = space
    this.chan = [0, 0, 0]
  }

  /** Copies a RGB or Eye instance into the current RGB instance. A regular array of 3 values can also be given (in this
   * order: R,G,B).
   * For RGB and Eye objects, color spaces are automatically converted into the color space of the current RGB instance
   * if they differ.
   * If a regular array of numbers is given, its values are expected to be compatible with the color space of the
   * current instance (either linear RGB or gamma-shaped RGB).
   * @returns the instance that you import into (so not the parameter) */
  import(from: number[] | RGB | Eye): RGB {
    if (Array.isArray(from) || from.space === this.space) {
      if (!Array.isArray(from)) {
        from = (from as RGB).chan
      }
      for (const k of [R, G, B]) {
        this.chan[k] = from[k]
      }
    }
    else {
      switch (from.space) {
        case CS.RGBlinear :
          for (const k of [R, G, B]) {
            this.chan[k] = Math.round(((from as RGB).chan[k] ** gammaInv) * 255)
          }
          break
        case CS.RGBgamma :
          for (const k of [R, G, B]) {
            this.chan[k] = ((from as RGB).chan[k] / 255) ** gamma
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
   * The color space of the current instance is automatically converted into the color space of the target instance if
   * needed.
   * @param to existing instance to fill, or if the identifier of a color space is given instead of an object, `export`
   * creates an instance for you, fills it and returns it.
   * @returns target instance (same as `to` if `to` was an instance) */
  export(to: RGB | Eye | CS): RGB & Eye {
    if (typeof to !== 'object') {
      to = (to === CS.RGBlinear || to === CS.RGBgamma) ? new RGB(to) : new Eye(to)
    }
    to.import(this)
    return to as RGB & Eye
  }

  /** corrects channel values that are not within the limits of the format (0-1 or 0-255) */
  limit(): void {
    const max = (this.space === CS.RGBlinear) ? 1 : 255
    for (const k of [R, G, B]) {
      if (this.chan[k] < 0) {
        this.chan[k] = 0
      }
      else if (this.chan[k] > max) {
        this.chan[k] = max
      }
    }
  }
}

/** Color space close to human perception, in two variants (depending on the parameter given to the constructor) :
 * - either the intensity of the light is stored in `i` and follows what a human eye perceives (so a blue and a green
 * having the same `i` will feel as luminous as each other),
 * - or it is stored in `j` and is normalized so it can take any value between 0 and 1. */
export class Eye {
  readonly space: CS
  /** Read and write the perceived rainbow color here. It indicates where the color is on the rainbow. */
  r: number
  /** Read and write the perceived purity here. It indicates how much of the pure color `r` is added to white. */
  p: number
  /** Read and write the perceived intensity of the light here. It is not normalized because `r` and `p` determine the
   * maximum intensity that can be perceived and it is often less than 1.
   * If needed, property `iMax` tells the maximum value that `i` can hold for the current values of `r` and `p`.
   * `i` makes sense only if `CS.EyePercI` has been passed to the constructor, otherwise its value is undefined and
   * setting a value has no effect. */
  i: number
  /** Read and write the normalized intensity of the light here. Any value between 0 and 1 is possible.
   * `j` makes sense only if `CS.EyeNormJ` has been passed to the constructor, otherwise its value is undefined and
   * setting a value has no effect. */
  j: number
  /** Maximum value that `i` can have for the current values of `r` and `p`. The intensity that a human perceives
   *  depends on the color. For example, in RGB (so on all monitors) a 255 blue looks less bright than a 255 green so
   *  the `iMax` of blue is lower than the `iMax` of green.
   *  If you set `i` to a value higher than `iMax`, the color will correspond to RGB values greater than 255.
   * `iMax` is kept up-to-date automatically when you change `r` or `p`. */
  get iMax(): number {
    if (this.Imax.rOfValue !== this.r || this.Imax.pOfValue !== this.p) {
      /* calculates the RGB values for the current r and p, with a standardized intensity value (the smallest iMax
       * possible) so the RGB cannot be black */
      this.convertToRGB(Eye.rgbLinear, Eye.lowestImax)
      const h = Eye.RGBtoHighestChannel(Eye.rgbLinear.chan)
      this.Imax.value = Eye.lowestImax / (Eye.rgbLinear.chan[h] ** gammaInv)
      this.Imax.rOfValue = this.r
      this.Imax.pOfValue = this.p
    }
    return this.Imax.value
  }

  // constants of our perception model, all obtained empirically
  /* sensitivity of the human eye to primaries, used to calculate the perceived intensity when channels add */
  protected static readonly sensicol = [15, 20, 5]
  /* controls the linearity of the perceived color with respect to r when primaries mix together to form a pure
     intermediary */
  protected static readonly rPowers = [1.2, 1.6, 0.9]
  /* perceived ability of the primaries to tint a white light when added to it (controls the width of the the grey part
     in a row where the purity goes from 0 to 1) */
  protected static readonly overwhite = [0.6, 1, 0.6]
  /* when the primaries of a given color are ordered by perceived intensities (so by `value * sensicol`), tells how
     much the second perceived dimmest contribute to the perceived intensity of the mix of the three */
  protected static readonly iotaM = 0.15
  /* when the primaries of a given color are ordered by perceived intensities(so by `value * sensicol`), tells how much
     the perceived dimmest contribute to the perceived intensity of the mix of the three */
  protected static readonly iotaD = 0.15
  // the next 2 constants will be filled by the constructor
  protected static readonly sensicolNorm = [0, 0, 0]
  protected static readonly rPowersInv = [0, 0, 0]
  protected static readonly Idivider = Eye.sensicol[R] * Eye.iotaM + Eye.sensicol[G] + Eye.sensicol[B] * Eye.iotaD

  /** Among all possible colors, this is the lowest `iMax` than can be met. In other words, the intensity `i` can be set
   * to `lowestImax` for any `r` and `p`. Greater values of `i` will be impossible for some colors. */
  static readonly lowestImax = (Eye.sensicol[B] / Eye.Idivider) ** gammaInv

  /**
   * @param space if `CS.EyePercI` is given, the intensity of the light will be stored in `i` and follow what a human
   * eye perceives; if `CS.EyeNormJ` is given, the intensity will be stored in `j` and normalized so it can take any
   * value between 0 and 1. */
  constructor(space: CS) {
    if (space !== CS.EyePercI && space !== CS.EyeNormJ) {
      throw new Error('an Eye object can carry RPI/J information only')
    }
    if (!Eye.sensicolNorm[0]) {
      for (const k of [R, G, B]) {
        Eye.sensicolNorm[k] = Eye.sensicol[k] / Eye.Idivider
        Eye.rPowersInv[k] = 1 / Eye.rPowers[k]
      }
    }
    this.space = space
    this.r = this.p = this.i = this.j = 0
  }

  private Imax = {
    value: 0,
    rOfValue: -1,
    pOfValue: -1,
  }

  protected static rgbLinear = new RGB(CS.RGBlinear)
  protected static rgbGamma = new RGB(CS.RGBgamma)

  /** Copies a RGB or Eye object or an array of values into the current Eye instance.
   * Color spaces are automatically converted into the color space of the current Eye instance if they differ.
   * If an array of values is given, it is assumed to contain gamma-shaped RGB.
   * @returns the instance that you import into (so not the parameter) */
  import(from: RGB | Eye | number[]): Eye {
    if (Array.isArray(from)) {
      from = Eye.rgbGamma.import(from)
    }
    if (from.space === CS.RGBgamma) {
      from = Eye.rgbLinear.import(from)
    }
    if (from.space === CS.RGBlinear) {
      // conversion of RGB into Eye
      const rgb = (from as RGB).chan
      const l = Eye.RGBtoLowestChannel(rgb)
      const [h1, h2] = Eye.lowestChanToAnchors(l)
      const anchor2Contribution = (rgb[h2] - rgb[l])
      const anchors12Contributions = (rgb[h1] - rgb[l]) + anchor2Contribution
      if (anchors12Contributions < SmallestValue) { // the 3 channels are equal, so we have black, grey or white
        this.r = 0.5
        this.p = 0
      }
      else {
        let d = anchor2Contribution / anchors12Contributions
        if (d < 1 / 2) {
          d = (2 * d) ** Eye.rPowers[h1] / 2
        }
        else {
          d = (2 * d - 1) ** Eye.rPowersInv[h2] / 2 + 1 / 2
        }
        this.r = (h1 + d) / 3
        const w = (rgb[h1] - rgb[l]) * Eye.overwhite[h1] + (rgb[h2] - rgb[l]) * Eye.overwhite[h2]
        this.p = w / (rgb[l] + w)
      }
      this.fillIntensityFromRGB(rgb)
    }
    else {
      from = from as Eye // for the static checker
      this.r = from.r
      this.p = from.p
      this.i = from.i
      this.j = from.j
      this.Imax.value = from.Imax.value
      this.Imax.rOfValue = from.Imax.rOfValue
      this.Imax.pOfValue = from.Imax.pOfValue
      if (from.space !== this.space) {
        if (this.space === CS.EyePercI) {
          this.i = this.j * this.iMax
        }
        else {
          this.j = this.i / from.iMax
        }
      }
    }
    return this
  }

  /** Copies this Eye instance into another RGB or Eye instance, or in a gamma-shaped RGB array.
   * The color space of the current instance is automatically converted into the color space of the target instance if
   * needed.
   * @param to existing instance or array to fill, or if the identifier of a color space is given instead of an object,
   * `export` creates an instance for you, fills it and returns it.
   * @returns target instance or array (same as `to` if `to` was an instance or an array) */
  export(to: RGB | Eye | CS | number[]): RGB & Eye & number[] {
    if (Array.isArray(to)) {
      this.convertToRGB(Eye.rgbLinear, this.i)
      Eye.rgbGamma.import(Eye.rgbLinear)
      for (const k of [R, G, B]) {
        to[k] = Eye.rgbGamma.chan[k]
      }
    }
    else {
      if (typeof to !== 'object') {
        to = (to === CS.RGBlinear || to === CS.RGBgamma) ? new RGB(to) : new Eye(to)
      }
      if (to.space === CS.EyePercI || to.space === CS.EyeNormJ) {
        to.import(this)
      }
      else
        if (to.space === CS.RGBgamma) {
          this.convertToRGB(Eye.rgbLinear, this.i)
          to.import(Eye.rgbLinear)
        }
        else {
          this.convertToRGB(to as RGB, this.i)
        }
    }
    return to as RGB & Eye & number[]
  }

  /* shifts `r` into [0-1] and trims `p` and `i` or `j` if they exceed their respective limits */
  limit(): void {
    if (this.r < 0) {
      this.r = 1 - ((-this.r) % 1)
    }
    else if (this.r > 1) {
      this.r = this.r % 1
    }
    if (this.p < 0) {
      this.p = 0
    }
    else if (this.p > 1) {
      this.p = 1
    }
    if (this.space === CS.EyePercI) {
      if (this.i < 0) {
        this.i = 0
      }
      else if (this.i > this.iMax) {
        this.i = this.iMax
      }
    }
    else {
      if (this.j < 0) {
        this.j = 0
      }
      else if (this.j > 1) {
        this.j = 1
      }
    }
  }

  protected convertToRGB(to: RGB, i: number): void {
    const [l, , h] = Eye.rToChannelOrder(this.r)
    const [h1, h2] = Eye.lowestChanToAnchors(l)
    const ratio = [0, 0, 0]
    const p = this.p
    let d = 3 * this.r - h1
    if (d < 1 / 2) {
      d = (2 * d) ** Eye.rPowersInv[h1] / 2
    }
    else {
      d = (2 * d - 1) ** Eye.rPowers[h2] / 2 + 1 / 2
    }
    if ((d < SmallestValue && p > (1 - SmallestValue)) || (this.space === CS.EyePercI && i < SmallestValue)
      || (this.space === CS.EyeNormJ && this.j < SmallestValue)) {
      to.chan[l] = 0
      to.chan[h1] = (this.space === CS.EyePercI) ? i ** gamma / Eye.sensicolNorm[h1] : this.j ** gamma
      to.chan[h2] = 0
    }
    else {
      const S = (1 - p) * (d * Eye.overwhite[h2] + (1 - d) * Eye.overwhite[h1])
      const T = p * d
      const U = S + T
      ratio[l] = S / U
      ratio[h1] = (p - T) / U + ratio[l]
      ratio[h2] = 1
      if (this.space === CS.EyePercI) {
        const Ir = [Eye.sensicolNorm[R] * ratio[R], Eye.sensicolNorm[G] * ratio[G], Eye.sensicolNorm[B] * ratio[B]]
        const [x, y, z] = Eye.RGBtoChannelOrder(Ir)
        to.chan[h2] = i ** gamma / (Ir[x] * Eye.iotaD + Ir[y] * Eye.iotaM + Ir[z])
        to.chan[h1] = to.chan[h2] * ratio[h1]
      }
      else {
        to.chan[h] = this.j ** gamma
        if (h === h2) {
          to.chan[h1] = to.chan[h2] * ratio[h1]
        }
        else {
          to.chan[h2] = to.chan[h1] / ratio[h1] // if h ≠ h2 then h1 is the highest channel so ratio[h1] > 0
        }
      }
      to.chan[l] = to.chan[h2] * ratio[l]
    }
  }

  protected fillIntensityFromRGB(rgb: number[]): void {
    if (this.space === CS.EyePercI) {
      const I = [Eye.sensicolNorm[R] * rgb[R], Eye.sensicolNorm[G] * rgb[G], Eye.sensicolNorm[B] * rgb[B]]
      const [x, y, z] = Eye.RGBtoChannelOrder(I)
      this.i = (I[x] * Eye.iotaD + I[y] * Eye.iotaM + I[z]) ** gammaInv
    }
    else {
      const h = Eye.RGBtoHighestChannel(rgb)
      // due to our definitions of r and p, rgb[h] turns out to be the ratio between i (no matter how i is calculated)
      // and the maximum i that could be reached with r and p constant
      this.j = rgb[h] ** gammaInv
    }
  }

  /** @returns the channel carrying the lowest value. In case of equality(ies), B is preferred over R and R is preferred
   * over G. */
  protected static RGBtoLowestChannel(rgb: number[]): Channel {
    if (rgb[R] <= rgb[G]) {
      if (rgb[R] < rgb[B]) {
        return R
      }
    }
    else
      if (rgb[G] < rgb[B]) {
        return G
      }
    return B
  }

  /** @returns the channel carrying the highest value. In case of equality(ies), G is preferred over R and R is
   * preferred over B. */
  protected static RGBtoHighestChannel(rgb: number[]): Channel {
    if (rgb[R] > rgb[G]) {
      if (rgb[R] >= rgb[B]) {
        return R
      }
    }
    else
      if (rgb[G] >= rgb[B]) {
        return G
      }
    return B
  }

  /** @returns the order of the channels from the lowest value to the highest-value. In case of equality(ies), B is
   * placed before R and R is placed before G. */
  protected static RGBtoChannelOrder(rgb: number[]): Channel[] {
    if (rgb[R] <= rgb[G]) {
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

  /** @returns the order of the channels from the lowest value to the highest-value. In case of equality(ies), B is
   * placed before R and R is placed before G. */
  protected static rToChannelOrder(r: number): Channel[] {
    if (r <= 2 / 6) {
      return (r < 1 / 6) ? [B, G, R] : [B, R, G]
    }
    if (r <= 4 / 6) {
      return (r <= 3 / 6) ? [R, B, G] : [R, G, B]
    }
    return (r < 5 / 6) ? [G, R, B] : [G, B, R]
  }

  /** @returns the anchors in the same order as on the rainbow (note that R is both before G and after B) */
  protected static lowestChanToAnchors(lowestChan: Channel): Channel[] {
    switch (lowestChan) {
      case R : return [G, B]
      case G : return [B, R]
      case B : return [R, G]
    }
    return [] // impossible but the static checker believes it can happen
  }
}

export function distance(eye1: Eye, eye2: Eye): number {
  function colorOnSlice(e: Eye): number[] {
    const i = (e.space === CS.EyePercI) ? e.i : e.j * e.iMax
    const pPi3 = Math.PI / 3 * e.p
    return [i * Math.sin(pPi3), i * Math.cos(pPi3), 0]
  }
  const [r1, r2] = (eye1.r < eye2.r) ? [eye1.r, eye2.r] : [eye2.r, eye1.r]
  // arbitrary / subjective: a distance of 1/6 on r (so the distance between a primary and its closest secondary) is how
  // much two colors can feel different at most. This results in a distance of 1. Change 6 for 3 to get a value evolving
  // linerarly between a twice larger range (so the distance primary-secondary would be 0.5).
  const rDist = Math.min(6 * Math.min(r2 - r1, 1 + r1 - r2), 1)
  const cosSlices = (2 - rDist * rDist) / 2
  const slice1 = colorOnSlice(eye1)
  const slice2 = colorOnSlice(eye2)
  const x = slice1[0] - slice2[0] * cosSlices
  const y = slice1[1] - slice2[1]
  const z = slice2[0] * Math.sqrt(1 - cosSlices * cosSlices)
  return Math.sqrt(x * x + y * y + z * z)
}

export enum ColorBlindness { Red, Green }
let colorBlindness: ColorBlindness

export function setColorBlindness(blindness: ColorBlindness): void {
  colorBlindness = blindness
}

const rgbPO2D = new RGB(CS.RGBlinear)
const cbSightPO2D = new Eye(CS.EyePercI)

/** @param rgb must be in the CS.RGBlinear format
 *  @param rgbWorkingBuffer if this function is used by threads, each thread must provide one `RGB(CS.RGBlinear)` object
 *  @param cbWorkingBuffer if this function is used by threads, each thread must provide one `Eye(CS.EyePercI)` object
 *  @returns the projection of `rgb` into the Eye color-space of the color-blind person */
export function projectOntoCBPlane(rgb: number[], rgbWorkingBuffer?: RGB, cbWorkingBuffer?: Eye): Eye {
  if (!rgbWorkingBuffer || !cbWorkingBuffer) {
    rgbWorkingBuffer = rgbPO2D
    cbWorkingBuffer = cbSightPO2D
  }
  switch (colorBlindness) {
    /* The matrices have been calculated by following the prodedure of
     * Viénot, F., Brettel, H., & Mollon, J. D. (1999) "Digital video colourmaps for checking the legibility of
     * displays by dichromats". Color Research and Application, 24, 4, 243-251. */
    case ColorBlindness.Red :
      rgbWorkingBuffer.chan[R] = 0.1123822674257492 * rgb[R] + 0.8876119706946073 * rgb[G]
      rgbWorkingBuffer.chan[G] = rgbWorkingBuffer.chan[R]
      rgbWorkingBuffer.chan[B] = 0.00400576009958313 * rgb[R] - 0.004005734096601488 * rgb[G] + rgb[B]
      break
    case ColorBlindness.Green :
      for (const k of [R, G, B]) {
        rgbWorkingBuffer.chan[k] = 0.99 * rgb[k] + 0.005
      }
      rgbWorkingBuffer.chan[B] = -0.02233647741916585 * rgbWorkingBuffer.chan[R]
      + 0.02233656063451689 * rgbWorkingBuffer.chan[G] + rgbWorkingBuffer.chan[B]
      rgbWorkingBuffer.chan[G] = 0.292750775976202 * rgbWorkingBuffer.chan[R]
      + 0.7072518589062524 * rgbWorkingBuffer.chan[G]
      rgbWorkingBuffer.chan[R] = rgbWorkingBuffer.chan[G]
      for (const k of [R, G, B]) {
        rgbWorkingBuffer.chan[k] = 0.99 * rgbWorkingBuffer.chan[k] + 0.005
      }
      break
    default :
      rgbWorkingBuffer.import(rgb)
      break
  }
  // in rare cases, a value can happen to be slightly below 0 or slightly above 1 (the 0—255 range becomes
  // approximately -3—258).
  rgbWorkingBuffer.limit()
  cbWorkingBuffer.import(rgbWorkingBuffer)
  return cbWorkingBuffer
}
