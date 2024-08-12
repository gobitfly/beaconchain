import type { RGB, ColorBlindness } from './colorspaces'
import { CS, distance, setColorBlindness } from './colorspaces'
import { setExplorer, SearchMode, TabuSearcher, Shaker } from './explorer'

const debug = true
const cons = console

let originalColors: RGB[]
let distancesOrig: number[][]

/** @param input must be in the CS.RGBgamma format.
* @returns enchanced colors in the CS.RGBgamma format */
export function optimize(input: RGB[], blindness: ColorBlindness): RGB[] {
  setColorBlindness(blindness)
  originalColors = input.map(col => col.export(CS.RGBlinear))
  calulateDistances()
  const searchDepth = 4000 * originalColors.length
  setExplorer(originalColors, distancesOrig, searchDepth)

  // now, search phase
  let currentDepth = 0
  let disturbanceMode = 0
  const tabuSum = new TabuSearcher(SearchMode.Sum, true)
  const tabuMax = new TabuSearcher(SearchMode.Max, false)
  const shaker = new Shaker(SearchMode.Max, false)
  let bestError = 2.0 ** 24
  let bestFinding: RGB[] = []
  let bestIteration = 0
  tabuSum.prepareSearch(originalColors, currentDepth)
  cons.log('Measurement of distance errors before search:', tabuSum.bestError)

  while (currentDepth < searchDepth) {
    tabuSum.explore()
    if (tabuSum.bestError < bestError) {
      bestError = tabuSum.bestError
      bestFinding = tabuSum.bestWip3D
      bestIteration = currentDepth + tabuSum.bestIteration
      if (debug) {
        cons.log('Globally-better set found at iteration', bestIteration)
      }
    }
    currentDepth += tabuSum.iterated
    if (currentDepth >= searchDepth) {
      break
    }
    let toBeImproved: RGB[]
    disturbanceMode = (++disturbanceMode) % 2
    if (disturbanceMode) {
      tabuMax.prepareSearch(tabuSum.bestWip3D, currentDepth)
      tabuMax.explore()
      toBeImproved = tabuMax.bestWip3D
      currentDepth += tabuMax.iterated
    }
    else {
      shaker.prepareSearch(bestFinding, currentDepth)
      shaker.explore()
      toBeImproved = shaker.bestWip3D
      currentDepth += shaker.iterated
    }
    tabuSum.prepareSearch(toBeImproved, currentDepth)
  }

  cons.log('Iterated', currentDepth, 'times. Best color set found at iteration', bestIteration)
  cons.log('Measurement of distance errors after search:', bestError)
  return bestFinding.map(col => col.export(CS.RGBgamma))
}

function calulateDistances(): void {
  distancesOrig = []
  const originalEye = originalColors.map(col => col.export(CS.EyePercI))
  for (const colA of originalEye) {
    const line: number[] = []
    for (const colB of originalEye) {
      line.push(distance(colA, colB))
    }
    distancesOrig.push(line)
  }
  if (debug) {
    cons.log('Original distances:', distancesOrig)
  }
}
