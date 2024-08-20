<script setup lang="ts">
import {
  PlaygroundConversion, PlaygroundEncryption, PlaygroundToasts,
} from '#components'
import type { HashTabs } from '~/types/hashTabs'

const {
  bounce: bounceTrue,
  instant: instantTrue,
  temp: isTrueTemp,
  value: isTrue,
} = useDebounceValue<boolean>(false, 2000)
const {
  bounce: bounceNum,
  instant: instantNum,
  temp: numTemp,
  value: num,
} = useDebounceValue<number>(1, 2000)

const tabs: HashTabs = {
  bounce: {
    index: '0',
    title: 'Bounce',
  },
  conversion: {
    component: PlaygroundConversion,
    index: '1',
    title: 'Conversion',
  },
  encryption: {
    component: PlaygroundEncryption,
    index: '3',
    title: 'Encryption',
  },
  toasts: {
    component: PlaygroundToasts,
    index: '2',
    title: 'Toasts',
  },

}
</script>

<template>
  <BcTabList
    :tabs default-tab="summary"
  >
    <template #tab-panel-bounce>
      <div class="element_container">
        Is true: {{ isTrue }} Temp: {{ isTrueTemp }}
        <Button @click="bounceTrue(!isTrueTemp)">
          Toggle
        </Button>
        <Button @click="bounceTrue(!isTrueTemp, false, true)">
          Toggle endles
        </Button>
        <Button @click="bounceTrue(!isTrueTemp, true)">
          Toggle if no timer
        </Button>
        <Button @click="bounceTrue(!isTrueTemp, true, true)">
          Toggle if no timer endles
        </Button>
        <Button @click="instantTrue(!isTrueTemp)">
          Instant Toggle
        </Button>
      </div>
      <div class="element_container">
        Num: {{ num }} Temp: {{ numTemp }}
        <Button @click="bounceNum(numTemp + 1)">
          Add
        </Button>
        <Button @click="bounceNum(numTemp + 1, false, true)">
          Add endles
        </Button>
        <Button @click="bounceNum(numTemp + 1, true)">
          Add if no timer
        </Button>
        <Button @click="bounceNum(numTemp + 1, true, true)">
          Add if no timer endles
        </Button>
        <Button @click="instantNum(numTemp + 1)">
          Instant Toggle
        </Button>
      </div>
    </template>
  </BcTabList>
</template>

<style lang="scss" scoped>
.element_container {
  margin: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: var(--padding);
}
</style>
