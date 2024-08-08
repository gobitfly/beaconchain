<script setup lang="ts">
interface Props {
  graffiti?: string
  noCopy?: boolean
}
const props = defineProps<Props>()
</script>

<template>
  <div
    v-if="props.graffiti"
    class="format-graffiti"
  >
    <BcTooltip class="tt-container">
      <template #tooltip>
        <div class="tt">
          {{ props.graffiti }}
        </div>
      </template>
      {{ props.graffiti }}
    </BcTooltip>
    <BcCopyToClipboard
      v-if="!props.noCopy"
      :value="props.graffiti"
      class="copy"
    />
  </div>
  <div v-else>
    -
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

.format-graffiti {
  width: 100%;
  display: flex;

  .tt-container {
    @include utils.truncate-text;
  }

  .copy {
    margin-left: var(--padding);
    line-height: 100%;
  }
}

.tt {
  max-width: 200px;
  text-align: left;
}
</style>
