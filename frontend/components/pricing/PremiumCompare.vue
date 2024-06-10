<script lang="ts" setup>
import { get } from 'lodash-es'
import {
  faInfoCircle
} from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { PremiumPerks } from '~/types/api/user'

const { t: $t } = useI18n()
const { products } = useProductsStore()

type CompareValue = {
  value?: string | boolean,
  tooltip?: string,
}

type RowType = 'header' | 'group' | 'perc'

type CompareRow = {
  type: RowType,
  label?: string,
  comingSoon?: boolean,
  values?: CompareValue[]
}

const showContent = ref(false)

const rows = computed(() => {
  const sorted = products.value?.premium_products?.sort((a, b) => a.price_per_month_eur - b.price_per_month_eur) ?? []
  const rows: CompareRow[] = []
  const mapValue = (property: string, perks: PremiumPerks) => {
    if (['support_us', 'bulk_adding'].includes(property)) {
      return { value: perks.ad_free }
    }
    let value = get(perks, property)
    if (value === 0) {
      value = false
    } else if (property.includes('_per_day')) {
      value = $t('time_duration.days', { amount: value }, value === 1 ? 1 : 2)
    } else if (property.includes('_seconds')) {
      value = formatTimeDuration(value as number, $t)
    }

    let tooltip: string | undefined
    if (property === 'validators_per_dashboard') {
      tooltip = $t('pricing.pectra_tooltip', { effectiveBalance: formatNumber(perks.validators_per_dashboard * 32) })
    }

    return {
      value,
      tooltip
    }
  }
  const addRow = (type: RowType, property?: string, comingSoon = false) => {
    const row: CompareRow = { type, comingSoon }
    switch (type) {
      case 'header':
        row.values = sorted.map(p => ({ value: p.product_name }))
        break
      case 'group':
        row.label = $t(`pricing.groups.${property}`)
        row.values = sorted.map(_p => ({}))
        break
      case 'perc':
        row.label = $t(`pricing.percs.${property}`)
        row.values = sorted.map((p) => {
          if (!property) {
            return {}
          }
          return mapValue(property, p.premium_perks)
        })
        break
    }
    rows.push(row)
  }
  addRow('header')

  addRow('group', 'general')
  addRow('perc', 'ad_free')
  addRow('perc', 'support_us')

  addRow('group', 'dashboard')
  addRow('perc', 'validator_dashboards')
  addRow('perc', 'validators_per_dashboard')
  addRow('perc', 'validator_groups_per_dashboard')
  addRow('perc', 'share_custom_dashboards')
  addRow('perc', 'manage_dashboard_via_api', true)
  addRow('perc', 'heatmap_history_seconds')
  addRow('perc', 'summary_chart_history_seconds')

  addRow('group', 'notification', true)
  addRow('perc', 'email_notifications_per_day')
  addRow('perc', 'configure_notifications_via_api')
  addRow('perc', 'validator_group_notifications')
  addRow('perc', 'webhook_endpoints')

  addRow('group', 'mobille_app')
  addRow('perc', 'mobile_app_custom_themes')
  addRow('perc', 'mobile_app_widget')
  addRow('perc', 'monitor_machines')
  addRow('perc', 'machine_monitoring_history_seconds')
  addRow('perc', 'custom_machine_alerts')
  return rows
})

</script>

<template>
  <div class="compare-plans-container">
    <h1>{{ $t('pricing.compare') }}</h1>
    <div class="content" :class="{ 'show-content': showContent }">
      <div v-for="(row, index) in rows" :key="index" :class="row.type" class="row">
        <div class="label">
          <span>{{ row.label }}</span>
          <span v-if="row.comingSoon" class="coming-soon"> {{ $t('pricing.premium_product.coming_soon') }}</span>
        </div>
        <div v-for="(value, vIndex) in row.values" :key="vIndex" class="value">
          <span v-if="typeof value.value === 'boolean'">
            <BcFeatureCheck :available="value.value" />
          </span>
          <span v-else>
            {{ value.value }}
          </span>
          <BcTooltip v-if="value.tooltip" :fit-content="true" :text="value.tooltip" class="info-icon">
            <FontAwesomeIcon :icon="faInfoCircle" />
          </BcTooltip>
        </div>
      </div>
      <BcBlurOverlay class="blur" />
      <div class="button-row">
        <Button class="pricing_button" @click="() => showContent = !showContent">
          {{ $t(showContent ? 'pricing.hide_feature' : 'pricing.show_feature') }}
        </Button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.compare-plans-container {
  --border-style: 1px solid var(--container-border-color);
  width: 100%;

  border-radius: 7px;
  font-size: 15px;
  font-weight: 400;

  margin-bottom: 94px;

  @media (max-width: 1360px) {
    font-size: 12px;
    margin-bottom: 45px;
  }

  h1 {
    font-size: 31px;
    font-weight: 600;
    margin: 22px 0 22px 0;
    text-align: center;
    width: 100%;

    @media (max-width: 1360px) {
      font-size: 28px;
    }
  }

  .content {
    overflow-x: auto;
    overflow-y: hidden;
    width: 100%;
    padding-bottom: 8px;
    max-height: 300px;
    position: relative;

    .blur {
      bottom: 0;
      left: 0;
      right: 0;
      height: 75%;
    }

    .button-row {
      position: absolute;
      z-index: 10;
      bottom: 20px;
      left: 0;
      right: 0;
      display: flex;
      justify-content: center;
    }

    &.show-content {
      max-height: unset;

      .blur {
        display: none;
      }

      .button-row {
        position: unset;
        padding-top: 20px;
      }
    }

    .row {
      display: flex;
      gap: 7px;
      min-height: 51px;
      width: calc(100% - 1px);
      border-left: 1px solid transparent;

      &.header,
      &.group {
        font-size: 18px;
        font-weight: 600;

        @media (max-width: 1360px) {
          font-size: 12px;
        }
      }

      &.header {
        .value {
          border-top: var(--border-style);
          border-top-left-radius: var(--border-radius);
          border-top-right-radius: var(--border-radius);
        }
      }

      .label {
        display: flex;
        justify-content: flex-end;
        align-items: center;
        flex-wrap: wrap;
        gap: 4px;
        flex-grow: 1;
        min-height: 100%;
        padding-right: 10px;
        padding-left: 21px;
        text-align: end;
        min-width: 121px;

        .coming-soon {
          font-size: 11px;

          @media (max-width: 1360px) {
            font-size: 12px;
          }
        }

        @media (max-width: 1360px) {
          justify-content: flex-start;
          text-align: left;
        }
      }

      &.group {
        border-left: var(--border-style);
        border-top: var(--border-style);
        border-bottom: var(--border-style);
        border-top-left-radius: var(--border-radius);
        border-bottom-left-radius: var(--border-radius);

        .label {
          .comming-soon {
            font-size: 13px;

            @media (max-width: 1360px) {
              font-size: 12px;
            }
          }
        }
      }

      .value {
        display: flex;
        justify-content: center;
        align-items: center;
        gap: 4px;
        width: 166px;
        flex-shrink: 0;
        min-height: 100%;

        background-color: var(--container-background);

        border-left: var(--border-style);
        border-right: var(--border-style);

        @media (max-width: 1360px) {
          width: 100px;
        }

        .info-icon {
          height: 9px;
          width: 9px;
          display: inline-flex;

          svg {
            width: 100%;
            height: 100%;
          }
        }
      }

      &:last-child {
        .value {
          border-bottom: var(--border-style);

          border-bottom-left-radius: var(--border-radius);
          border-bottom-right-radius: var(--border-radius);
        }
      }
    }

  }
}
</style>
