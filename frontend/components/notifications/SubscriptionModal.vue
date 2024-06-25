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

const closeDialog = (response: boolean) => {
  dialogRef?.value.close(response)
}
</script>

<template>
  <div class="content">
    <div class="title">
      {{ props?.title }}
    </div>
    <div class="question">
      {{ props?.question }}
    </div>
    <BcTooltip :fit-content="true">
      <FontAwesomeIcon :icon="faInfoCircle" />
      <template #tooltip>
        <div class="info-label-list">
          <div v-for="info in props.data.infos" :key="info.label">
            <div><b>{{ info.label }}:</b> {{ info.value }}</div>
          </div>
        </div>
      </template>
    </BcTooltip>
    <div class="footer">
      <Button type="button" :label="props?.noLabel || $t('navigation.no')" @click="closeDialog(false)" />
      <Button type="button" :label="props?.yesLabel || $t('navigation.yes')" @click="closeDialog(true)" />
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

  .question {
    flex-grow: 1;
    margin: var(--padding) 0;
    @include fonts.subtitle_text;
  }

  .footer {
    display: flex;
    justify-content: center;
    margin-top: var(--padding);
    gap: var(--padding);
  }
}
</style>
