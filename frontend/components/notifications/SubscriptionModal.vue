<script setup lang="ts">
import {
  faInfoCircle
} from '@fortawesome/pro-regular-svg-icons'
import type { ValidatorSubscriptionState, AccountSubscriptionState } from '~/types/subscriptionModal'

interface Props {
  validatorSub?: ValidatorSubscriptionState,
  accountSub?: AccountSubscriptionState,
  premiumUser: boolean

}

const { props, dialogRef } = useBcDialog<Props>({ showHeader: false })
const { t } = useI18n()

const tPath = computed(() => props.value?.validatorSub ? 'notifications.subscriptions.validators' : 'notifications.subscriptions.accounts')

const closeDialog = () => {
  const changements = true
  dialogRef?.value.close(changements)
}
</script>
{{ tOf($t, 'cookies.text', 0) }}
<template>
  <div class="content">
    <div class="title">
      {{ t('notifications.subscriptions.dialog_title') }}
    </div>
    <div class="explanation">
      {{ t(tPath+'.explanation') }}
    </div>

    <div v-if="!!props?.validatorSub">
      validators
    </div>

    <div v-else-if="!!props?.accountSub">
      accounts
    </div>

    <BcTooltip :fit-content="true">
      <FontAwesomeIcon :icon="faInfoCircle" />
      <template #tooltip />
    </BcTooltip>
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

  .footer {
    display: flex;
    justify-content: center;
    margin-top: var(--padding);
    gap: var(--padding);
  }
}
</style>
