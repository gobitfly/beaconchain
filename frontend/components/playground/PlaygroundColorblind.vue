<script lang="ts" setup>
import { warn } from 'vue'

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
const SmallestValue = 0.95 * (1 / 255) ** gamma

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
      for (const k of [R, G, B]) {
        this.chan[k] = from[k]
      }
    } else {
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
   * The color space of the current instance is automatically converted into the color space of the target instance if needed.
   * @param to existing instance to fill, or if the identifier of a color space is given instead of an object, `export` creates an instance for you, fills it and returns it.
   * @returns target instance (same as `to` if `to` was an instance) */
  export (to: RGB | Eye | CS) : RGB & Eye {
    if (typeof to !== 'object') {
      to = (to === CS.RGBlinear || to === CS.RGBgamma) ? new RGB(to) : new Eye(to)
    }
    to.import(this)
    return to as RGB & Eye
  }

  /** corrects channel values that are not within the limits of the format (0-1 or 0-255) */
  limit () : void {
    const max = (this.space === CS.RGBlinear) ? 1 : 255
    for (const k of [R, G, B]) {
      if (this.chan[k] < 0) { this.chan[k] = 0 }
      if (this.chan[k] > max) { this.chan[k] = max }
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
      this.convertToRGB(Eye.rgbLinear, Eye.lowestImax) // calculates the RGB values for the current r and p, with a standardized intensity value (the smallest iMax possible) so the RGB cannot be black
      const h = Eye.RGBtoHighestChannel(Eye.rgbLinear.chan)
      this.Imax.value = Eye.lowestImax / (Eye.rgbLinear.chan[h] ** gammaInv)
      this.Imax.rOfValue = this.r
      this.Imax.pOfValue = this.p
    }
    return this.Imax.value
  }

  // constants of our perception model, all obtained empirically
  protected static readonly sensicol = [15, 20, 5] // sensitivity of the human eye to primaries, used to calculate the perceived intensity when channels add
  protected static readonly rPowers = [1.2, 1.6, 0.9] // controls the linearity of the perceived color with respect to r when primaries mix together to form a pure intermediary
  protected static readonly overwhite = [0.6, 1, 0.6] // perceived ability of the primaries to tint a white light when added to it (controls the width of the the grey part in a row where the purity goes from 0 to 1)
  protected static readonly iotaM = 0.15 // when the primaries of a given color are ordered by perceived intensities (so by `value * sensicol`), tells how much the second perceived dimmest contribute to the perceived intensity of the mix of the three
  protected static readonly iotaD = 0.15 // when the primaries of a given color are ordered by perceived intensities (so by `value * sensicol`), tells how much the perceived dimmest contribute to the perceived intensity of the mix of the three
  // the next 2 constants will be filled by the constructor
  protected static readonly sensicolNorm = [0, 0, 0]
  protected static readonly rPowersInv = [0, 0, 0]
  protected static readonly Idivider = Eye.sensicol[R] * Eye.iotaM + Eye.sensicol[G] + Eye.sensicol[B] * Eye.iotaD

  /** Among all possible colors, this is the lowest `iMax` than can be met. In other words, the intensity `i` can be set to `lowestImax` for any `r` and `p`. Greater values of `i` will be impossible for some colors. */
  static readonly lowestImax = (Eye.sensicol[B] / Eye.Idivider) ** gammaInv

  /**
   * @param space if `CS.EyePercI` is given, the intensity of the light will be stored in `i` and follow what a human eye perceives;
   * if `CS.EyeNormJ` is given, the intensity will be stored in `j` and normalized so it can take any value between 0 and 1. */
  constructor (space: CS) {
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
    pOfValue: -1
  }

  protected static rgbLinear = new RGB(CS.RGBlinear)
  protected static rgbGamma = new RGB(CS.RGBgamma)

  /** Copies a RGB or Eye object or an array of values into the current Eye instance.
   * Color spaces are automatically converted into the color space of the current Eye instance if they differ.
   * If an array of values is given, it is assumed to contain gamma-shaped RGB.
   * @returns the instance that you import into (so not the parameter) */
  import (from: RGB | Eye | number[]) : Eye {
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
      } else {
        let d = anchor2Contribution / anchors12Contributions
        if (d < 1 / 2) {
          d = (2 * d) ** Eye.rPowers[h1] / 2
        } else {
          d = (2 * d - 1) ** Eye.rPowersInv[h2] / 2 + 1 / 2
        }
        this.r = (h1 + d) / 3
        const w = (rgb[h1] - rgb[l]) * Eye.overwhite[h1] + (rgb[h2] - rgb[l]) * Eye.overwhite[h2]
        this.p = w / (rgb[l] + w)
      }
      this.fillIntensityFromRGB(rgb)
    } else {
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
        } else {
          this.j = this.i / from.iMax
        }
      }
    }
    return this
  }

  /** Copies this Eye instance into another RGB or Eye instance, or in a gamma-shaped RGB array.
   * The color space of the current instance is automatically converted into the color space of the target instance if needed.
   * @param to existing instance or array to fill, or if the identifier of a color space is given instead of an object, `export` creates an instance for you, fills it and returns it.
   * @returns target instance or array (same as `to` if `to` was an instance or an array) */
  export (to: RGB | Eye | CS | number[]) : RGB & Eye & number[] {
    if (Array.isArray(to)) {
      this.convertToRGB(Eye.rgbLinear, this.i)
      Eye.rgbGamma.import(Eye.rgbLinear)
      for (const k of [R, G, B]) {
        to[k] = Eye.rgbGamma.chan[k]
      }
    } else {
      if (typeof to !== 'object') {
        to = (to === CS.RGBlinear || to === CS.RGBgamma) ? new RGB(to) : new Eye(to)
      }
      if (to.space === CS.EyePercI || to.space === CS.EyeNormJ) {
        to.import(this)
      } else
        if (to.space === CS.RGBgamma) {
          this.convertToRGB(Eye.rgbLinear, this.i)
          to.import(Eye.rgbLinear)
        } else {
          this.convertToRGB(to as RGB, this.i)
        }
    }
    return to as RGB & Eye & number[]
  }

  protected convertToRGB (to: RGB, i: number) : void {
    const [l, , h] = Eye.rToChannelOrder(this.r)
    const [h1, h2] = Eye.lowestChanToAnchors(l)
    const ratio = [0, 0, 0]
    const p = this.p
    let d = 3 * this.r - h1
    if (d < 1 / 2) {
      d = (2 * d) ** Eye.rPowersInv[h1] / 2
    } else {
      d = (2 * d - 1) ** Eye.rPowers[h2] / 2 + 1 / 2
    }
    if ((d < SmallestValue && p > (1 - SmallestValue)) || (this.space === CS.EyePercI && i < SmallestValue) || (this.space === CS.EyeNormJ && this.j < SmallestValue)) {
      to.chan[l] = 0
      to.chan[h1] = (this.space === CS.EyePercI) ? i ** gamma / Eye.sensicolNorm[h1] : this.j ** gamma
      to.chan[h2] = 0
    } else {
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
      } else {
        to.chan[h] = this.j ** gamma
        if (h === h2) {
          to.chan[h1] = to.chan[h2] * ratio[h1]
        } else {
          to.chan[h2] = to.chan[h1] / ratio[h1] // if h ≠ h2 then h1 is the highest channel so ratio[h1] > 0
        }
      }
      to.chan[l] = to.chan[h2] * ratio[l]
    }
  }

  protected fillIntensityFromRGB (rgb : number[]) : void {
    if (this.space === CS.EyePercI) {
      const I = [Eye.sensicolNorm[R] * rgb[R], Eye.sensicolNorm[G] * rgb[G], Eye.sensicolNorm[B] * rgb[B]]
      const [x, y, z] = Eye.RGBtoChannelOrder(I)
      this.i = (I[x] * Eye.iotaD + I[y] * Eye.iotaM + I[z]) ** gammaInv
    } else {
      const h = Eye.RGBtoHighestChannel(rgb)
      // due to our definitions of r and p, rgb[h] turns out to be the ratio between i (no matter how i is calculated) and the maximum i that could be reached with r and p constant
      this.j = rgb[h] ** gammaInv
    }
  }

  /** @returns the channel carrying the lowest value. In case of equality(ies), B is preferred over R and R is preferred over G. */
  protected static RGBtoLowestChannel (rgb : number[]) : Channel {
    if (rgb[R] <= rgb[G]) {
      if (rgb[R] < rgb[B]) { return R }
    } else
      if (rgb[G] < rgb[B]) { return G }
    return B
  }

  /** @returns the channel carrying the highest value. In case of equality(ies), G is preferred over R and R is preferred over B. */
  protected static RGBtoHighestChannel (rgb : number[]) : Channel {
    if (rgb[R] > rgb[G]) {
      if (rgb[R] >= rgb[B]) { return R }
    } else
      if (rgb[G] >= rgb[B]) { return G }
    return B
  }

  /** @returns the order of the channels from the lowest value to the highest-value. In case of equality(ies), B is placed before R and R is placed before G. */
  protected static RGBtoChannelOrder (rgb : number[]) : Channel[] {
    if (rgb[R] <= rgb[G]) {
      if (rgb[G] < rgb[B]) { return [R, G, B] }
      return (rgb[R] < rgb[B]) ? [R, B, G] : [B, R, G]
    }
    if (rgb[R] < rgb[B]) { return [G, R, B] }
    return (rgb[G] < rgb[B]) ? [G, B, R] : [B, G, R]
  }

  /** @returns the order of the channels from the lowest value to the highest-value. In case of equality(ies), B is placed before R and R is placed before G. */
  protected static rToChannelOrder (r : number) : Channel[] {
    if (r <= 2 / 6) {
      return (r < 1 / 6) ? [B, G, R] : [B, R, G]
    }
    if (r <= 4 / 6) {
      return (r <= 3 / 6) ? [R, B, G] : [R, G, B]
    }
    return (r < 5 / 6) ? [G, R, B] : [G, B, R]
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
}

