import { CS, R, G, B, RGB, Eye, distance, projectOntoCBPlane } from './colorspaces'

const stallingPeriod = 8 // controls how long the progress can stall before a run of Tabu search is considered finished
const maxModeDuration = 4 // controls how long the max-Tabu search lasts
const debug = true

let originalColors: RGB[] // CS.RGBlinear
let distancesOrig: number[][]
let depth: number
const cons = console

export function setExplorer(colors: RGB[], distances: number[][], searchDepth: number): void {
  originalColors = colors
  distancesOrig = distances
  depth = searchDepth
}

export enum SearchMode {
  Sum,
  Max,
}

export class Explorer {
  iterated = 0
  bestIteration = 0
  bestError = 2.0 ** 24
  bestWip3D: RGB[] = []
  protected maxIterations = 0
  protected willStallAt = 0
  protected computationMode: SearchMode
  protected watchStalling = false
  protected wip3D: RGB[] = [] // CS.RGBlinear
  protected wip2D: Eye[] = [] // CS.EyePercI
  protected errors2D: number[] = []
  /** must be passed to `projectOntoCBPlane()` at every call made from this class */
  protected rgbProjectionBuffer = new RGB(CS.RGBlinear)
  /** must be passed to `projectOntoCBPlane()` at every call made from this class */
  protected cbProjectionBuffer = new Eye(CS.EyePercI)

  /** @param watchStalling if false, the loop stops after its predefined number of iterations, each better solution does
   *  not restart the count.
   */
  constructor(computationMode: SearchMode, watchStalling: boolean) {
    this.computationMode = computationMode
    this.watchStalling = watchStalling
  }

  prepareSearch(source: RGB[], startingDepth: number): void {
    this.maxIterations = (this.computationMode === SearchMode.Sum)
      ? depth - startingDepth
      : maxModeDuration * originalColors.length
    this.iterated = 0
    this.bestIteration = 0
    // copying the original colors into wip3D: they will be modified progressively by the search phase in such a way
    // that they get more and more distinguishable when viewed by a color blind person
    this.wip3D = source.map(col => col.export(CS.RGBlinear))
    this.bestWip3D = this.wip3D.map(col => col.export(CS.RGBlinear))
    this.bestError = this.totalError(false) // this call fills `wip2D` and `errors2D`
    this.shiftStallingLimit()
  }

  explore: undefined | (() => void)

  /** the implementation of `explore()` calls it when it finds a better solution, to postpone the stalling limit */
  protected shiftStallingLimit(): void {
    if (this.watchStalling) {
      this.willStallAt = this.iterated
      + (this.computationMode === SearchMode.Sum ? stallingPeriod : maxModeDuration) * originalColors.length
    }
    else {
      this.willStallAt = this.maxIterations
    }
  }

  /** for a given color k stored in `wipColor2D`, this function calculates a value that reflects
   *    the sum over every color l of
   *      the difference between:
   *        the distance between the original colors k and l
   *        the distance between the projected colors `wipColor2D` and l */
  protected distError(k: number, wipColor2D: Eye): number {
    let result = 0.0
    for (let l = 0; l < distancesOrig.length; l++) {
      if (k === l) {
        continue
      }
      // -1 means that dist(k,l) for the CB person is much shorter (not good), 0 means equal (perfect), > 0 means longer
      const diff = (distancesOrig[k][l] <= 0.001)
        ? distance(wipColor2D, this.wip2D[l])
        : distance(wipColor2D, this.wip2D[l]) / distancesOrig[k][l] - 1
      let error: number
      if (diff < 0) {
        error = (diff <= -1) ? 2.0 ** 16 : 1 / (1 + diff) - 1
      }
      else {
        error = 0 // larger distances do not penalize, but we do not want to allow them to compensate for shorter ones
      }
      result += error
    }
    return result
  }

