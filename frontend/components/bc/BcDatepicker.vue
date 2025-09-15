<script setup lang="ts">
const {
  dateFormat = 'M dd, yy',
  label,
} = defineProps<{
  dateFormat?: string,
  id: string,
  label: string,
  maxDate: Date,
  minDate: Date,
}>()

const modelValue = defineModel<Date>({
  required: true,
})
</script>

<template>
  <label
    :id
    class="bc-datepicker__label"
  >
    <span>{{ label }}</span>
    <PvDatePicker
      :id
      v-model="modelValue"
      :date-format
      :min-date
      :max-date
    />
  </label>
</template>

<style lang="scss">
.bc-datepicker__label {
  display: flex;
  flex-direction: column;
  font-size: 0.8rem;
  color: var(--text-color-discreet);
}

div.p-datepicker-input {
  max-width: min-content;
}

div.p-datepicker-panel {
  background: var(--input-background);
  width: 268px;
  margin-top: var(--padding-tiny);
  padding: var(--padding-small);
  border-radius: var(--border-radius);
  border: 1px solid var(--input-border-color);

  .p-datepicker-header {
    padding: var(--padding-small);
    padding-bottom: var(--padding);
    border-bottom: 1px solid var(--light-grey-7);

    .p-datepicker-prev-button,
    .p-datepicker-next-button {
      height: 30px;
    }

    .p-datepicker-title {
      gap: var(--padding);

      .p-datepicker-select-month,
      .p-datepicker-select-year {
        font-size: 1rem
      }
    }
  }

  .p-datepicker-day-view {
    margin-top: var(--padding-small);

    .p-datepicker-weekday-cell {
      padding: var(--padding-tiny);
    }

    .p-datepicker-day-cell {
      padding: var(--padding-tiny);

      &.p-datepicker-other-month {
        color: var(--text-color-disabled);

        &.p-datepicker-today > .p-datepicker-day {
          color: var(--white);
        }
      }

      &.p-datepicker-today > .p-datepicker-day {
        background-color: var(--grey);
      }

      .p-datepicker-day {
        border-radius: 50%;
        width: 30px;
        height: 30px;
        transition: background-color 0.2s, color 0.2s, border-color 0.2s;

        &.p-datepicker-day-selected {
          background-color: var(--primary-color);
        }

        &:not(.p-datepicker-day-selected):not(.p-disabled):hover {
          background-color: var(--grey);
        }
      }

      .p-disabled {
        color: var(--text-color-disabled);
      }
    }
  }

  .p-datepicker-month-view {
    margin-top: var(--padding-small);

    .p-datepicker-month {
      padding: var(--padding-small);
      border-radius: var(--border-radius);
      transition: background-color 0.2s, color 0.2s, border-color 0.2s;

      &.p-datepicker-month-selected {
        background-color: var(--primary-color);
      }

      &:not(.p-datepicker-month-selected):not(.p-disabled):hover {
        background-color: var(--grey);
      }
    }

    .p-disabled {
      color: var(--text-color-disabled);
    }
  }

  .p-datepicker-year-view {
    margin-top: var(--padding-small);

    .p-datepicker-year {
      padding: var(--padding-small);
      border-radius: var(--border-radius);
      transition: background-color 0.2s, color 0.2s, border-color 0.2s;

      &.p-datepicker-year-selected {
        background-color: var(--primary-color);
      }

      &:not(.p-datepicker-year-selected):not(.p-disabled):hover {
        background-color: var(--grey);
      }
    }

    .p-disabled {
      color: var(--text-color-disabled);
    }
  }
}
</style>