function distance (eye1: Eye, eye2: Eye) : number {
  function colorOnSlice (e: Eye) : number[] {
    const i = (e.space === CS.EyePercI) ? e.i : e.j * e.iMax
    const pPi3 = Math.PI / 3 * e.p
    return [i * Math.sin(pPi3), i * Math.cos(pPi3), 0]
  }
  const [r1, r2] = (eye1.r < eye2.r) ? [eye1.r, eye2.r] : [eye2.r, eye1.r]
  const rDist = Math.min(6 * Math.min(r2 - r1, 1 + r1 - r2), 1) // arbitrary / subjective: a distance of 1/6 on r (so the distance between a primary and its closest secondary) is how much two colors can feel different at most. This results in a distance of 1. Change 6 for 3 to get a value evolving linerarly between a twice larger range (so the distance primary-secondary would be 0.5).
  const cosSlices = (2 - rDist * rDist) / 2
  const slice1 = colorOnSlice(eye1)
  const slice2 = colorOnSlice(eye2)
  const x = slice1[0] - slice2[0] * cosSlices
  const y = slice1[1] - slice2[1]
  const z = slice2[0] * Math.sqrt(1 - cosSlices * cosSlices)
  return Math.sqrt(x * x + y * y + z * z)
}

type ColorDefinition = { color: string, identifier: string }
enum ColorBlindness { Red, Green }

const timeAllowed = 200 // ms
/** If `privilege``is between 0 and 1,
 *   the shorter distances are better reproduced, but the long distances can change noticeably, which is a problem for those getting too short.
 *  If `privilege` is above 1,
 *   the long distances are better reproduced, but the short distances can get shorter or longer than they were. */
