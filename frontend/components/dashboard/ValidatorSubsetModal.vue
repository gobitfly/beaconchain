<script lang="ts" setup>
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
      <Button class="p-button-icon-only copy_button_position">
        <i class="fas fa-copy" />
      </Button>
    </div>
  </BcDialog>
</template>

<style lang="scss" scoped>
 :global(.validator_subset_modal_container) {
    width: 450px;
    height: 569px;
  }

  .top_line_container {
    padding: 10px 0px 14px 0px;
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
    background-color: var(--subcontainer-background);
    padding: 10px 10px 7px 10px;
    border: 1px solid var(--container-border-color);
    border-radius: var(--border-radius);
    min-height: 125px;
    max-height: 453px;
    overflow-y: auto;
    word-break: break-all;
  }

  .copy_button_position {
    position: absolute;
    bottom: 10px;
    right: 10px;
  }
</style>
