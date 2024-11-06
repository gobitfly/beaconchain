<script setup lang="ts">
import type { NotificationSettingsValidatorDashboard } from '~/types/api/notifications'

const {
  dialogRef,
  props,
} = useBcDialog<NotificationSettingsValidatorDashboard>({ showHeader: false })
const { t: $t } = useTranslation()
const {
  secondsPerEpoch,
} = useNetworkStore()

const { userPremiumPerks } = useUserStore()
const hasPremiumPerkGroupEfficiency = computed(
  () => userPremiumPerks.value?.notifications_validator_dashboard_group_efficiency,
)
function closeDialog(): void {
  dialogRef?.value.close()
}

const checkboxes = ref({
  is_attestations_missed_subscribed: props.value?.is_attestations_missed_subscribed ?? false,
  is_block_proposal_subscribed: props.value?.is_block_proposal_subscribed ?? false,
  is_group_efficiency_below_subscribed: props.value?.is_group_efficiency_below_subscribed ?? false,
  is_max_collateral_subscribed: props.value?.is_max_collateral_subscribed ?? false,
  is_min_collateral_subscribed: props.value?.is_min_collateral_subscribed ?? false,
  is_slashed_subscribed: props.value?.is_slashed_subscribed ?? false,
  is_sync_subscribed: props.value?.is_sync_subscribed ?? false,
  is_upcoming_block_proposal_subscribed: props.value?.is_upcoming_block_proposal_subscribed ?? false,
  is_validator_offline_subscribed: props.value?.is_validator_offline_subscribed ?? false,
  is_withdrawal_processed_subscribed: props.value?.is_withdrawal_processed_subscribed ?? false,
})
const thresholds = ref({
  group_efficiency_below_threshold: formatFraction(props.value?.group_efficiency_below_threshold ?? 0),
  max_collateral_threshold: formatFraction(props.value?.max_collateral_threshold ?? 0),
  min_collateral_threshold: formatFraction(props.value?.min_collateral_threshold ?? 0),
})
const emit = defineEmits<{
  (e: 'change-settings', settings: Omit<NotificationSettingsValidatorDashboard,
  | 'is_webhook_discord_enabled'
  | 'webhook_url'>): void,
}>()
watchDebounced([
  checkboxes,
  thresholds,
], () => {
  emit('change-settings', {
    ...checkboxes.value,
    group_efficiency_below_threshold: Number(formatToFraction(thresholds.value.group_efficiency_below_threshold)),
    max_collateral_threshold: Number(formatToFraction(thresholds.value.max_collateral_threshold)),
    min_collateral_threshold: Number(formatToFraction(thresholds.value.min_collateral_threshold)),
  })
}, {
  deep: true,
})

const hasAllEvents = ref(Object.values(checkboxes.value).every(value => value === true))
watch(hasAllEvents, () => {
  if (hasAllEvents.value) {
    (Object.keys(checkboxes.value) as Array<keyof typeof checkboxes.value>)
      .forEach((key) => {
        if (
          key === 'is_group_efficiency_below_subscribed'
          && !hasPremiumPerkGroupEfficiency.value
        ) {
          return
        }
        checkboxes.value[key] = true
      })
    return
  }
  (Object.keys(checkboxes.value) as Array<keyof typeof checkboxes.value>)
    .forEach((key) => {
      checkboxes.value[key] = false
    })
})
</script>

<template>
  <div
    class="content"
  >
    <div class="title">
      {{ $t("notifications.subscriptions.title") }}
    </div>

    <div class="explanation">
      {{ $t('notifications.subscriptions.validators.explanation') }}
    </div>
    <div
      class="row-container"
    >
      <BcSettings>
        <BcSettingsRow
          v-model:checkbox="checkboxes.is_validator_offline_subscribed"
          :label="$t('notifications.subscriptions.validators.validator_is_offline.label')"
        >
          <template #info>
            <BcTranslation
              keypath="notifications.subscriptions.validators.validator_is_offline.info.template"
              listpath="notifications.subscriptions.validators.validator_is_offline.info._list"
            />
          </template>
        </BcSettingsRow>
        <BcSettingsRow
          v-model:checkbox="checkboxes.is_attestations_missed_subscribed"
          :label="$t('notifications.subscriptions.validators.attestation_missed.label')"
          :info="$t('notifications.subscriptions.validators.attestation_missed.info', { count: Number(formatSecondsTo(secondsPerEpoch, { minimumFractionDigits: 1 }).minutes) })"
        />
        <BcSettingsRow
          v-model:checkbox="checkboxes.is_block_proposal_subscribed"
          :label="$t('notifications.subscriptions.validators.block_proposal.label')"
        />
        <BcSettingsRow
          v-model:checkbox="checkboxes.is_upcoming_block_proposal_subscribed"
          :label="$t('notifications.subscriptions.validators.upcoming_block_proposal.label')"
          :info="$t('notifications.subscriptions.validators.upcoming_block_proposal.info')"
        />
        <BcSettingsRow
          v-model:checkbox="checkboxes.is_sync_subscribed"
          :label="$t('notifications.subscriptions.validators.sync_committee.label')"
        />
        <BcSettingsRow
          v-model:checkbox="checkboxes.is_withdrawal_processed_subscribed"
          :label="$t('notifications.subscriptions.validators.withdrawal_processed.label')"
        />
        <BcSettingsRow
          v-model:checkbox="checkboxes.is_slashed_subscribed"
          :label="$t('notifications.subscriptions.validators.validator_got_slashed.label')"
        />
        <BcSettingsRow
          v-model:checkbox="checkboxes.is_group_efficiency_below_subscribed"
          v-model:input="thresholds.group_efficiency_below_threshold"
          has-unit
          :info="$t('notifications.subscriptions.validators.group_efficiency.info', { percentage: thresholds.group_efficiency_below_threshold })"
          :label="$t('notifications.subscriptions.validators.group_efficiency.label')"
          :has-premium-gem="!hasPremiumPerkGroupEfficiency"
        />
        <BcSettingsRow
          v-model:checkbox="checkboxes.is_min_collateral_subscribed"
          v-model:input="thresholds.min_collateral_threshold"
          has-unit
          :label="$t('notifications.subscriptions.validators.min_collateral_reached.label')"
        />
        <BcSettingsRow
          v-model:checkbox="checkboxes.is_max_collateral_subscribed"
          v-model:input="thresholds.max_collateral_threshold"
          has-unit
          :label="$t('notifications.subscriptions.validators.max_collateral_reached.label')"
        />
        <BcSettingsRow
          v-model:checkbox="hasAllEvents"
          :label="$t('notifications.subscriptions.validators.all_events.label')"
          has-border-top
        />
      </BcSettings>

      <div class="footer">
        <BcButton
          @click="closeDialog"
        >
          {{ $t('navigation.done') }}
        </BcButton>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.content {
  display: flex;
  flex-direction: column;

  .title {
    @include fonts.dialog_header;
    text-align: center;
    margin-bottom: var(--padding-large);
  }

  .explanation {
    @include fonts.small_text;
    text-align: center;
  }

  .row-container {
    position: relative;
    margin-top: 8px;
    margin-bottom: 8px;
    .separation {
      height: 1px;
      background-color: var(--container-border-color);
      margin-bottom: 16px;
    }
  }

  .footer {
    display: flex;
    justify-content: right;
    margin-top: var(--padding);
    gap: var(--padding);
  }
}
</style>
