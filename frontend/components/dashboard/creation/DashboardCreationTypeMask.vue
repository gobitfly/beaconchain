<script lang="ts" setup>
import type { Icon } from '~/components/bc/icon/BcIcon.vue'

const { t: $t } = useTranslation()
const { isLoggedIn } = useUserStore()

interface Props {
  validatorsDisabled: boolean,
}
const props = defineProps<Props>()

const typeButtons = [
  {
    disabled: props.validatorsDisabled,
    icon: 'desktop' as Icon,
    text: $t('dashboard.creation.type.validators'),
    value: 'validator',
  },
  {
    disabled: true,
    icon: 'user' as Icon,
    subText: $t('common.coming_soon'),
    text: $t('dashboard.creation.type.accounts'),
    value: 'account',
  },
]

const name = defineModel<string>('name', { required: true })

const emit = defineEmits<{ (e: 'next'): void }>()

const continueDisabled = computed(() => {
  return (
    name.value === ''
    || name.value.length > 32
    || !REGEXP_VALID_NAME.test(name.value)
  )
})

const next = () => {
  name.value = name.value.trim()
  if (continueDisabled.value) {
    return
  }

  emit('next')
}
const type = 'validator'
</script>

<template>
  <div class="mask-container">
    <div class="element-container">
      <div class="big_text">
        {{ $t("dashboard.creation.title") }}
      </div>
      <div class="subtitle_text">
        {{ $t("dashboard.creation.type.subtitle") }}
      </div>
      <BcToggleSingleBar
        v-model="type"
        class="single-bar"
        :buttons="typeButtons"
        layout="gaudy"
      />
      <div class="row-container">
        <InputText
          v-if="isLoggedIn"
          v-model="name"
          :placeholder="$t('dashboard.creation.type.placeholder')"
          class="input-field"
          @keypress.enter="next"
        />
        <Button
          class="button"
          :disabled="continueDisabled"
          @click="next"
        >
          {{ $t("navigation.continue") }}
        </Button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.mask-container {
  width: 100%;
  .element-container {
    display: flex;
    flex-direction: column;
    gap: var(--padding);

    .single-bar {
      height: 100px;
    }

    .row-container {
      display: flex;
      justify-content: flex-end;
      gap: var(--padding);

      input {
        min-width: 250px;
        max-width: 320px;
        width: 100%;
      }

      button {
        width: 90px;
      }
    }
  }
}
</style>
