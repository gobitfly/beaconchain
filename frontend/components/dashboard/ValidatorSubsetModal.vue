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
  <BcDialog v-model="visible" :header="header" class="modal_container">
    <div class="top_line_container">
      <span class="subtitle_text">
        {{ props.caption }}
      </span>
      <span class="search_elements_container">
        <InputText v-model="filter" placeholder="Index" class="remove_right_border_radius" />
        <Button class="p-button-icon-only remove_left_border_radius">
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
 :global(.modal_container) {
    width: 450px;
    height: 569px;
  }

  .top_line_container {
    padding: 10px 0px 14px 0px;
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
  }

  .search_elements_container {
    display: flex;
    align-items: center;
  }

  .remove_right_border_radius{
    border-top-right-radius: 0px;
    border-bottom-right-radius: 0px;
  }

  .remove_left_border_radius{
    border-top-left-radius: 0px;
    border-bottom-left-radius: 0px;
  }

  .text_container {
    padding: 10px 10px 7px 10px;
    border: 1px solid var(--container-border-color);
    border-radius: var(--border-radius);
    height: 453px;
  }

  .copy_button_position {
    position: absolute;
    bottom: 10px;
    right: 10px;
  }
</style>
