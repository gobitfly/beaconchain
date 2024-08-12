<script lang="ts" setup>
import { warn } from 'vue'
import { CS, R, G, B, RGB, Eye, distance, ColorBlindness, projectOntoCBPlane } from './colorspaces'
import { optimize } from './optimizer'

type ColorDefinition = { color: string, identifier: string }

function enhanceColors(colors: ColorDefinition[], colorBlindness: ColorBlindness): ColorDefinition[] {
  const original = []
  for (const def of colors) {
    if (!def.color.includes('#')) {
      warn('Color {', def.color, def.identifier, '} is given in an unknown format.')
      return colors
    }
    const rgb = new RGB(CS.RGBgamma)
    rgb.import(CSStoRGBarray(def.color))
    original.push(rgb)
  }
  const startTime = performance.now()
  const enhanced = optimize(original, colorBlindness) // 🪄✨
  cons.log('Time spent: ', Math.round((performance.now() - startTime) / 1000), 's.')

  const result: ColorDefinition[] = []
  for (let c = 0; c < colors.length; c++) {
    result.push({
      color: RGBarrayToCSS(enhanced[c].chan),
      identifier: colors[c].identifier,
    })
  }
  return result
}

function CSStoRGBarray(CSS: string): number[] {
  const i = CSS.indexOf('#')
  const col = parseInt(CSS.slice(i + 1), 16)
  const arr: number[] = []
  arr[B] = col & 0xFF
  arr[G] = (col >> 8) & 0xFF
  arr[R] = (col >> 16) & 0xFF
  return arr
}

function RGBarrayToCSS(RGB: number[]): string {
  let hex = (RGB[B] | RGB[G] << 8 | RGB[R] << 16).toString(16)
  hex = ('000000' + hex).slice(-6)
  return '#' + hex
}

//
// TESTS AND ADJUSTEMENTS
//

const cons = console

let randColorsEnhanced: string[] = []
let randColorsCB: string[] = []
let randColorsEnhancedCB: string[] = []
const enhancements = ref(0)
const randColors: string[] = []
const RGBgamma = new RGB(CS.RGBgamma)

onMounted(() => {
  generateNewColors()
  enhanceColorSet()
})

function generateNewColors(): void {
  randColors.length = 0
  for (let i = 0; i < 16; i++) {
    // from Alexander:
    const letters = '0123456789ABCDEF'
    let color = '#'
    for (let i = 0; i < 6; i++) {
      color += letters[Math.floor(Math.random() * 16)]
    }
    randColors.push(color)
  }
}

function enhanceColorSet(): void {
  randColorsEnhanced = enhanceColors(randColors.map(col => ({ color: col, identifier: '' })), ColorBlindness.Green)
    .map(obj => obj.color)
  randColorsCB = randColors.map(col => RGBarrayToCSS(projectOntoCBPlane(RGBgamma.import(
    CSStoRGBarray(col)).export(CS.RGBlinear).chan).export(CS.RGBgamma).chan))
  randColorsEnhancedCB = randColorsEnhanced.map(col => RGBarrayToCSS(projectOntoCBPlane(
    RGBgamma.import(CSStoRGBarray(col)).export(CS.RGBlinear).chan).export(CS.RGBgamma).chan))
  enhancements.value++
}

const colorI = new Eye(CS.EyePercI)
const colorJ = new Eye(CS.EyeNormJ)

