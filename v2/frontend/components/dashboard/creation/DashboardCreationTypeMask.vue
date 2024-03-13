<script lang="ts" setup>
import { type DashboardType } from '~/types/dashboard'
import { IconAccount, IconValidator } from '#components'

const { t: $t } = useI18n()

const type = defineModel<DashboardType | ''>('type', { required: true })
const typeButtons = [{ text: $t('dashboard.creation.type.accounts'), value: 'account', component: IconAccount }, { text: $t('dashboard.creation.type.validators'), value: 'validator', component: IconValidator }]

const name = defineModel<string>('name', { required: true })

const emit = defineEmits<{(e: 'next'): void }>()

const continueDisabled = computed(() => {
  return type.value === '' || name.value === '' || name.value.length > 32
})
</script>

<template>
  <div class="mask-container">
    <div class="element-container">
      <div class="big_text">
        {{ $t('dashboard.creation.title') }}
      </div>
      <div class="subtitle_text">
        {{ $t('dashboard.creation.type.subtitle') }}
      </div>
      <BcToggleSingleBar v-model="type" class="single-bar" :buttons="typeButtons" :initial="type" />
      <div class="row-container">
        <InputText v-model="name" :placeholder="$t('dashboard.creation.type.placeholder')" class="input-field" />
        <Button class="button" :disabled="continueDisabled" @click="emit('next')">
          {{ $t('navigation.continue') }}
        </Button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .mask-container{
    width: 100%;
    .element-container{
      display: flex;
      flex-direction: column;
      gap: var(--padding);

      .single-bar{
        height: 100px;
      }

      .row-container{
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
