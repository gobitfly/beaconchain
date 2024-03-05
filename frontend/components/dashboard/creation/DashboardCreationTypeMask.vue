<script lang="ts" setup>
import { type DashboardCreationState, type DashboardType } from '~/types/dashboard/creation'
import { IconAccount, IconValidator } from '#components'

const { t: $t } = useI18n()

const type = defineModel<DashboardType>('type', { required: true })
const state = defineModel<DashboardCreationState>('state', { required: true })
const typeButtons = ref([{ text: $t('dashboard.creation.type.accounts'), value: 'account', component: IconAccount }, { text: $t('dashboard.creation.type.validators'), value: 'validator', component: IconValidator }])

const name = defineModel<string>('name', { required: true })

const continueDisabled = computed(() => {
  return type.value === '' || name.value === ''
})

function onContinue () {
  if (type.value === 'account') {
    state.value = ''
  } else {
    state.value = 'network'
  }
}
</script>

<template>
  <div class="mask_container">
    <div class="element_container">
      <div class="big_text">
        {{ $t('dashboard.creation.title') }}
      </div>
      <div class="subtitle_text">
        {{ $t('dashboard.creation.type.subtitle') }}
      </div>
      <BcToggleSingleBar v-model="type" :buttons="typeButtons" :initial="type" />
      <div class="row_container">
        <InputText v-model="name" :placeholder="$t('dashboard.creation.type.name')" class="input-field" />
        <Button class="button" :disabled="continueDisabled" @click="onContinue()">
          {{ $t('dashboard.creation.continue') }}
        </Button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .mask_container{
    width: 460px;
    height: 248px;
    padding: var(--padding-large);

    .element_container{
      display: flex;
      flex-direction: column;
      gap: 10px;

      .row_container{
        display: flex;
        gap: 10px;

        input {
            width: 320px;
        }

        button {
            width: 90px;
        }
      }
    }
  }
</style>
