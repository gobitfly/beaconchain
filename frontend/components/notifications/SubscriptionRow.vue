<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { ChainIDs } from '~/types/network'

const props = defineProps<{
  tPath: string,
  lacksPremiumSubscription: boolean
}>()

const { t } = useI18n()

const liveState = defineModel<boolean|number|ChainIDs[]>({ required: true })
const checked = ref<boolean>(false)
const inputted = ref<string>('')

// ###### parser for the formatting of the text in the tooltips

type Text = { content: string, bold: boolean }
type List = Text[]
type TooltipContent = (Text|List)[]

function parseRawArray (raw : string[], start : number, output: TooltipContent) : number {
  if (start >= raw.length || (raw.length === 1 && !raw[0])) {
    return start
  }
  const inAList = raw[start][0] === '-'
  let i = start
  while (i < raw.length) {
    const first = raw[i][0]
    if (inAList && first !== '-') {
      // end of list, we return the index of this first text following the list
      return i
    }
    if (first === '-') {
      if (!inAList) {
        // beginning of a list
        const list: List = []
        output.push(list)
        i = parseRawArray(raw, i, list)
      } else {
        // already in a list
        output.push(parseRawText(raw[i].slice(1)))
        i++
      }
    } else {
      output.push(parseRawText(raw[i]))
      i++
    }
  }

  return i
}

function parseRawText (text: string) : Text {
  if (text[0] === '*') {
    return { content: text.slice(1), bold: true }
  }
  return { content: text, bold: false }
}

// end of parsing ######

const tooltip : ComputedRef<TooltipContent> = computed(() => {
  let options
  if (Array.isArray(liveState.value)) {
    options = { plural: liveState.value.length, count: liveState.value.length, list: liveState.value.join(', ') }
  } else {
    const plural = (typeof liveState.value === 'number') ? liveState.value : (liveState.value ? 2 : 1)
    options = { plural, count: liveState.value }
  }

  const output: TooltipContent = []
  const translation = tAll(t, props.tPath + '.hint', options)
  parseRawArray(translation, 0, output)
  return output
})

const deactivationClass = props.lacksPremiumSubscription ? 'deactivated' : ''
</script>

<template>
  <div class="option-row">
    <span class="caption" :class="deactivationClass">
      {{ t(tPath+'.option') }}
    </span>
    <BcTooltip v-if="tooltip.length" :fit-content="true">
      <FontAwesomeIcon :icon="faInfoCircle" class="info" />
      <template #tooltip>
        <div class="tt-content">
          <span v-for="(element,p) of tooltip" :key="p">
            <span v-if="!Array.isArray(element)">
              <b v-if="element.bold">{{ element.content }}</b>
              <span v-else>{{ element.content }}</span>
            </span>
            <ul v-else>
              <li v-for="line of element" :key="line.content">
                <b v-if="line.bold">{{ line.content }}</b>
                <span v-else>{{ line.content }}</span>
              </li>
            </ul>
          </span>
        </div>
      </template>
    </BcTooltip>
    <BcPremiumGem v-if="lacksPremiumSubscription" class="gem" />
    <div class="right">
      <InputText v-if="typeof liveState == 'number'" v-model="inputted" :placeholder="t(tPath + '.placeholder')" :class="deactivationClass" />
      <Checkbox v-model="checked" :binary="true" :class="deactivationClass" />
    </div>
  </div>
</template>

<style scoped lang="scss">
@use "~/assets/css/fonts.scss";

.deactivated {
  opacity: 0.6;
  pointer-events: none;
}

.option-row {
  display: flex;
  @include fonts.small_text;
  height: 40px;

  .caption {

  }
  .info {
    margin-left: 6px;
  }
  .gem {
    margin-left: 6px;
  }
  .right {
    margin-left: auto;
  }
}

.tt-content {
  width: 220px;
  min-width: 100%;
  text-align: left;
  ul {
    padding: 0;
    margin: 0;
    padding-left: 1.5em;
  }
}
</style>