const allColors: Array<Array<Array<{ rgb: RGB, eye: Eye }>>> = []
const numR = 6
const numP = 21
const numI = 40
let maxIntensityMinIndex = 1000
for (let r = 0; r <= numR; r++) {
  allColors.push([])
  for (let p = 0; p <= numP; p++) {
    allColors[r].push([])
    for (let i = 0; i <= numI; i++) {
      if (i > 0 && (i / numI) > colorI.iMax) {
        if (i - 1 < maxIntensityMinIndex) {
          maxIntensityMinIndex = i - 1
        }
        break
      }
      colorI.r = r / numR
      colorI.p = p / numP
      colorI.i = i / numI
      allColors[r][p].push({ rgb: colorI.export(CS.RGBgamma), eye: colorI.export(CS.EyePercI) })
      if (allColors[r][p][i].rgb.chan[R] > 255 || allColors[r][p][i].rgb.chan[G] > 255
        || allColors[r][p][i].rgb.chan[B] > 255 || allColors[r][p][i].rgb.chan[R] < 0
        || allColors[r][p][i].rgb.chan[G] < 0 || allColors[r][p][i].rgb.chan[B] < 0) {
        cons.log('#### PercI -> RGB  out of bounds')
        cons.log(allColors[r][p][i].rgb)
      }
      let rgb2 = ((new Eye(CS.EyePercI)).import(allColors[r][p][i].rgb)).export(CS.RGBgamma).chan
      if (rgb2[R] > 255 || rgb2[G] > 255 || rgb2[B] > 255 || rgb2[R] < 0 || rgb2[G] < 0 || rgb2[B] < 0) {
        cons.log('#### PercI -> RGB -> PercI -> RGB  out of bounds')
        cons.log(rgb2)
      }
      if (allColors[r][p][i].rgb.chan[0] !== rgb2[0] || allColors[r][p][i].rgb.chan[1] !== rgb2[1]
        || allColors[r][p][i].rgb.chan[2] !== rgb2[2]) {
        cons.log('#### PercI -> RGB  different from  PercI -> RGB -> PercI -> RGB.')
        cons.log('PercI:', allColors[r][p][i].eye)
        cons.log('PercI -> RGB -> PercI :', (new Eye(CS.EyePercI)).import(allColors[r][p][i].rgb))
        cons.log('PercI -> RGB :', allColors[r][p][i].rgb.chan)
        cons.log('PercI -> RGB -> PercI -> RGB :', rgb2)
      }
      rgb2 = allColors[r][p][i].eye.export(CS.RGBlinear).export(CS.EyeNormJ).export(CS.RGBgamma).chan
      if (rgb2[R] > 255 || rgb2[G] > 255 || rgb2[B] > 255 || rgb2[R] < 0 || rgb2[G] < 0 || rgb2[B] < 0) {
        cons.log('#### PercI -> RGB -> NormJ -> RGB  out of bounds')
        cons.log(rgb2)
      }
      if (allColors[r][p][i].rgb.chan[0] !== rgb2[0] || allColors[r][p][i].rgb.chan[1] !== rgb2[1]
        || allColors[r][p][i].rgb.chan[2] !== rgb2[2]) {
        cons.log('#### PercI -> RGB  different from  PercI -> RGB -> NormJ -> RGB.')
        cons.log('PercI -> RGB :', allColors[r][p][i].rgb.chan)
        cons.log('PercI -> RGB -> NormJ -> RGB :', rgb2)
        cons.log('PercI:', allColors[r][p][i].eye)
        cons.log('PercI -> RGB -> NormJ :', allColors[r][p][i].eye.export(CS.RGBlinear).export(CS.EyeNormJ))
      }
      rgb2 = allColors[r][p][i].eye.export(CS.EyeNormJ).export(CS.RGBgamma).chan
      if (rgb2[R] > 255 || rgb2[G] > 255 || rgb2[B] > 255 || rgb2[R] < 0 || rgb2[G] < 0 || rgb2[B] < 0) {
        cons.log('#### PercI -> NormJ -> RGB  out of bounds')
        cons.log(rgb2)
      }
      if (allColors[r][p][i].rgb.chan[0] !== rgb2[0] || allColors[r][p][i].rgb.chan[1] !== rgb2[1]
        || allColors[r][p][i].rgb.chan[2] !== rgb2[2]) {
        cons.log('#### PercI -> RGB  different from  PercI -> NormJ -> RGB.')
        cons.log('PercI -> RGB :', allColors[r][p][i].rgb.chan)
        cons.log('PercI -> NormJ -> RGB :', rgb2)
        cons.log('PercI:', allColors[r][p][i].eye)
        cons.log('PercI -> NormJ :', allColors[r][p][i].eye.export(CS.EyeNormJ))
      }
      rgb2 = allColors[r][p][i].eye.export(CS.EyeNormJ).export(CS.EyePercI).export(CS.RGBgamma).chan
      if (allColors[r][p][i].rgb.chan[0] !== rgb2[0] || allColors[r][p][i].rgb.chan[1] !== rgb2[1]
        || allColors[r][p][i].rgb.chan[2] !== rgb2[2]) {
        cons.log('#### PercI -> RGB  different from  PercI -> NormJ -> PercI -> RGB.')
        cons.log('PercI -> RGB :', allColors[r][p][i].rgb.chan)
        cons.log('PercI -> NormJ -> PercI -> RGB :', rgb2)
        cons.log('PercI:', allColors[r][p][i].eye)
        cons.log('PercI -> NormJ -> PercI :', allColors[r][p][i].eye.export(CS.EyeNormJ).export(CS.EyePercI))
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

const purityGradientJ: Array<Array<number[]>> = []
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
function showDistanceTo(kr: number, kp: number, ki: number) {
  krkpki1 = krkpki2
  krkpki2 = [kr, kp, ki]
  distanceBetweenLastTwoColors.value = distance(
    allColors[krkpki1[0]][krkpki1[1]][krkpki1[2]].eye, allColors[kr][kp][ki].eye)
  distanceBetweenLastTwoColors.value = Math.round(1000 * distanceBetweenLastTwoColors.value) / 1000
}
</script>

<template>
  <TabView>
    <TabPanel header="Demo">
      <div
        v-if="!enhancements"
        style="margin: 10px; font-weight:bold; color:red"
      >
        Calculations in progress, please wait approximately 10 s...
      </div>
      <div
        v-else
        :key="enhancements"
        style="display: flex"
      >
        <div style="background-color: rgb(128,128,128); padding: 20px;">
          <div style="display: flex; gap: 50px; margin:30px; margin-bottom: 10px;">
            <div>
              <div style="text-align: center; width:240px;">
                seen with normal vision <br><br>
              </div>
              <div
                v-for="(_, m) of 4"
                :key="m"
              >
                <div
                  v-for="(__, n) of 4"
                  :key="n"
                  style="display: inline-block; width: 60px; height: 60px;"
                  :style="'background-color:'+randColors[4*m+n]"
                />
              </div>
              <div style="text-align: center; width:240px; font-size: 50px; margin-top: 10px;">
                ⬇
              </div>
            </div>
            <div style="margin-top: 130px; font-size: 30px;">
              →
            </div>
            <div>
              <div style="text-align: center; width:240px;">
                seen by a color-blind person <br><br>
              </div>
              <div
                v-for="(_, m) of 4"
                :key="m"
              >
                <div
                  v-for="(__, n) of 4"
                  :key="n"
                  style="display: inline-block; width: 60px; height: 60px;"
                  :style="'background-color:'+randColorsCB[4*m+n]"
                />
              </div>
            </div>
          </div>
          <div style="display: flex; gap: 50px; margin:30px; margin-top: 10px;">
            <div>
              <div style="text-align: center; width:240px;">
                enhanced by Bitfly's technology <br><br>
              </div>
              <div
                v-for="(_, m) of 4"
                :key="m"
              >
                <div
                  v-for="(__, n) of 4"
                  :key="n"
                  style="display: inline-block; width: 60px; height: 60px;"
                  :style="'background-color:'+randColorsEnhanced[4*m+n]"
                />
              </div>
            </div>
            <div style="margin-top: 130px; font-size: 30px;">
              →
            </div>
            <div>
              <div style="text-align: center; width:240px;">
                seen by a color-blind person <br><br>
              </div>
              <div
                v-for="(_, m) of 4"
                :key="m"
              >
                <div
                  v-for="(__, n) of 4"
                  :key="n"
                  style="display: inline-block; width: 60px; height: 60px;"
                  :style="'background-color:'+randColorsEnhancedCB[4*m+n]"
                />
              </div>
            </div>
          </div>
        </div>
        <div style="margin-top: 190px; margin-left: 80px; flex-direction: column;">
          <BcButton @click="() => { generateNewColors(); enhanceColorSet() }">
            New colors
          </BcButton>
          <br>
          <BcButton
            style="margin-top: 350px"
            @click="enhanceColorSet()"
          >
            Re-enhance
          </BcButton>
        </div>
      </div>
    </TabPanel>
    <TabPanel header="Screen calibration">
      <div style="background-color: rgb(128,128,128); padding: 20px">
        <div style="background-color: black; height: 1000px; display: flex; flex-direction: column; padding: 5px">
          <h1>Screen quality check: wavelenghts of the primaries</h1>
          This test tells you whether the primaries of your screen are far enough from each other or if a primary
          activates too much two cones on your retina. This cannot be improved with the settings of your screen.<br>
          Watch this square with a spectroscope (or measure it with a spectrometer). <br>
          The center of the blue band must be below 467 nm¹² and the bright part of the band should remain below
          470 nm³.<br>
          The center of the green band must be between 532 nm² and 549 nm¹ and the bright part of the band should stay
          above 510 nm³ and below 560 nm³.<br>
          The center of the red band must above 612 nm¹ (ideally at least 630 nm²) and the bright part of the band
          should remain above 600 nm³.<br>
          <span style="width: 300px; height: 300px; background-color: #E0E0E0; margin: auto;" />
          1. according to the recommendation Rec.709 of the ITU-R<br>
          2. according to the recommendation BT.2020 of the ITU-R<br>
          3. according to the best commercial screen found on https://clarkvision.com/articles/color-spaces
        </div>

        <h1>Screen calibration: color balance</h1>
        The secondaries must stand between the black lines.
        <br><br>
        <div
          v-for="(m, n) of 3"
          :key="m"
        >
          <span
            v-for="(c, i) of 17"
            :key="i"
            style="display: inline-block; width: 24px; height: 40px;"
            :style="'background-color: rgb(' + granularRainbow48[16*n+i].chan[R] + ',' + granularRainbow48[16*n+i].chan[G] + ',' + granularRainbow48[16*n+i].chan[B] + ')'"
          >
            <span
              v-if="(i/8-1)%2 == 0"
              style="display: inline-block; height: 100%; width: 100%; border: 1px solid black; border-bottom: 0; border-top: 0;"
            />
          </span>
          <br><br>
        </div>

        <h1>Screen calibration: gamma in medium brightness.</h1>
        Your screen gamma is 2.2 on the three channels if the following squares look plain (the center parts must not
        look brighter or dimmer).<br>
        <i>For the test to work properly: the zoom of your browser must be 100% and you should look from far enough
          (or without glasses)</i>
        <br><br>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(135,135,135)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + (k%2)*186 + ',' + (k%2)*186 + ',' + (k%2)*186 + ')'"
          /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(186,186,186)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + (k%2)*255 + ',' + (k%2)*255 + ',' + (k%2)*255 + ')'"
          /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(223,223,223)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + (255-(k%2)*69) + ',' + (255-(k%2)*69) + ',' + (255-(k%2)*69) + ')'"
          /><br>
        </div>
        <br><br>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(135,0,0)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + (k%2)*186 + ',' + 0 + ',' + 0 + ')'"
          /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(186,0,0)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + (k%2)*255 + ',' + 0 + ',' + 0 + ')'"
          /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(223,0,0)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + (255-(k%2)*69) + ',' + 0 + ',' + 0 + ')'"
          /><br>
        </div>
        <br><br>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,135,0)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + 0 + ',' + (k%2)*186 + ',' + 0 + ')'"
          /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,186,0)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + 0 + ',' + (k%2)*255 + ',' + 0 + ')'"
          /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,223,0)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + 0 + ',' + (255-(k%2)*69) + ',' + 0 + ')'"
          /><br>
        </div>
        <br><br>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,0,135)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (k%2)*186 + ')'"
          /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,0,186)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (k%2)*255 + ')'"
          /><br>
        </div>
        <div style="display: inline-block; padding:50px; margin-left: 50px; background-color: rgb(0,0,223)">
          <div
            v-for="(i, k) of 80"
            :key="k"
            style="display: inline-block; width: 1px; height: 60px;"
            :style="'background-color: rgb(' + 0 + ',' + 0 + ',' + (255-(k%2)*69) + ')'"
          /><br>
        </div>
        <br><br>

        <h1>Screen calibration: perceived linearity of extremes greys</h1>
        The perceived linearity of your screen in extreme greys (near black and white) is good if each middle square
        feels as different from its left square as from its right square.
        <br><br>
        <div
          v-for="(i, k) of 15"
          :key="k"
          style="text-align: center; background-color: #7030f0"
        >
          <span v-if="k == 0 || k==1 || k==2 || k==12 || k==13 || k==14">
            <br>
            <div
              style="display: inline-block; width: 60px; height: 60px;"
              :style="'background-color: rgb(' + extremeGeysI[0+k].chan[R] + ',' + extremeGeysI[0+k].chan[G] + ',' + extremeGeysI[0+k].chan[B] + ')'"
            >
              I
            </div>
            <div
              style="display: inline-block; width: 60px; height: 60px;"
              :style="'background-color: rgb(' + extremeGeysI[1+k].chan[R] + ',' + extremeGeysI[1+k].chan[G] + ',' + extremeGeysI[1+k].chan[B] + ')'"
            >
              I
            </div>
            <div
              style="display: inline-block; width: 60px; height: 60px;"
              :style="'background-color: rgb(' + extremeGeysI[2+k].chan[R] + ',' + extremeGeysI[2+k].chan[G] + ',' + extremeGeysI[2+k].chan[B] + ')'"
            >
              I
            </div>
            <br>
            <div
              style="display: inline-block; width: 60px; height: 60px;"
              :style="'background-color: rgb(' + extremeGeysJ[0+k].chan[R] + ',' + extremeGeysJ[0+k].chan[G] + ',' + extremeGeysJ[0+k].chan[B] + ')'"
            >
              J
            </div>
            <div
              style="display: inline-block; width: 60px; height: 60px;"
              :style="'background-color: rgb(' + extremeGeysJ[1+k].chan[R] + ',' + extremeGeysJ[1+k].chan[G] + ',' + extremeGeysJ[1+k].chan[B] + ')'"
            >
              J
            </div>
            <div
              style="display: inline-block; width: 60px; height: 60px;"
              :style="'background-color: rgb(' + extremeGeysJ[2+k].chan[R] + ',' + extremeGeysJ[2+k].chan[G] + ',' + extremeGeysJ[2+k].chan[B] + ')'"
            >
              J
            </div>
            <br>
          </span>
        </div>
      </div>
    </TabPanel>
    <TabPanel header="Adjustements of the perception model">
      <div style="background-color: rgb(128,128,128); padding: 20px">
        <h1>Adjustement of sensicol and iotaM</h1>
        <h2>sensicol</h2>
        In the first column, the framed primary must feel dimmer than its neighbors. <br>
        In the middle column, the framed primary must feel as bright as its neighbors. <br>
        In the last column, the framed primary must feel brighter than its neighbors. <br><br>
        <div
          v-for="(perm, k) of primaryPermutInPures"
          :key="k"
        >
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[0]][R] + ',' + pures[perm[0]][G] + ',' + pures[perm[0]][B] + ')'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + puresDimmed[perm[1]][R] + ',' + puresDimmed[perm[1]][G] + ',' + puresDimmed[perm[1]][B] + '); border: 1px solid black'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[2]][R] + ',' + pures[perm[2]][G] + ',' + pures[perm[2]][B] + ')'"
          />
          <span style="margin-left: 60px;">&nbsp;</span>
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[0]][R] + ',' + pures[perm[0]][G] + ',' + pures[perm[0]][B] + ')'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[1]][R] + ',' + pures[perm[1]][G] + ',' + pures[perm[1]][B] + '); border: 1px solid black'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[2]][R] + ',' + pures[perm[2]][G] + ',' + pures[perm[2]][B] + ')'"
          />
          <span style="margin-left: 60px;">&nbsp;</span>
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + puresDimmed[perm[0]][R] + ',' + puresDimmed[perm[0]][G] + ',' + puresDimmed[perm[0]][B] + ')'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[1]][R] + ',' + pures[perm[1]][G] + ',' + pures[perm[1]][B] + '); border: 1px solid black'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + puresDimmed[perm[2]][R] + ',' + puresDimmed[perm[2]][G] + ',' + puresDimmed[perm[2]][B] + ')'"
          />
          <span style="margin-left: 60px;">&nbsp;</span>sensicol[{{ k }}]
          <br>
        </div>

        <h2>iotaM</h2>
        In the first column, the framed secondary must feel dimmer than its neighbors. <br>
        In the middle column, the framed secondary must feel as bright as its neighbors. <br>
        In the last column, the framed secondary must feel brighter than its neighbors. <br><br>
        <div
          v-for="(perm, k) of secondaryPermutInPures"
          :key="k"
        >
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[0]][R] + ',' + pures[perm[0]][G] + ',' + pures[perm[0]][B] + ')'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + puresDimmed[perm[1]][R] + ',' + puresDimmed[perm[1]][G] + ',' + puresDimmed[perm[1]][B] + '); border: 1px solid black'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[2]][R] + ',' + pures[perm[2]][G] + ',' + pures[perm[2]][B] + ')'"
          />
          <span style="margin-left: 60px;">&nbsp;</span>
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[0]][R] + ',' + pures[perm[0]][G] + ',' + pures[perm[0]][B] + ')'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[1]][R] + ',' + pures[perm[1]][G] + ',' + pures[perm[1]][B] + '); border: 1px solid black'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[2]][R] + ',' + pures[perm[2]][G] + ',' + pures[perm[2]][B] + ')'"
          />
          <span style="margin-left: 60px;">&nbsp;</span>
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + puresDimmed[perm[0]][R] + ',' + puresDimmed[perm[0]][G] + ',' + puresDimmed[perm[0]][B] + ')'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + pures[perm[1]][R] + ',' + pures[perm[1]][G] + ',' + pures[perm[1]][B] + '); border: 1px solid black'"
          />
          <span
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + puresDimmed[perm[2]][R] + ',' + puresDimmed[perm[2]][G] + ',' + puresDimmed[perm[2]][B] + ')'"
          />
          <br>
        </div>

        <h2>control</h2>
        All colors in this rainbow must have the same perceived brightness: <br><br>
        <span
          v-for="(c, i) of rainbowSameI"
          :key="i"
          style="display: inline-block; width: 4px; height: 40px;"
          :style="'background-color: rgb(' + c.chan[R] + ',' + c.chan[G] + ',' + c.chan[B] + ')'"
        />
        <br><br>

        <h1>Adjustement of overwhite</h1>
        The three rows must transition from grey to pure at the same speed.<br>
        On the right, the middle square must feel as different from its left square as from its right square. <br><br>
        <div
          v-for="(row, k) of purityGradientJ"
          :key="k"
          style="border: 0px;"
        >
          <span
            v-for="(col, m) of row"
            :key="m"
          >
            <div
              style="display: inline-block; width: 2px; height: 40px;"
              :style="'background-color: rgb(' + col[R] + ',' + col[G] + ',' + col[B] + ')'"
            />
          </span>
          <span style="margin-left: 60px;">&nbsp;</span>
          <div
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + row[10][R] + ',' + row[10][G] + ',' + row[10][B] + ')'"
          />
          <div
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + row[row.length/2][R] + ',' + row[row.length/2][G] + ',' + row[row.length/2][B] + ')'"
          />
          <div
            style="display: inline-block; width: 40px; height: 40px;"
            :style="'background-color: rgb(' + row[row.length-11][R] + ',' + row[row.length-11][G] + ',' + row[row.length-11][B] + ')'"
          />
          <span style="margin-left: 60px;">&nbsp;</span>overwhite[{{ k }}]
          <br>
        </div>

        <div style="background-color: rgb(160,160,160); padding: 10px">
          <h1>Adjustement of iotaD</h1>
          In the first column, the framed color must feel dimmer than its neighbors.<br>
          In the middle column, the framed color must feel as bright as its neighbors.<br>
          In the last column, the framed color must feel brighter than its neighbors.<br><br>
          <div
            v-for="(prim, k) of primariesInPures"
            :key="k"
          >
            <span
              style="display: inline-block; width: 40px; height: 40px;"
              :style="'background-color: rgb(' + pures[prim][R] + ',' + pures[prim][G] + ',' + pures[prim][B] + ')'"
            />
            <span
              style="display: inline-block; width: 40px; height: 40px;"
              :style="'background-color: rgb(' + greyishDimmed[prim][R] + ',' + greyishDimmed[prim][G] + ',' + greyishDimmed[prim][B] + '); border: 1px solid black'"
            />
            <span
              style="display: inline-block; width: 40px; height: 40px;"
              :style="'background-color: rgb(' + pures[prim][R] + ',' + pures[prim][G] + ',' + pures[prim][B] + ')'"
            />
            <span style="margin-left: 60px;">&nbsp;</span>
            <span
              style="display: inline-block; width: 40px; height: 40px;"
              :style="'background-color: rgb(' + pures[prim][R] + ',' + pures[prim][G] + ',' + pures[prim][B] + ')'"
            />
            <span
              style="display: inline-block; width: 40px; height: 40px;"
              :style="'background-color: rgb(' + greyish[prim][R] + ',' + greyish[prim][G] + ',' + greyish[prim][B] + '); border: 1px solid black'"
            />
            <span
              style="display: inline-block; width: 40px; height: 40px;"
              :style="'background-color: rgb(' + pures[prim][R] + ',' + pures[prim][G] + ',' + pures[prim][B] + ')'"
            />
            <span style="margin-left: 60px;">&nbsp;</span>
            <span
              style="display: inline-block; width: 40px; height: 40px;"
              :style="'background-color: rgb(' + puresDimmed[prim][R] + ',' + puresDimmed[prim][G] + ',' + puresDimmed[prim][B] + ')'"
            />
            <span
              style="display: inline-block; width: 40px; height: 40px;"
              :style="'background-color: rgb(' + greyish[prim][R] + ',' + greyish[prim][G] + ',' + greyish[prim][B] + '); border: 1px solid black'"
            />
            <span
              style="display: inline-block; width: 40px; height: 40px;"
              :style="'background-color: rgb(' + puresDimmed[prim][R] + ',' + puresDimmed[prim][G] + ',' + puresDimmed[prim][B] + ')'"
            />
            <br>
          </div>
          <br>
          All of the previous parameters have an influence here and are supposed to have been properly adjusted.
          Therefore, this step acts indirectly as a control of the previous steps. <br>
          If a row here shows a discrepancy, that might indicate that a previous parameter was imperfectly set.
          Try to readjust the sensicol value or the overwhite value corresponding to the test that fails here.
          <br>
        </div>

        <h1>Adjustement of rPowers</h1>
        Adjust rPowers to give a feeling of similarity and linearity to these 6 gradients. <br>
        There are two ways to help yourself with this task: <br>
        - on the left side, try to make each row progress at the same speed, <br>
        - on the right side, try to make the middle square as different from its left square as from its right square.
        <br><br>
        <span
          v-for="(c, i) of 31"
          :key="i"
          style="display: inline-block; width: 6px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[180-i].chan[R] + ',' + rainbowSameJ[180-i].chan[G] + ',' + rainbowSameJ[180-i].chan[B] + ')'"
        />
        <span
          style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[180].chan[R] + ',' + rainbowSameJ[180].chan[G] + ',' + rainbowSameJ[180].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[165].chan[R] + ',' + rainbowSameJ[165].chan[G] + ',' + rainbowSameJ[165].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[150].chan[R] + ',' + rainbowSameJ[150].chan[G] + ',' + rainbowSameJ[150].chan[B] + ')'"
        />
        <span style="margin-left: 60px;">&nbsp;</span>rPowers[0]
        <br>
        <span
          v-for="(c, i) of 31"
          :key="i"
          style="display: inline-block; width: 6px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[i].chan[R] + ',' + rainbowSameJ[i].chan[G] + ',' + rainbowSameJ[i].chan[B] + ')'"
        />
        <span
          style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[0].chan[R] + ',' + rainbowSameJ[0].chan[G] + ',' + rainbowSameJ[0].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[15].chan[R] + ',' + rainbowSameJ[15].chan[G] + ',' + rainbowSameJ[15].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[30].chan[R] + ',' + rainbowSameJ[30].chan[G] + ',' + rainbowSameJ[30].chan[B] + ')'"
        />
        <br>
        <span
          v-for="(c, i) of 31"
          :key="i"
          style="display: inline-block; width: 6px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[60-i].chan[R] + ',' + rainbowSameJ[60-i].chan[G] + ',' + rainbowSameJ[60-i].chan[B] + ')'"
        />
        <span
          style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[60].chan[R] + ',' + rainbowSameJ[60].chan[G] + ',' + rainbowSameJ[60].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[45].chan[R] + ',' + rainbowSameJ[45].chan[G] + ',' + rainbowSameJ[45].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[30].chan[R] + ',' + rainbowSameJ[30].chan[G] + ',' + rainbowSameJ[30].chan[B] + ')'"
        />
        <span style="margin-left: 60px;">&nbsp;</span>rPowers[1]
        <br>
        <span
          v-for="(c, i) of 31"
          :key="i"
          style="display: inline-block; width: 6px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[60+i].chan[R] + ',' + rainbowSameJ[60+i].chan[G] + ',' + rainbowSameJ[60+i].chan[B] + ')'"
        />
        <span
          style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[60].chan[R] + ',' + rainbowSameJ[60].chan[G] + ',' + rainbowSameJ[60].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[75].chan[R] + ',' + rainbowSameJ[75].chan[G] + ',' + rainbowSameJ[75].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[90].chan[R] + ',' + rainbowSameJ[90].chan[G] + ',' + rainbowSameJ[90].chan[B] + ')'"
        />
        <br>
        <span
          v-for="(c, i) of 31"
          :key="i"
          style="display: inline-block; width: 6px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[120-i].chan[R] + ',' + rainbowSameJ[120-i].chan[G] + ',' + rainbowSameJ[120-i].chan[B] + ')'"
        />
        <span
          style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[120].chan[R] + ',' + rainbowSameJ[120].chan[G] + ',' + rainbowSameJ[120].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[105].chan[R] + ',' + rainbowSameJ[105].chan[G] + ',' + rainbowSameJ[105].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[90].chan[R] + ',' + rainbowSameJ[90].chan[G] + ',' + rainbowSameJ[90].chan[B] + ')'"
        />
        <span style="margin-left: 60px;">&nbsp;</span>rPowers[2]
        <br>
        <span
          v-for="(c, i) of 31"
          :key="i"
          style="display: inline-block; width: 6px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[120+i].chan[R] + ',' + rainbowSameJ[120+i].chan[G] + ',' + rainbowSameJ[120+i].chan[B] + ')'"
        />
        <span
          style="display: inline-block; margin-left: 80px; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[120].chan[R] + ',' + rainbowSameJ[120].chan[G] + ',' + rainbowSameJ[120].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[135].chan[R] + ',' + rainbowSameJ[135].chan[G] + ',' + rainbowSameJ[135].chan[B] + ')'"
        />
        <span
          style="display: inline-block; width: 40px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[150].chan[R] + ',' + rainbowSameJ[150].chan[G] + ',' + rainbowSameJ[150].chan[B] + ')'"
        />
        <br><br>
        Control: in the rainbow, the widths of the primary and secondary smudges must all look equal.<br><br>
        <span
          v-for="(c, i) of rainbowSameJ"
          :key="i"
          style="display: inline-block; width: 3px; height: 40px;"
          :style="'background-color: rgb(' + c.chan[R] + ',' + c.chan[G] + ',' + c.chan[B] + ')'"
        />
        <span
          v-for="(c, i) of 31"
          :key="i"
          style="display: inline-block; width: 3px; height: 40px;"
          :style="'background-color: rgb(' + rainbowSameJ[i].chan[R] + ',' + rainbowSameJ[i].chan[G] + ',' + rainbowSameJ[i].chan[B] + ')'"
        />
        <br><br>
      </div>
    </TabPanel>
    <TabPanel header="Color distances">
      <div style="background-color: rgb(128,128,128); padding: 20px">
        <div class="all-colors">
          Here we flatten the Eye color space. r, p and i progress linearly. <br>
          <b>Click two colors to see their distance.</b> <br>
          <br>
          <div
            v-for="(rRow, r) of allColors"
            :key="r"
          >
            <div
              v-for="(pRow, p) of rRow"
              :key="p"
            >
              <div
                v-for="(c, i) of pRow"
                :key="i"
                style="display: inline-block; width: 20px; height: 20px;"
                :style="'background-color: rgb(' + c.rgb.chan[R] + ',' + c.rgb.chan[G] + ',' + c.rgb.chan[B] + ')'"
                @click="showDistanceTo(r, p, i)"
              />
            </div>
            <br>
          </div>
          <div
            v-if="distanceBetweenLastTwoColors"
            class="meter"
          >
            Distance between the last two colors <br>
            that you clicked: {{ distanceBetweenLastTwoColors }} <br><br>
            {{ allColors[krkpki1[0]][krkpki1[1]][krkpki1[2]].eye }} <br>
            {{ allColors[krkpki2[0]][krkpki2[1]][krkpki2[2]].eye }}
          </div>
        </div>
      </div>
    </TabPanel>
  </TabView>
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