const privilege = 1
/** maximum number of colors that we do not want to touch temporarily */
const tabuLength = 7

function enhanceColors (colors: ColorDefinition[], colorBlindness: ColorBlindness) : ColorDefinition[] {
  const original = []
  for (const def of colors) {
    const i = def.color.indexOf('#')
    if (i < 0) {
      warn('Color', def.color, 'is given in an unknown format.')
      return colors
    }
    const col = parseInt(def.color.slice(i + 1), 16)
    const rgb = new RGB(CS.RGBgamma)
    rgb.chan[B] = col & 0xFF
    rgb.chan[G] = (col >> 8) & 0xFF
    rgb.chan[R] = (col >> 16) & 0xFF
    original.push(rgb)
  }

  const enhanced = search(original, colorBlindness)

  const result: ColorDefinition[] = []
  for (let c = 0; c < colors.length; c++) {
    let hex = (enhanced[c].chan[B] | enhanced[c].chan[G] << 8 | enhanced[c].chan[R] << 16).toString(16)
    hex = ('000000' + hex).slice(-6)
    result.push({
      color: '#' + hex,
      identifier: colors[c].identifier
    })
  }
  return result
}

let blindness: ColorBlindness
let original: RGB[] // CS.RGBlinear
let distancesOrig: number[][]
let wip3D: RGB[] // CS.RGBlinear
let wip2D: Eye[] // CS.EyePercI
let errors2D: number[]
const temp = new RGB(CS.RGBlinear)
const tabuQueue = new Array<number>(tabuLength) // queue of color indices that we do not want to touch for some time

/** @param input must be in the CS.RGBgamma format.
 * @returns enchanced colors in the CS.RGBgamma format
 */
function search (input : RGB[], colorBlindness: ColorBlindness) : RGB[] {
  const endTime = performance.now() + timeAllowed
  blindness = colorBlindness
  original = input.map(col => col.export(CS.RGBlinear))
  // storing in distancesOrig the distances between the original colors: the algorithm will try to reproduce them on the 2D plane
  distancesOrig = []
  const originalEye = original.map(col => col.export(CS.EyePercI))
  const line: number[] = []
  for (const colA of originalEye) {
    line.length = 0
    for (const colB of originalEye) {
      line.push(distance(colA, colB))
    }
    distancesOrig.push(line)
  }
  // copying the original colors into wip3D: they will be modified progressively by the search phase in such a way that they get more and more distinguishable when viewed by a color blind person
  wip3D = original.map(col => col.export(CS.RGBlinear))
  // projecting the original colors into wip2D: they are the starting point and the search phase will make them more and more distinguishable at each iteration
  wip2D = wip3D.map(col => projectOnto2D(col.chan).export(CS.EyePercI))
  errors2D = wip2D.map((col, k) => distError(k, col))

  // search phase
  tabuQueue.fill(-1)
  while (performance.now() < endTime) {
    optimizeOneStepFurther()
  }
  // 🪄✨
  return wip3D.map(col => col.export(CS.RGBgamma))
}

function optimizeOneStepFurther () {
  let bestK: number = 0
  let bestError: number = 0
  let bestErrorGain: number = Number.MAX_SAFE_INTEGER
  let bestColor: number[] = []
  const step = 2 /// /////////////////////////////  TODO: change the step dynamically
  for (let k = 0; k < wip3D.length; k++) {
    if (tabuQueue.includes(k)) { continue }
    temp.import(wip3D[k])
    for (const c of [R, G, B]) {
      for (const s of [-step, +step]) {
        const restoredValue = temp.chan[c]
        temp.chan[c] += s
        const error = distError(k, projectOnto2D(temp.chan))
        if (error - errors2D[k] < bestErrorGain) {
          bestErrorGain = error - errors2D[k]
          bestError = error
          bestK = k
          bestColor = [...temp.chan]
        }
        temp.chan[c] = restoredValue
      }
    }
  }
  if (bestErrorGain < Number.MAX_SAFE_INTEGER) {
    wip3D[bestK].import(bestColor)
    wip2D[bestK].import(projectOnto2D(bestColor))
    errors2D[bestK] = bestError
    tabuQueue.shift()
    if (bestErrorGain < 0) {
      tabuQueue.push(-1)
    } else {
      // to avoid oscillating back and forth on the same color until the time limit expires, we add it to the tabu FIFO
      tabuQueue.push(bestK)
    }
  } else {
    tabuQueue.fill(-1)
  }
}

const cbSight = new Eye(CS.EyePercI)

/** @returns the projection of `rgb` into the Eye color-space of the color-blind person */
function projectOnto2D (rgb: number[]) : Eye {
  switch (blindness) {
    // The matrices have been calculated by following the prodedure of
    // Viénot, F., Brettel, H., & Mollon, J. D. (1999) "Digital video colourmaps for checking the legibility of displays by dichromats". Color Research and Application, 24, 4, 243-251.
    case ColorBlindness.Red :
      temp.chan[R] = 0.1123822674257492 * rgb[R] + 0.8876119706946073 * rgb[G]
      temp.chan[G] = temp.chan[R]
      temp.chan[B] = 0.00400576009958313 * rgb[R] - 0.004005734096601488 * rgb[G] + rgb[B]
      break
    case ColorBlindness.Green :
      for (const k of [R, G, B]) { temp.chan[k] = 0.99 * rgb[k] + 0.005 }
      temp.chan[R] = 0.292750775976202 * temp.chan[R] + 0.7072518589062524 * temp.chan[G]
      temp.chan[G] = temp.chan[R]
      temp.chan[B] = -0.02233647741916585 * temp.chan[R] + 0.02233656063451689 * temp.chan[G] + temp.chan[B]
      for (const k of [R, G, B]) { temp.chan[k] = 0.99 * temp.chan[k] + 0.005 }
      break
    default :
      temp.import(rgb)
      break
  }
  temp.limit() // in rare cases, a value can happen to be slightly below 0 or slightly above 1 (the 0—255 range would become approximately -3—258).
  cbSight.import(temp)
  return cbSight
}

