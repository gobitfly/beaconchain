<script setup lang="ts">
import {
  faCheck,
  faEdit
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

interface Props {
  value?: string,
  label?: string, // used if not in edit mode, defaults to value,
  disabled?: boolean,
  canBeEmpty?: boolean,
  maxlength?: number,
}

const props = defineProps<Props>()
const inputRef = ref<ComponentPublicInstance | null>(null)

const emit = defineEmits<{(e: 'setValue', value: string): void }>()

const isEditing = ref(false)
const editValue = ref<string>(props.value ?? '')

const iconClick = () => {
  if (props.disabled) {
    return
  }
  if (!isEditing.value) {
    isEditing.value = true
    return
  }
  if (!editValue.value && !props.canBeEmpty) {
    return
  }
  if (editValue.value !== props.value) {
    emit('setValue', editValue.value)
  }

  isEditing.value = false
}

const icon = computed(() => ({
  icon: isEditing.value ? faCheck : faEdit,
  disabled: props.disabled || (isEditing.value && (!editValue.value && !props.canBeEmpty))
}))

watch(() => props.value, (v) => {
  editValue.value = v ?? ''
})

watch([isEditing, inputRef], ([edit, input]) => {
  if (edit) {
    input?.$el?.focus()
  }
})

</script>

<template>
  <div class="input-container">
    <div v-if="isEditing" class="input-wrapper">
      <InputText ref="inputRef" v-model="editValue" :maxlength="maxlength" @keypress.enter="iconClick" />
    </div>
    <span v-if="!isEditing" class="label">
      {{ label || value }}
    </span>
    <FontAwesomeIcon v-if="!disabled" class="link" :icon="icon.icon" :disabled="icon.disabled" @click="iconClick" />
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

.input-container {
  display: flex;
  gap: var(--padding);
  align-items: center;
  height: 24px;

  .input-wrapper {
    flex-grow: 1;

    input {
      width: 100%;
      height: 100%;
    }
  }

  .label {
    flex-grow: 1;
    margin-left: 8px;
    @include utils.truncate-text;
  }

  .link{
    margin-right: var(--padding);
  }
}
</style>
