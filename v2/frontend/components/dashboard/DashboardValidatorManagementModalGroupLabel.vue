<script setup lang="ts">
interface Props {
  canBeEmpty?: boolean,
  disabled?: boolean,
  label?: string, // used if not in edit mode, defaults to value,
  maxlength?: number,
  pattern?: RegExp,
  value?: string,
}

const props = defineProps<Props>()
const inputRef = ref<ComponentPublicInstance | null>(null)

const emit = defineEmits<{ (e: 'setValue', value: string): void }>()

const isEditing = ref(false)
const editValue = ref<string>(props.value ?? '')

const iconClick = () => {
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

watch(
  () => props.value,
  (v) => {
    editValue.value = v ?? ''
  },
)

watch([
  isEditing,
  inputRef,
], ([
  edit,
  input,
]) => {
  if (edit) {
    input?.$el?.focus()
  }
})
</script>

<template>
  <div class="input-container">
    <div
      v-if="isEditing"
      class="input-wrapper"
    >
      <PvInputText
        ref="inputRef"
        v-model.trim="editValue"
        :maxlength
        @keypress.enter="iconClick"
      />
    </div>
    <span
      v-if="!isEditing"
      class="label"
    >
      {{ label || value }}
    </span>
    <BcButtonIcon
      :screenreader-text="{
        key: 'dashboard.validator.group_management.edit_group_name',
        interpolation: { groupName: props.value },
      }"
      :disabled="props.disabled
        || (isEditing && !editValue && !props.canBeEmpty)
        || (props.pattern && !props.pattern.test(editValue))"
      class="edit-button"
      :name="isEditing ? 'check' : 'edit'"
      @click="iconClick"
    />
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

  .edit-button {
    color: var(--blue);
    margin-right: var(--padding);
  }

  > svg[disabled] {
    color: var(--button-color-disabled);
  }
}
</style>