/** for a given color `k` stored in `wipColor2D`, this function calculates a value continuously increasing with respect to
 *    the sum over all colors `l` of
 *      the difference between:
 *        the distance between the original colors k and l
 *        the distance between the projected colors `wipColor2D` and l */
function distError (k: number, wipColor2D: Eye) : number {
  let result = 0
  for (let l = 0; l < distancesOrig.length; l++) {
    if (k === l) { continue }
    result += Math.abs(distancesOrig[k][l] - distance(wipColor2D, wip2D[l])) ** privilege
  }
  return result
}

//
// TESTS AND ADJUSTEMENTS
//

const cons = console

const colorI = new Eye(CS.EyePercI)
const colorJ = new Eye(CS.EyeNormJ)

const colors : Array<Array<Array<{rgb: RGB, eye: Eye}>>> = []
const numR = 6
const numP = 21
const numI = 40
let maxIntensityMinIndex = 1000
for (let r = 0; r <= numR; r++) {
  colors.push([])
  for (let p = 0; p <= numP; p++) {
    colors[r].push([])
    for (let i = 0; i <= numI; i++) {
      if (i > 0 && (i / numI) > colorI.iMax) {
        if (i - 1 < maxIntensityMinIndex) { maxIntensityMinIndex = i - 1 }
        break
      }
      colorI.r = r / numR
      colorI.p = p / numP
      colorI.i = i / numI
      colors[r][p].push({ rgb: colorI.export(CS.RGBgamma), eye: colorI.export(CS.EyePercI) })
      if (colors[r][p][i].rgb.chan[R] > 255 || colors[r][p][i].rgb.chan[G] > 255 || colors[r][p][i].rgb.chan[B] > 255 || colors[r][p][i].rgb.chan[R] < 0 || colors[r][p][i].rgb.chan[G] < 0 || colors[r][p][i].rgb.chan[B] < 0) {
        cons.log('#### PercI -> RGB  out of bounds')
        cons.log(colors[r][p][i].rgb)
      }
      let rgb2 = ((new Eye(CS.EyePercI)).import(colors[r][p][i].rgb)).export(CS.RGBgamma).chan
      if (rgb2[R] > 255 || rgb2[G] > 255 || rgb2[B] > 255 || rgb2[R] < 0 || rgb2[G] < 0 || rgb2[B] < 0) {
        cons.log('#### PercI -> RGB -> PercI -> RGB  out of bounds')
        cons.log(rgb2)
      }
      if (colors[r][p][i].rgb.chan[0] !== rgb2[0] || colors[r][p][i].rgb.chan[1] !== rgb2[1] || colors[r][p][i].rgb.chan[2] !== rgb2[2]) {
        cons.log('#### PercI -> RGB  different from  PercI -> RGB -> PercI -> RGB.')
        cons.log('PercI:', colors[r][p][i].eye)
        cons.log('PercI -> RGB -> PercI :', (new Eye(CS.EyePercI)).import(colors[r][p][i].rgb))
        cons.log('PercI -> RGB :', colors[r][p][i].rgb.chan)
        cons.log('PercI -> RGB -> PercI -> RGB :', rgb2)
      }
      rgb2 = colors[r][p][i].eye.export(CS.RGBlinear).export(CS.EyeNormJ).export(CS.RGBgamma).chan
      if (rgb2[R] > 255 || rgb2[G] > 255 || rgb2[B] > 255 || rgb2[R] < 0 || rgb2[G] < 0 || rgb2[B] < 0) {
        cons.log('#### PercI -> RGB -> NormJ -> RGB  out of bounds')
        cons.log(rgb2)
      }
      if (colors[r][p][i].rgb.chan[0] !== rgb2[0] || colors[r][p][i].rgb.chan[1] !== rgb2[1] || colors[r][p][i].rgb.chan[2] !== rgb2[2]) {
        cons.log('#### PercI -> RGB  different from  PercI -> RGB -> NormJ -> RGB.')
        cons.log('PercI -> RGB :', colors[r][p][i].rgb.chan)
        cons.log('PercI -> RGB -> NormJ -> RGB :', rgb2)
        cons.log('PercI:', colors[r][p][i].eye)
        cons.log('PercI -> RGB -> NormJ :', colors[r][p][i].eye.export(CS.RGBlinear).export(CS.EyeNormJ))
      }
      rgb2 = colors[r][p][i].eye.export(CS.EyeNormJ).export(CS.RGBgamma).chan
      if (rgb2[R] > 255 || rgb2[G] > 255 || rgb2[B] > 255 || rgb2[R] < 0 || rgb2[G] < 0 || rgb2[B] < 0) {
        cons.log('#### PercI -> NormJ -> RGB  out of bounds')
        cons.log(rgb2)
      }
      if (colors[r][p][i].rgb.chan[0] !== rgb2[0] || colors[r][p][i].rgb.chan[1] !== rgb2[1] || colors[r][p][i].rgb.chan[2] !== rgb2[2]) {
        cons.log('#### PercI -> RGB  different from  PercI -> NormJ -> RGB.')
        cons.log('PercI -> RGB :', colors[r][p][i].rgb.chan)
        cons.log('PercI -> NormJ -> RGB :', rgb2)
        cons.log('PercI:', colors[r][p][i].eye)
        cons.log('PercI -> NormJ :', colors[r][p][i].eye.export(CS.EyeNormJ))
      }
      rgb2 = colors[r][p][i].eye.export(CS.EyeNormJ).export(CS.EyePercI).export(CS.RGBgamma).chan
      if (colors[r][p][i].rgb.chan[0] !== rgb2[0] || colors[r][p][i].rgb.chan[1] !== rgb2[1] || colors[r][p][i].rgb.chan[2] !== rgb2[2]) {
        cons.log('#### PercI -> RGB  different from  PercI -> NormJ -> PercI -> RGB.')
        cons.log('PercI -> RGB :', colors[r][p][i].rgb.chan)
        cons.log('PercI -> NormJ -> PercI -> RGB :', rgb2)
        cons.log('PercI:', colors[r][p][i].eye)
        cons.log('PercI -> NormJ -> PercI :', colors[r][p][i].eye.export(CS.EyeNormJ).export(CS.EyePercI))
      }
    }
  }
}