  protected totalError(errors2DisUpToDate: boolean): number {
    if (!errors2DisUpToDate) {
      this.wip2D = this.wip3D.map(
        col => projectOntoCBPlane(col.chan, this.rgbProjectionBuffer, this.cbProjectionBuffer).export(CS.EyePercI))
      this.errors2D = this.wip2D.map((col, k) => this.distError(k, col))
    }
    const initial = (this.computationMode === SearchMode.Sum) ? 0.0 : -(2.0 ** 16)
    return this.errors2D.reduce(
      (prev, curr) => (this.computationMode === SearchMode.Sum) ? prev + curr : Math.max(prev, curr),
      initial)
  }
}

export class TabuSearcher extends Explorer {
  protected tabuMoves: Array<number[]> = []

  override explore = () => {
    this.tabuMoves.length = 0
    for (let k = 0; k < this.wip3D.length; k++) {
      this.tabuMoves.push([0, 0, 0])
    }
    const errorBefore = this.bestError
    while (this.iterated < this.maxIterations && this.iterated < this.willStallAt) {
      const error = this.totalError(true)
      if (error < this.bestError) {
        this.bestError = error
        this.bestIteration = this.iterated
        this.bestWip3D = this.wip3D.map(col => col.export(CS.RGBlinear))
        this.shiftStallingLimit()
      }
      this.optimizeOneStepFurther()
      this.iterated++
    }
    if (debug) {
      cons.log('Tabu search, mode', this.computationMode, '. Iterated', this.iterated,
        'times. Distance errors: before', errorBefore, 'after', this.bestError)
    }
  }

  protected triedRGB = new RGB(CS.RGBlinear)

  protected optimizeOneStepFurther() {
    let bestK: number = 0
    let lowestError: number = 0
    let bestErrorGain: number = 2.0 ** 24
    let moveToForbid: number[] = []
    let bestColor: number[] = []

    for (let k = 0; k < this.wip3D.length; k++) {
      this.triedRGB.import(this.wip3D[k])
      const step = 2 / 256
      for (const c of [R, G, B]) {
        for (const s of [-1, +1]) {
          if (this.tabuMoves[k][c] === s) {
            continue
          }
          const restoredValue = this.triedRGB.chan[c]
          this.triedRGB.chan[c] += s * step
          if (this.triedRGB.chan[c] < 0) {
            this.triedRGB.chan[c] = 0
          }
          if (this.triedRGB.chan[c] > 1) {
            this.triedRGB.chan[c] = 1
          }
          const error = this.distError(k,
            projectOntoCBPlane(this.triedRGB.chan, this.rgbProjectionBuffer, this.cbProjectionBuffer))
          if (error - this.errors2D[k] < bestErrorGain) {
            bestErrorGain = error - this.errors2D[k]
            lowestError = error
            bestK = k
            moveToForbid = [0, 0, 0]
            moveToForbid[c] = -s
            bestColor = [...this.triedRGB.chan]
          }
          this.triedRGB.chan[c] = restoredValue
        }
      }
    }
    this.wip3D[bestK].import(bestColor)
    this.wip2D[bestK].import(projectOntoCBPlane(bestColor, this.rgbProjectionBuffer, this.cbProjectionBuffer))
    this.errors2D[bestK] = lowestError
    // forbidding only the last move is enough to set to 4 the minimum number of steps required to go back here
    this.tabuMoves[bestK] = moveToForbid
  }
}

export class Shaker extends Explorer {
  override explore = () => {
    if (debug) {
      cons.log('Shaking.')
    }
    this.wip3D.forEach((col) => {
      const eye = col.export(CS.EyeNormJ)
      eye.r += 1 / 6 * (1 - 2 * Math.random())
      eye.j = 0.1 + 0.8 * eye.j
      eye.p = 0.1 + 0.8 * eye.p
      eye.limit()
      col.import(eye)
    })
    this.bestWip3D = this.wip3D.map(col => col.export(CS.RGBlinear))
    this.bestError = this.totalError(false)
  }
}
