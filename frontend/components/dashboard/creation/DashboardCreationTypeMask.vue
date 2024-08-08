<script lang="ts" setup>
import { type DashboardType } from '~/types/dashboard'
import { IconAccount, IconValidator } from '#components'

const { t: $t } = useTranslation()
const { isLoggedIn } = useUserStore()

interface Props {
  accountsDisabled: boolean
  validatorsDisabled: boolean
}
const props = defineProps<Props>()

const type = defineModel<DashboardType | ''>('type', { required: true })

const typeButtons = [
  {
    text: $t('dashboard.creation.type.validators'),
    value: 'validator',
    component: IconValidator,
    disabled: props.validatorsDisabled,
  },
  {
    text: $t('dashboard.creation.type.accounts'),
    subText: $t('common.coming_soon'),
    value: 'account',
    component: IconAccount,
    disabled: props.accountsDisabled,
  },
]

const name = defineModel<string>('name', { required: true })

const emit = defineEmits<{ (e: 'next'): void }>()

const continueDisabled = computed(() => {
  return (
    type.value === ''
    || name.value === ''
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
