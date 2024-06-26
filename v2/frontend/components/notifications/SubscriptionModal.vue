<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { ValidatorSubscriptionState, AccountSubscriptionState } from '~/types/subscriptionModal'

interface Props {
  validatorSub?: ValidatorSubscriptionState,
  accountSub?: AccountSubscriptionState,
  premiumUser: boolean

}

const { props, dialogRef } = useBcDialog<Props>({ showHeader: false })
const { t } = useI18n()

const newDataReceived = ref<number>(0) // used in <template> to trigger Vue to refresh or hide the content
let tPath: string
let validatorSubModifiable: ValidatorSubscriptionState
let accountSubModifiable: AccountSubscriptionState

watch(props, (props) => {
  if (!props || (!props.validatorSub && !props.accountSub)) {
    newDataReceived.value = 0
    return
  }
  newDataReceived.value++
  if (props.validatorSub) {
    tPath = 'notifications.subscriptions.validators.'
    validatorSubModifiable = structuredClone(toRaw(props.validatorSub))
  } else {
    tPath = 'notifications.subscriptions.accounts.'
    accountSubModifiable = structuredClone(toRaw(props.accountSub!))
  }
}, { immediate: true })

const closeDialog = () => {
  const changements = true
  dialogRef?.value.close(changements)
}
</script>

<template>
  <div v-if="newDataReceived" class="content">
    <div class="title">
      {{ t('notifications.subscriptions.dialog_title') }}
    </div>

    <div v-if="props?.validatorSub">
      <div class="explanation">
        {{ t(tPath+'explanation') }}
      </div>
      <div class="option-row">
        {{ t(tPath+'offlineValidator.option') }}
        <BcTooltip :fit-content="true">
          <FontAwesomeIcon :icon="faInfoCircle" />
          <template #tooltip>
            <div class="tt-content">
              {{ tOf(t, tPath+'offlineValidator.hint', 0) }}
              <ul><li>{{ tOf(t, tPath+'offlineValidator.hint', 1) }}</li> <li>{{ tOf(t, tPath+'offlineValidator.hint', 2) }}</li> <li>{{ tOf(t, tPath+'offlineValidator.hint', 3) }}</li> </ul>
            </div>
          </template>
        </BcTooltip>
      </div>
    </div>

    <div v-else-if="props?.accountSub">
      <div class="explanation">
        {{ t(tPath+'explanation') }}
      </div>
    </div>

    <div class="footer">
      <Button type="button" :label="t('notifications.subscriptions.save')" @click="closeDialog" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/fonts.scss';

.content {
  display: flex;
  flex-direction: column;

  .title {
    @include fonts.subtitle_text;
    color: var(--primary-color);
    margin-bottom: var(--padding-small);
  }

  .explanation {
    color: var(--text-color-disabled);
  }

  .option-row {

  }

  .footer {
    display: flex;
    justify-content: center;
    margin-top: var(--padding);
    gap: var(--padding);
  }
}

.tt-content {
    width: 220px;
    min-width: 100%;
    text-align: left;
    ul {
      margin-left: 0;
      padding-left: 1.5em;
    }
  }
</style>
