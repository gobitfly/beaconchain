<script lang="ts" setup>
import { warn } from 'vue'

interface Props {
  caption?: string,
  validators: number[],
}
const props = defineProps<Props>()

const visible = defineModel<boolean>()
const header = ref<string>('Validator Subset Modal')
const shownValidators = ref<number[]>(props.validators)

const handleEvent = (filter: string) => {
  if (filter === '') {
    shownValidators.value = props.validators
    return
  }

  shownValidators.value = []

  const index = parseInt(filter)
  if (props.validators.includes(index)) {
    shownValidators.value = [index]
  }
}

watch(visible, (value) => {
  if (!value) {
    shownValidators.value = props.validators
  }
})

function copyValidatorsToClipboard (): void {
  if (shownValidators.value.length === 0) {
    return
  }

  let text = ''
  shownValidators.value.forEach((v, i) => {
    text += v
    if (i !== shownValidators.value.length - 1) {
      text += ','
    }
  })
  navigator.clipboard.writeText(text)
    .catch((error) => {
      warn('Error copying text to clipboard:', error)
    })
}
</script>

<template>
  <BcDialog v-model="visible" :header="header" class="validator_subset_modal_container">
    <div class="top_line_container">
      <span class="subtitle_text">
        {{ props.caption }}
      </span>
      <BcContentFilter @filter-changed="handleEvent" />
    </div>
    <div class="text_container">
      <span v-for="(v, i) in shownValidators" :key="v">
        <NuxtLink :to="`/validator/${v}`" target="blank" class="link">
          {{ v }}
        </NuxtLink>
        <span v-if="i !== shownValidators.length - 1">, </span>
      </span>
    </div>
    <Button class="p-button-icon-only copy_button" @click="copyValidatorsToClipboard">
      <i class="fas fa-copy" />
    </Button>
  </BcDialog>
</template>

<style lang="scss" scoped>
 :global(.validator_subset_modal_container) {
    width: 450px;
    height: 569px;
  }

  :global(.validator_subset_modal_container .p-dialog-content) {
      display: flex;
      flex-direction: column;
      flex-grow: 1;
  }

  :global(.validator_subset_modal_container .p-dialog-content .copy_button) {
    position: absolute;
    bottom: calc(var(--padding-large) + var(--padding));
    right: calc(var(--padding-large) + var(--padding));
  }

  .top_line_container {
    padding: var(--padding) 0 14px 0;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .text_container {
    position: relative;
    flex-grow: 1;
    background-color: var(--subcontainer-background);
    padding: var(--padding) var(--padding) 7px var(--padding);
    border: 1px solid var(--container-border-color);
    border-radius: var(--border-radius);
    height: 453px;
    overflow-y: auto;
    word-break: break-all;
  }
</style>
