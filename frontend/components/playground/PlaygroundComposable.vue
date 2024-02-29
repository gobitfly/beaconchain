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
        <Button @click="bounceTrue(!isTrueTemp, true)">
          Toggle if no timer
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
        <Button @click="bounceNum(numTemp+1, true)">
          Toggle if no timer
        </Button>
        <Button @click="instantNum(numTemp+1)">
          Instant Toggle
        </Button>
      </div>
    </TabPanel>
  </TabView>
</template>

<style lang="scss" scoped>
.element_container {
  margin: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
</style>