const rainbowSameI: Array<RGB> = []
const rainbowSameJ: Array<RGB> = []
for (let r = 0; r <= 180; r++) {
  colorI.r = r / 180
  colorI.p = 1
  colorI.i = Eye.lowestImax
  rainbowSameI.push(colorI.export(CS.RGBgamma))
  colorI.i = colorI.iMax
  rainbowSameJ.push(colorI.export(CS.RGBgamma))
}

const granularRainbow48: Array<RGB> = []
colorJ.p = 1
colorJ.j = 1
for (let r = 0; r <= 48; r++) {
  colorJ.r = r / 48
  granularRainbow48.push(colorJ.export(CS.RGBgamma))
}

const purityGradientJ : Array<Array<number[]>> = []
colorJ.j = 0.7
for (let r = 0; r < 1; r += 1 / 3) {
  const rowJ: Array<number[]> = []
  for (let p = 0; p <= 81; p++) {
    colorJ.r = r
    colorJ.p = p / 81
    rowJ.push(colorJ.export(CS.RGBgamma).chan)
  }
  purityGradientJ.push(rowJ)
}

const pures: Array<number[]> = []
const puresDimmed: Array<number[]> = []
const greyish: Array<number[]> = []
const greyishDimmed: Array<number[]> = []
for (let r = 0; r <= 6; r++) {
  colorI.r = r / 6
  colorI.i = Eye.lowestImax
  colorI.p = 1
  pures.push(colorI.export(CS.RGBgamma).chan)
  colorI.p = 0.4
  greyish.push(colorI.export(CS.RGBgamma).chan)
  colorI.i = 0.8 * Eye.lowestImax
  greyishDimmed.push(colorI.export(CS.RGBgamma).chan)
  colorI.p = 1
  puresDimmed.push(colorI.export(CS.RGBgamma).chan)
}
const primaryPermutInPures = [[2, 0, 4], [0, 2, 4], [0, 4, 2]]
const secondaryPermutInPures = [[0, 1, 2], [2, 3, 4], [4, 5, 6]]
const primariesInPures = [0, 2, 4]

const extremeGeysI: Array<RGB> = []
colorI.p = 0
for (let k = 0; k <= 16; k++) {
  colorI.i = colorI.iMax * k / 16
  extremeGeysI.push(colorI.export(CS.RGBgamma))
}
const extremeGeysJ: Array<RGB> = []
colorJ.p = 0
for (let k = 0; k <= 16; k++) {
  colorJ.j = k / 16
  extremeGeysJ.push(colorJ.export(CS.RGBgamma))
}

let krkpki1 = [0, 0, 0]
let krkpki2 = [0, 0, 0]
const distanceBetweenLastTwoColors = ref<number>(0)
function showDistanceTo (kr: number, kp: number, ki: number) {
  krkpki1 = krkpki2
  krkpki2 = [kr, kp, ki]
  distanceBetweenLastTwoColors.value = distance(colors[krkpki1[0]][krkpki1[1]][krkpki1[2]].eye, colors[kr][kp][ki].eye)
  distanceBetweenLastTwoColors.value = Math.round(1000 * distanceBetweenLastTwoColors.value) / 1000
}
</script>

