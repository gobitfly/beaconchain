<script lang="ts" setup>
import { warn } from 'vue'

interface Props {
  caption?: string,
  validators: number[],
}
const props = defineProps<Props>()

const visible = defineModel<boolean>()
const header = ref<string>('Validator Subset Modal')
const filter = ref<string>('')
const shownValidators = ref<number[]>([])

watch(filter, (newFilter) => {
  if (newFilter === '') {
    shownValidators.value = props.validators
    return
  }

  shownValidators.value = []

  const index = parseInt(newFilter)
  if (props.validators.includes(index)) {
    shownValidators.value = [index]
  }
}, { immediate: true })

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
      <span class="search_elements_container">
        <InputText v-model="filter" placeholder="Index" />
        <Button class="p-button-icon-only">
          <i class="fas fa-magnifying-glass" />
        </Button>
      </span>
    </div>
    <div class="text_container">
      <span v-for="(v, i) in shownValidators" :key="v">
        <NuxtLink :to="`/validator/${v}`" class="link">
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
    padding: var(--padding) 0px 14px 0px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  // TODO: This will become its own component in the near future
  .search_elements_container {
    display: flex;
    align-items: center;

    :first-child{
      border-top-right-radius: 0px;
      border-bottom-right-radius: 0px;
      height: var(--default-button-height);
    }

    :last-child{
      border-top-left-radius: 0px;
      border-bottom-left-radius: 0px;
    }
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
