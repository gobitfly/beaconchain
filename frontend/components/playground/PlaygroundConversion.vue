<script setup lang="ts">
const store = useLatestStateStore()
const { latestState } = storeToRefs(store)

const currentEpoch = computed(
  () => (latestState.value?.current_slot || 0) * 32,
)
</script>

<template>
  <div>
    Conversions
  </div>
  <b> Format numbers </b>
  <div>100000, no settings: <BcFormatNumber :value="100000" /></div>
  <div>100000.1234, no settings: <BcFormatNumber :value="100000.1234" /></div>
  <div>
    100000, min 2 decimals: <BcFormatNumber
      :value="100000"
      :min-decimals="2"
    />
  </div>
  <div>
    100000.1234, min/max 3:
    <BcFormatNumber
      :value="100000.1234"
      :min-decimals="3"
      :max-decimals="3"
    />
  </div>
  <div>0, no settings: <BcFormatNumber :value="0" /></div>
  <div>0.00001, no settings: <BcFormatNumber :value="0.00001" /></div>
  <div>0.01, no settings: <BcFormatNumber :value="0.01" /></div>
  <div>no value, no settings: <BcFormatNumber /></div>
  <div>no value, default '-': <BcFormatNumber default="-" /></div>
  <div>-100000, no settings: <BcFormatNumber :value="-100000" /></div>

  <b> Format percent </b>

  <div>
    1234567.89123, color, +:
    <BcFormatPercent
      :percent="1234567.89123"
      :color-break-point="80"
      :add-positive-sign="true"
    />
  </div>
  <div>
    -1234567.89123, color, +:
    <BcFormatPercent
      :percent="-1234567.89123"
      :color-break-point="80"
      :add-positive-sign="true"
    />
  </div>

  <div>
    1 - no settings
    <BcFormatPercent :percent="1" />
  </div>
  <div>
    percent equal
    <BcFormatPercent
      :percent="85.123"
      :compare-percent="85.43"
    />
  </div>
  <div>
    percent lower
    <BcFormatPercent
      :percent="85.123"
      :compare-percent="85.7"
    />
  </div>
  <div>
    percent lower
    <BcFormatPercent
      :percent="85.123"
      :compare-percent="84.5"
    />
  </div>
  <div>
    Epoch 1 ->
    <BcFormatTimePassed :value="1" />
  </div>

  <div>
    Epoch 272684 ->
    <BcFormatTimePassed :value="272684" />
  </div>
  <div>
    latest Epoch ->
    <BcFormatTimePassed :value="currentEpoch" />
  </div>
  <div>
    latest Epoch absolute ->
    <BcFormatTimePassed
      :value="currentEpoch"
      format="absolute"
    />
  </div>
  <div>
    latest Epoch relative ->
    <BcFormatTimePassed
      :value="currentEpoch"
      format="relative"
    />
  </div>
  <div>
    latest Epoch - 1 ->
    <BcFormatTimePassed :value="currentEpoch - 1" />
  </div>
  <div>
    latest Epoch - 10 ->
    <BcFormatTimePassed :value="currentEpoch - 10" />
  </div>
  <div>
    next Epoch ->
    <BcFormatTimePassed :value="currentEpoch + 1" />
  </div>
  <div>
    next Epoch no tick ->
    <BcFormatTimePassed
      :value="currentEpoch + 1"
      :no-update="true"
    />
  </div>
  <div>
    the Epoch after ->
    <BcFormatTimePassed :value="currentEpoch + 2" />
  </div>
  <div>
    latest Epoch long format->
    <BcFormatTimePassed
      :value="currentEpoch"
      unit-length="long"
    />
  </div>
  <div>
    latest Epoch short format->
    <BcFormatTimePassed
      :value="currentEpoch"
      unit-length="short"
    />
  </div>
</template>

<style lang="scss" scoped>
:deep(.bad-color) {
  color: pink;
}
</style>
