<script lang="ts" setup>
const props = defineProps<{
  optionIdentifier: string,
  optionLabel: string,
  options: any[],
  optionValue: string,
  screenreaderHeading?: string,
  screenreaderText: string,
  text: string,
}>()

const hasOptions = computed(() => !!props.options.length)

const popover = ref()

const isVisible = ref(false)
const toggle = (event: Event) => {
  popover.value?.toggle(event)
  isVisible.value = !isVisible.value
}
const idScreenreaderHeading = useId()
const handleFocus = () => {
  const screenreaderHeadingElement = document.getElementById(idScreenreaderHeading)
  screenreaderHeadingElement?.focus()
}
const emit = defineEmits<{
  change: [{
    id: number,
    value: boolean,
  }],
}>()
</script>

<template>
  <span>
    <button
      type="button"
      class="bc-dropdown-toggle__button"
      :aria-label="screenreaderText"
      @click="hasOptions && toggle($event)"
    >
      <span>{{ text }}</span>
      <IconChevron
        v-if="hasOptions"
        width="0.5rem"
        :direction="isVisible ? 'left' : 'bottom'"
      />
    </button>

    <Popover
      ref="popover"
      unstyled
      @keydown.esc.stop
      @show="handleFocus"
      @hide="isVisible = false"
    >
      <ul
        class="content"
      >
        <BcScreenreaderOnly
          :id="idScreenreaderHeading"
          tabindex="-1"
          tag="h2"
        >
          {{ props.screenreaderHeading }}
        </BcScreenreaderOnly>
        <li
          v-for="option in props.options"
          :key="option[props.optionIdentifier]"
          class="content-item"
        >
          <label
            class="content-item__label"
            :for="`${option[props.optionIdentifier]}`"
          >
            {{ option[optionLabel] }}
          </label>
          <BcToggle
            v-model="option.is_subscribed"
            :input-id="`${option[props.optionIdentifier]}`"
            @update:model-value="emit('change', {
              id: option[props.optionIdentifier],
              value: option[props.optionValue],
            })"
          />
        </li>
      </ul>
    </Popover>
  </span>
</template>

<style lang="scss" scoped>
.content {
  margin-block: 0.25rem;
  max-width: 25rem;
  min-width: 15rem;
  padding: var(--padding);
  border-radius: var(--border-radius);
  border: 1px solid var(--input-border-color);
  background-color: var(--input-background);
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.content-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.625rem;
    &__label {
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
}
.bc-dropdown-toggle__button {
  padding-inline: 0.5rem;
  padding-block: 0.25rem;
  width: 9rem;
  background-color: var(--input-background);
  border-radius: var(--border-radius);
  border: 1px solid var(--input-border-color);
  color: var(--input-active-text-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
