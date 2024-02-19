<script lang="ts" setup>
interface Props {
  caption?: string,
}
const props = defineProps<Props>()

const header = ref('Validator Subset Modal')
const visible = defineModel<boolean>()
</script>

<template>
  <BcDialog v-model="visible" :header="header">
    <div>
      {{ props.caption }} Search Element
    </div>
    <div>
      Validator List and copy button
    </div>
  </BcDialog>
</template>
