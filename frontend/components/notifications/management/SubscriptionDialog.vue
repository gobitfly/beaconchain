<script setup lang="ts">
import { warn } from 'vue'
import type {
  APIentry,
  InternalEntry,
} from '~/types/notifications/subscriptionModal'
import type {
  NotificationSettingsAccountDashboard,
  NotificationSettingsValidatorDashboard,
} from '~/types/api/notifications'
import { ChainFamily } from '~/types/network'
import type { DashboardType } from '~/types/dashboard'

type AllOptions = NotificationSettingsAccountDashboard &
  NotificationSettingsValidatorDashboard
type DefinedAPIentry = Exclude<APIentry, null | undefined>

interface Props {
  dashboardType: DashboardType,
  initialSettings: AllOptions,
  saveUserSettings: (
    settings: Record<keyof AllOptions, DefinedAPIentry>,
  ) => void,
}

const {
  dialogRef,
  props,
} = useBcDialog<Props>({ showHeader: false })
const { t: $t } = useTranslation()
const { networkInfo } = useNetworkStore()
// const { user } = useUserStore()

function closeDialog(): void {
  dialogRef?.value.close()
}
</script>

<template>
  <pre>
    {{ props }}
  </pre>
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
        <!-- <BcSettingsRow> -->
        <span>
          hello
        </span>
        <span>
          world
        </span>
        <!-- </BcSettingsRow> -->
        <!-- <div>
          <BcInputCheckbox
            :label="$t('notifications.subscriptions.validators.validator_offline')"
          />
        </div>
        <div>
          <BcInputCheckbox
            :label="$t('notifications.subscriptions.validators.group_is_offline')"
          />
        </div> -->
      </BcSettings>

      <div class="footer">
        <Button
          type="button"
          :label="$t('navigation.done')"
          @click="closeDialog"
        />
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
