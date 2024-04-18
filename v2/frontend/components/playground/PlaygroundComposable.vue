<script setup lang="ts">

const { value: isTrue, temp: isTrueTemp, bounce: bounceTrue, instant: instantTrue } = useDebounceValue<boolean>(false, 2000)
const { value: num, temp: numTemp, bounce: bounceNum, instant: instantNum } = useDebounceValue<number>(1, 2000)

</script>

<template>
  <TabView lazy>
    <TabPanel header="Bounce">
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
        <Button @click="bounceNum(numTemp+1)">
          Add
        </Button>
        <Button @click="bounceNum(numTemp+1, false, true)">
          Add endles
        </Button>
        <Button @click="bounceNum(numTemp+1, true)">
          Add if no timer
        </Button>
        <Button @click="bounceNum(numTemp+1, true, true)">
          Add if no timer endles
        </Button>
        <Button @click="instantNum(numTemp+1)">
          Instant Toggle
        </Button>
      </div>
    </TabPanel>
    <TabPanel header="Conversion">
      <PlaygroundConversion />
    </TabPanel>
    <TabPanel header="Toasts">
      <PlaygroundToasts />
    </TabPanel>
    <TabPanel header="Encryption">
      <PlaygroundEncryption />
    </TabPanel>
  </TabView>
</template>

<style lang="scss" scoped>
.element_container {
  margin: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: var(--padding);
}
</style>
