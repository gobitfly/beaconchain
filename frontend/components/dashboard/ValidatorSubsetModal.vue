<script lang="ts" setup>
interface Props {
  caption?: string,
  validators: number[],
}
const props = defineProps<Props>()

const header = ref('Validator Subset Modal')
const visible = defineModel<boolean>()
</script>

<template>
  <BcDialog v-model="visible" :header="header">
    <div>
      <span class="">
        {{ props.caption }}
      </span> Search Element
    </div>
    <div>
      <span v-for="(v, i) in props.validators" :key="v">
        <NuxtLink :to="`/validator/${v}`" class="link">
          {{ v }}
        </NuxtLink>
        <span v-if="i !== props.validators.length - 1">, </span>
      </span>
      Copy button<br>
    </div>
  </BcDialog>
</template>