<template>
  <div style="background-color: rgb(128,128,128); padding: 20px">
    <TabView>
      <TabPanel header="Demo">
        {{ enhanceColors([{color: ' #1200FF', identifier: 'hey'}, {color: '#000A00', identifier: 'hi'}], ColorBlindness.Red) }}
      </TabPanel>
      <TabPanel header="Screen calibration">
        <div style="background-color: black; height: 1000px; display: flex; flex-direction: column; padding: 5px">
          <h1>Screen quality check: wavelenghts of the primaries</h1>
          This test tells you whether the primaries of your screen are far enough from each other or if a primary activates too much two cones on your retina. This cannot be improved with the settings of your screen.<br>
          Watch this square with a spectroscope (or measure it with a spectrometer). <br>
          The center of the blue band must be below 467 nm¹² and the bright part of the band should remain below 470 nm³.<br>
          The center of the green band must be between 532 nm² and 549 nm¹ and the bright part of the band should stay above 510 nm³ and below 560 nm³.<br>
          The center of the red band must above 612 nm¹ (ideally at least 630 nm²) and the bright part of the band should remain above 600 nm³.<br>
          <span style="width: 300px; height: 300px; background-color: #E0E0E0; margin: auto;" />
          1. according to the recommendation Rec.709 of the ITU-R<br>
          2. according to the recommendation BT.2020 of the ITU-R<br>
          3. according to the best commercial screen found on https://clarkvision.com/articles/color-spaces
        </div>

        <h1>Screen calibration: color balance</h1>
        The secondaries must stand between the black lines.
        <br><br>
        <div v-for="(m,n) of 3" :key="m">
          <span v-for="(c,i) of 17" :key="i" style="display: inline-block; width: 24px; height: 40px;" :style="'background-color: rgb(' + granularRainbow48[16*n+i].chan[R] + ',' + granularRainbow48[16*n+i].chan[G] + ',' + granularRainbow48[16*n+i].chan[B] + ')'">
            <span v-if="(i/8-1)%2 == 0" style="display: inline-block; height: 100%; width: 100%; border: 1px solid black; border-bottom: 0; border-top: 0;" />
          </span>
          <br><br>
        </div>

        <h1>Screen calibration: gamma in medium brightness.</h1>
        Your screen gamma is 2.2 on the three channels if the following squares look plain (the center parts must not look brighter or dimmer).<br>
        For the test to work properly: the zoom of your browser must be 100% and you should look from far enough (or without glasses)
        <br><br>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(135,135,135)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + (k%2)*186 + ',' + (k%2)*186 + ',' + (k%2)*186 + ')'" /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(186,186,186)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + (k%2)*255 + ',' + (k%2)*255 + ',' + (k%2)*255 + ')'" /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(223,223,223)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + (255-(k%2)*69) + ',' + (255-(k%2)*69) + ',' + (255-(k%2)*69) + ')'" /><br>
        </div>
        <br><br>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(135,0,0)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + (k%2)*186 + ',' + 0 + ',' + 0 + ')'" /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(186,0,0)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + (k%2)*255 + ',' + 0 + ',' + 0 + ')'" /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(223,0,0)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + (255-(k%2)*69) + ',' + 0 + ',' + 0 + ')'" /><br>
        </div>
        <br><br>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,135,0)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + 0 + ',' + (k%2)*186 + ',' + 0 + ')'" /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,186,0)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + 0 + ',' + (k%2)*255 + ',' + 0 + ')'" /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,223,0)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + 0 + ',' + (255-(k%2)*69) + ',' + 0 + ')'" /><br>
        </div>
        <br><br>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,0,135)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (k%2)*186 + ')'" /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,0,186)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (k%2)*255 + ')'" /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,0,223)">
          <div v-for="(i,k) of 80" :key="k" style="display: inline-block; width: 1px; height: 60px;" :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (255-(k%2)*69) + ')'" /><br>
        </div>
        <br><br>

        <h1>Screen calibration: perceived linearity of extremes greys</h1>
        The perceived linearity of your screen in extreme greys (near black and white) is good if each middle square feels as different from its left square as from its right square.
        <br><br>
        <div v-for="(i,k) of 15" :key="k" style="text-align: center; background-color: #7030f0">
          <span v-if="k == 0 || k==1 || k==2 || k==12 || k==13 || k==14">
            <br>
            <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + extremeGeysI[0+k].chan[R] + ',' + extremeGeysI[0+k].chan[G] + ',' + extremeGeysI[0+k].chan[B] + ')'">
              I
            </div>
            <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + extremeGeysI[1+k].chan[R] + ',' + extremeGeysI[1+k].chan[G] + ',' + extremeGeysI[1+k].chan[B] + ')'">
              I
            </div>
            <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + extremeGeysI[2+k].chan[R] + ',' + extremeGeysI[2+k].chan[G] + ',' + extremeGeysI[2+k].chan[B] + ')'">
              I
            </div>
            <br>
            <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + extremeGeysJ[0+k].chan[R] + ',' + extremeGeysJ[0+k].chan[G] + ',' + extremeGeysJ[0+k].chan[B] + ')'">
              J
            </div>
            <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + extremeGeysJ[1+k].chan[R] + ',' + extremeGeysJ[1+k].chan[G] + ',' + extremeGeysJ[1+k].chan[B] + ')'">
              J
            </div>
            <div style="display: inline-block; width: 60px; height: 60px;" :style="'background-color: rgb(' + extremeGeysJ[2+k].chan[R] + ',' + extremeGeysJ[2+k].chan[G] + ',' + extremeGeysJ[2+k].chan[B] + ')'">
              J
            </div>
            <br>
          </span>
        </div>
      </TabPanel>
      <TabPanel header="Adjustements of the perception model">
        <h1>Adjustement of sensicol and iotaM</h1>
        <h2>sensicol</h2>
        In the first column, the framed primary must feel dimmer than its neighbors. <br>
        In the middle column, the framed primary must feel as bright as its neighbors. <br>
        In the last column, the framed primary must feel brighter than its neighbors. <br><br>
        <div v-for="(perm,k) of primaryPermutInPures" :key="k">
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[0]][R] + ',' + pures[perm[0]][G] + ',' + pures[perm[0]][B] + ')'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + puresDimmed[perm[1]][R] + ',' + puresDimmed[perm[1]][G] + ',' + puresDimmed[perm[1]][B] + '); border: 1px solid black'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[2]][R] + ',' + pures[perm[2]][G] + ',' + pures[perm[2]][B] + ')'" />
          <span style="margin-left: 60px;">&nbsp;</span>
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[0]][R] + ',' + pures[perm[0]][G] + ',' + pures[perm[0]][B] + ')'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[1]][R] + ',' + pures[perm[1]][G] + ',' + pures[perm[1]][B] + '); border: 1px solid black'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[2]][R] + ',' + pures[perm[2]][G] + ',' + pures[perm[2]][B] + ')'" />
          <span style="margin-left: 60px;">&nbsp;</span>
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + puresDimmed[perm[0]][R] + ',' + puresDimmed[perm[0]][G] + ',' + puresDimmed[perm[0]][B] + ')'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[1]][R] + ',' + pures[perm[1]][G] + ',' + pures[perm[1]][B] + '); border: 1px solid black'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + puresDimmed[perm[2]][R] + ',' + puresDimmed[perm[2]][G] + ',' + puresDimmed[perm[2]][B] + ')'" />
          <span style="margin-left: 60px;">&nbsp;</span>sensicol[{{ k }}]
          <br>
        </div>

        <h2>iotaM</h2>
        In the first column, the framed secondary must feel dimmer than its neighbors. <br>
        In the middle column, the framed secondary must feel as bright as its neighbors. <br>
        In the last column, the framed secondary must feel brighter than its neighbors. <br><br>
        <div v-for="(perm,k) of secondaryPermutInPures" :key="k">
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[0]][R] + ',' + pures[perm[0]][G] + ',' + pures[perm[0]][B] + ')'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + puresDimmed[perm[1]][R] + ',' + puresDimmed[perm[1]][G] + ',' + puresDimmed[perm[1]][B] + '); border: 1px solid black'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[2]][R] + ',' + pures[perm[2]][G] + ',' + pures[perm[2]][B] + ')'" />
          <span style="margin-left: 60px;">&nbsp;</span>
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[0]][R] + ',' + pures[perm[0]][G] + ',' + pures[perm[0]][B] + ')'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[1]][R] + ',' + pures[perm[1]][G] + ',' + pures[perm[1]][B] + '); border: 1px solid black'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[2]][R] + ',' + pures[perm[2]][G] + ',' + pures[perm[2]][B] + ')'" />
          <span style="margin-left: 60px;">&nbsp;</span>
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + puresDimmed[perm[0]][R] + ',' + puresDimmed[perm[0]][G] + ',' + puresDimmed[perm[0]][B] + ')'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[perm[1]][R] + ',' + pures[perm[1]][G] + ',' + pures[perm[1]][B] + '); border: 1px solid black'" />
          <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + puresDimmed[perm[2]][R] + ',' + puresDimmed[perm[2]][G] + ',' + puresDimmed[perm[2]][B] + ')'" />
          <br>
        </div>

        <h2>control</h2>
        All colors in this rainbow must have the same perceived brightness: <br><br>
        <span v-for="(c,i) of rainbowSameI" :key="i" style="display: inline-block; width: 4px; height: 40px;" :style="'background-color: rgb(' + c.chan[R] + ',' + c.chan[G] + ',' + c.chan[B] + ')'" />
        <br><br>

        <h1>Adjustement of overwhite</h1>
        The three rows must transition from grey to pure at the same speed.<br>
        On the right, the middle square must feel as different from its left square as from its right square. <br><br>
        <div v-for="(row,k) of purityGradientJ" :key="k" style="border: 0px;">
          <span v-for="(col,m) of row" :key="m">
            <div style="display: inline-block; width: 2px; height: 40px;" :style="'background-color: rgb(' + col[R] + ',' + col[G] + ',' + col[B] + ')'" />
          </span>
          <span style="margin-left: 60px;">&nbsp;</span>
          <div style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + row[10][R] + ',' + row[10][G] + ',' + row[10][B] + ')'" />
          <div style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + row[row.length/2][R] + ',' + row[row.length/2][G] + ',' + row[row.length/2][B] + ')'" />
          <div style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + row[row.length-11][R] + ',' + row[row.length-11][G] + ',' + row[row.length-11][B] + ')'" />
          <span style="margin-left: 60px;">&nbsp;</span>overwhite[{{ k }}]
          <br>
        </div>

        <div style="background-color: rgb(160,160,160); padding: 10px">
          <h1>Adjustement of iotaD</h1>
          In the first column, the framed color must feel dimmer than its neighbors.<br>
          In the middle column, the framed color must feel as bright as its neighbors.<br>
          In the last column, the framed color must feel brighter than its neighbors.<br><br>
          <div v-for="(prim,k) of primariesInPures" :key="k">
            <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[prim][R] + ',' + pures[prim][G] + ',' + pures[prim][B] + ')'" />
            <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + greyishDimmed[prim][R] + ',' + greyishDimmed[prim][G] + ',' + greyishDimmed[prim][B] + '); border: 1px solid black'" />
            <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[prim][R] + ',' + pures[prim][G] + ',' + pures[prim][B] + ')'" />
            <span style="margin-left: 60px;">&nbsp;</span>
            <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[prim][R] + ',' + pures[prim][G] + ',' + pures[prim][B] + ')'" />
            <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + greyish[prim][R] + ',' + greyish[prim][G] + ',' + greyish[prim][B] + '); border: 1px solid black'" />
            <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + pures[prim][R] + ',' + pures[prim][G] + ',' + pures[prim][B] + ')'" />
            <span style="margin-left: 60px;">&nbsp;</span>
            <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + puresDimmed[prim][R] + ',' + puresDimmed[prim][G] + ',' + puresDimmed[prim][B] + ')'" />
            <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + greyish[prim][R] + ',' + greyish[prim][G] + ',' + greyish[prim][B] + '); border: 1px solid black'" />
            <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + puresDimmed[prim][R] + ',' + puresDimmed[prim][G] + ',' + puresDimmed[prim][B] + ')'" />
            <br>
          </div>
          <br>
          All of the previous parameters have an influence here and are supposed to have been properly adjusted. Therefore, this step acts indirectly as a control of the previous steps. <br>
          If a row here shows a discrepancy, that might indicate that a previous parameter was imperfectly set. Try to readjust the sensicol value or the overwhite value corresponding to the test that fails here.
          <br>
        </div>

        <h1>Adjustement of rPowers</h1>
        Adjust rPowers to give a feeling of similarity and linearity to these 6 gradients. <br>
        There are two ways to help yourself with this task: <br>
        - on the left side, try to make each row progress at the same speed, <br>
        - on the right side, try to make the middle square as different from its left square as from its right square.
        <br><br>
        <span v-for="(c,i) of 31" :key="i" style="display: inline-block; width: 6px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[180-i].chan[R] + ',' + rainbowSameJ[180-i].chan[G] + ',' + rainbowSameJ[180-i].chan[B] + ')'" />
        <span style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[180].chan[R] + ',' + rainbowSameJ[180].chan[G] + ',' + rainbowSameJ[180].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[165].chan[R] + ',' + rainbowSameJ[165].chan[G] + ',' + rainbowSameJ[165].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[150].chan[R] + ',' + rainbowSameJ[150].chan[G] + ',' + rainbowSameJ[150].chan[B] + ')'" />
        <span style="margin-left: 60px;">&nbsp;</span>rPowers[0]
        <br>
        <span v-for="(c,i) of 31" :key="i" style="display: inline-block; width: 6px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[i].chan[R] + ',' + rainbowSameJ[i].chan[G] + ',' + rainbowSameJ[i].chan[B] + ')'" />
        <span style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[0].chan[R] + ',' + rainbowSameJ[0].chan[G] + ',' + rainbowSameJ[0].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[15].chan[R] + ',' + rainbowSameJ[15].chan[G] + ',' + rainbowSameJ[15].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[30].chan[R] + ',' + rainbowSameJ[30].chan[G] + ',' + rainbowSameJ[30].chan[B] + ')'" />
        <br>
        <span v-for="(c,i) of 31" :key="i" style="display: inline-block; width: 6px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[60-i].chan[R] + ',' + rainbowSameJ[60-i].chan[G] + ',' + rainbowSameJ[60-i].chan[B] + ')'" />
        <span style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[60].chan[R] + ',' + rainbowSameJ[60].chan[G] + ',' + rainbowSameJ[60].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[45].chan[R] + ',' + rainbowSameJ[45].chan[G] + ',' + rainbowSameJ[45].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[30].chan[R] + ',' + rainbowSameJ[30].chan[G] + ',' + rainbowSameJ[30].chan[B] + ')'" />
        <span style="margin-left: 60px;">&nbsp;</span>rPowers[1]
        <br>
        <span v-for="(c,i) of 31" :key="i" style="display: inline-block; width: 6px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[60+i].chan[R] + ',' + rainbowSameJ[60+i].chan[G] + ',' + rainbowSameJ[60+i].chan[B] + ')'" />
        <span style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[60].chan[R] + ',' + rainbowSameJ[60].chan[G] + ',' + rainbowSameJ[60].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[75].chan[R] + ',' + rainbowSameJ[75].chan[G] + ',' + rainbowSameJ[75].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[90].chan[R] + ',' + rainbowSameJ[90].chan[G] + ',' + rainbowSameJ[90].chan[B] + ')'" />
        <br>
        <span v-for="(c,i) of 31" :key="i" style="display: inline-block; width: 6px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[120-i].chan[R] + ',' + rainbowSameJ[120-i].chan[G] + ',' + rainbowSameJ[120-i].chan[B] + ')'" />
        <span style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[120].chan[R] + ',' + rainbowSameJ[120].chan[G] + ',' + rainbowSameJ[120].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[105].chan[R] + ',' + rainbowSameJ[105].chan[G] + ',' + rainbowSameJ[105].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[90].chan[R] + ',' + rainbowSameJ[90].chan[G] + ',' + rainbowSameJ[90].chan[B] + ')'" />
        <span style="margin-left: 60px;">&nbsp;</span>rPowers[2]
        <br>
        <span v-for="(c,i) of 31" :key="i" style="display: inline-block; width: 6px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[120+i].chan[R] + ',' + rainbowSameJ[120+i].chan[G] + ',' + rainbowSameJ[120+i].chan[B] + ')'" />
        <span style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[120].chan[R] + ',' + rainbowSameJ[120].chan[G] + ',' + rainbowSameJ[120].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[135].chan[R] + ',' + rainbowSameJ[135].chan[G] + ',' + rainbowSameJ[135].chan[B] + ')'" />
        <span style="display: inline-block; width: 40px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[150].chan[R] + ',' + rainbowSameJ[150].chan[G] + ',' + rainbowSameJ[150].chan[B] + ')'" />
        <br><br>
        Control: in the rainbow, the widths of the primary and secondary smudges must all look equal.<br><br>
        <span v-for="(c,i) of rainbowSameJ" :key="i" style="display: inline-block; width: 3px; height: 40px;" :style="'background-color: rgb(' + c.chan[R] + ',' + c.chan[G] + ',' + c.chan[B] + ')'" />
        <span v-for="(c,i) of 31" :key="i" style="display: inline-block; width: 3px; height: 40px;" :style="'background-color: rgb(' + rainbowSameJ[i].chan[R] + ',' + rainbowSameJ[i].chan[G] + ',' + rainbowSameJ[i].chan[B] + ')'" />
        <br><br>

        <br><br>
        <div class="all-colors">
          <h1>For each primary and each secondary, all purities and perceived brightness :</h1>
          Here we flatten the color space. r, p and i progress linearly. <br>
          Click two colors to see their distance. <br>
          <br>
          <div v-for="(rRow,r) of colors" :key="r">
            <div v-for="(pRow,p) of rRow" :key="p">
              <div v-for="(c,i) of pRow" :key="i" style="display: inline-block; width: 20px; height: 20px;" :style="'background-color: rgb(' + c.rgb.chan[R] + ',' + c.rgb.chan[G] + ',' + c.rgb.chan[B] + ')'" @click="showDistanceTo(r,p,i)" />
            </div>
            <br>
          </div>
          <div v-if="distanceBetweenLastTwoColors" class="meter">
            Distance between the last two colors <br>
            that you clicked: {{ distanceBetweenLastTwoColors }} <br><br>
            {{ colors[krkpki1[0]][krkpki1[1]][krkpki1[2]].eye }} <br>
            {{ colors[krkpki2[0]][krkpki2[1]][krkpki2[2]].eye }}
          </div>
        </div>
      </TabPanel>
    </TabView>
  </div>
</template>

<style lang="scss" scoped>
.all-colors {
  position: relative;
  overflow: visible;
  .meter {
    background-color: rgb(150, 109, 247);
    padding: 10px;
    position: fixed;
    bottom: 0px;
    right: 0px;
    display: inline-block;
  }
}
</style>
