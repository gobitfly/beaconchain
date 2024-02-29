<script lang="ts" setup>
interface Props {
  startEpoch: number,
  groupInfos: {
    name: string,
    efficiency: number,
    color: string
  }[]
}

const props = defineProps<Props>()

// TODO: This function is duplicated, don't do that!
function timeToDateString (time: number): string {
  const options: Intl.DateTimeFormatOptions = {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  }
  return new Date(time * 1000).toLocaleDateString(undefined, options)
}

const dateText = computed(() => {
  const ts = epochToTs(props.startEpoch)
  if (ts === undefined) {
    return undefined
  }

  const date = timeToDateString(ts)
  return `${date}`
})

const epochText = computed(() => {
  const endEpoch = props.startEpoch + epochsPerDay()
  return `Epoch ${props.startEpoch} - ${endEpoch}`
})

</script>

<template>
  <div class="tooltip-container">
    <div>
      {{ dateText }}
    </div>
    <div>
      {{ epochText }}
    </div>
    <div v-for="(entry, index) in props.groupInfos" :key="index" class="line-container">
      <div class="circle" :style="{ 'background-color': entry.color }" />
      <div>
        {{ entry.name }}:
      </div>
      <div class="efficiency">
        {{ entry.efficiency }}%
      </div>
    </div>
  </div>
</template>

<style lang="scss">
@use '~/assets/css/fonts.scss';

.tooltip-container {
  @include fonts.tooltip_text_bold;
  background-color: var(--tooltip-background);
  color: var(--tooltip-text-color);
  line-height: 1.5;

  .line-container{
    display: flex;
    align-items: center;
    gap: 3px;

    .circle{
      width: 10px;
      height: 10px;
      border-radius: 50%;
    }

    .efficiency{
      @include fonts.tooltip_text;
    }
  }
}
</style>
