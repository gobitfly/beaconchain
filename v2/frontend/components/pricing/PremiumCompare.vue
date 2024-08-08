<script lang="ts" setup>
import { get } from 'lodash-es'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { PremiumPerks } from '~/types/api/user'

const { t: $t } = useTranslation()
const { products } = useProductsStore()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

type CompareValue = {
  value?: string | boolean
  tooltip?: string
  class?: string
}

type RowType = 'header' | 'group' | 'perc' | 'label'

type CompareRow = {
  type: RowType
  label?: string
  subText?: string
  values?: CompareValue[]
  className?: string
}

const showContent = ref(false)

const rows = computed(() => {
  const sorted
    = products.value?.premium_products?.toSorted(
      (a, b) => a.price_per_month_eur - b.price_per_month_eur,
    ) ?? []
  const rows: CompareRow[] = []
  const mapValue = (property: string, perks: PremiumPerks): CompareValue => {
    if (['support_us', 'bulk_adding'].includes(property)) {
      return { value: perks.ad_free }
    }
    let value = get(perks, property)

    if (!value) {
      value = false
    }
    else if (property.includes('_seconds')) {
      if (value === Number.MAX_SAFE_INTEGER) {
        value = $t('pricing.full_history')
      }
      else {
        value = $t('common.last_x', {
          duration: formatTimeDuration(value as number, $t),
        })
      }
    }

    let tooltip: string | undefined
    if (property === 'validators_per_dashboard') {
      tooltip = $t('pricing.pectra_tooltip', {
        effectiveBalance: formatNumber(perks.validators_per_dashboard * 32),
      })
    }

    return {
      value,
      tooltip,
    }
  }
  const addRow = (
    type: RowType,
    property?: string,
    className?: string,
    subText?: string,
    hidePositiveValues = false,
    translationKey?: string,
  ) => {
    const row: CompareRow = { type, subText, className }
    switch (type) {
      case 'header':
        row.values = sorted.map(p => ({ value: p.product_name }))
        break
      case 'group':
        row.label = $t(`pricing.groups.${property}`)
        row.values = sorted.map(_p => ({}))
        break
      case 'label':
        row.label = $t(translationKey || `pricing.percs.${property}`)
        row.values = sorted.map(_p => ({}))
        break
      case 'perc':
        row.label = $t(translationKey || `pricing.percs.${property}`)
        row.values = sorted.map((p) => {
          if (!property) {
            return {}
          }
          const mv = mapValue(property, p.premium_perks)
          if (hidePositiveValues && mv.value) {
            mv.value = $t('common.soon')
            mv.class = 'soon'
          }
          return mv
        })
        break
    }
    rows.push(row)
  }

  const comingSoon = $t('pricing.premium_product.coming_soon')

  addRow('header')

  addRow('group', 'general')
  addRow('perc', 'ad_free', 'first-in-group')
  addRow('perc', 'support_us', 'last-in-group')

  addRow('group', 'dashboard')
  addRow('perc', 'validator_dashboards', 'first-in-group')
  addRow('perc', 'validators_per_dashboard')
  addRow('perc', 'validator_groups_per_dashboard')
  addRow('perc', 'share_custom_dashboards')
  addRow('perc', 'manage_dashboard_via_api', undefined, comingSoon)
  addRow(
    'perc',
    'bulk_adding',
    'last-in-group',
    $t('pricing.percs.bulk_adding_subtext'),
  )
  addRow('group', 'dashboard_charts')
  addRow('label', 'summary_chart_history', 'first-in-group')
  const chartProps = ['epoch', 'hourly', 'daily', 'weekly']
  chartProps.forEach(p =>
    addRow(
      'perc',
      `chart_history_seconds.${p}`,
      undefined,
      undefined,
      undefined,
      `time_frames.${p}`,
    ),
  )

  addRow('label', 'heatmap_history', 'last-in-group', comingSoon)

  addRow(
    'group',
    'notification',
    undefined,
    showInDevelopment ? undefined : comingSoon,
  )
  addRow(
    'perc',
    'email_notifications_per_day',
    'first-in-group',
    undefined,
    !showInDevelopment,
  )
  addRow('perc', 'configure_notifications_via_api')

  addRow(
    'perc',
    'validator_group_notifications',
    undefined,
    undefined,
    !showInDevelopment,
  )
  addRow(
    'perc',
    'webhook_endpoints',
    'last-in-group',
    undefined,
    !showInDevelopment,
  )

  addRow('group', 'mobille_app')
  addRow('perc', 'mobile_app_custom_themes', 'first-in-group')
  addRow('perc', 'mobile_app_widget')
  addRow('perc', 'monitor_machines')
  addRow('perc', 'machine_monitoring_history_seconds')
  addRow(
    'perc',
    'custom_machine_alerts',
    'last last-in-group',
    $t('pricing.percs.custom_machine_alerts_subtext'),
  )

  return rows
})
</script>

<template>
  <div class="compare-plans-container">
    <h1>{{ $t("pricing.compare") }}</h1>
    <div
      class="content"
      :class="{ 'show-content': showContent }"
    >
      <div
        v-for="(row, index) in rows"
        :key="index"
        :class="[row.type, row.className]"
        class="row"
      >
        <div class="label">
          <span>{{ row.label }}</span>
          <span
            v-if="row.subText"
            class="sub-text"
          > {{ row.subText }}</span>
        </div>
        <div
          v-for="(value, vIndex) in row.values"
          :key="vIndex"
          class="value"
          :class="value.class"
        >
          <span v-if="typeof value.value === 'boolean'">
            <BcFeatureCheck :available="value.value" />
          </span>
          <span v-else>
            {{ value.value }}
          </span>
          <BcTooltip
            v-if="value.tooltip"
            :fit-content="true"
            :text="value.tooltip"
            class="info-icon"
          >
            <FontAwesomeIcon :icon="faInfoCircle" />
          </BcTooltip>
        </div>
      </div>
      <BcBlurOverlay class="blur" />
    </div>
    <div
      class="button-row"
      :class="{ 'show-content': showContent }"
    >
      <Button
        class="pricing_button"
        @click="() => (showContent = !showContent)"
      >
        {{ $t(showContent ? "pricing.hide_feature" : "pricing.show_feature") }}
      </Button>
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
    overflow-x: hidden;
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

    &.show-content {
      max-height: unset;
      overflow-x: auto;

      .blur {
        display: none;
      }
    }

    .row {
      display: flex;
      gap: 7px;
      min-height: 51px;
      width: calc(100% - 1px);
      min-width: fit-content;
      border-left: 1px solid transparent;

      &.label,
      &.header,
      &.group {
        font-size: 18px;
        font-weight: 600;

        @media (max-width: 1360px) {
          font-size: 12px;
        }

        .label {
          padding-left: 21px;
        }
      }

      &.label,
      &.perc {
        min-height: 36px;
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
        text-align: right;
        min-width: 121px;

        .sub-text {
          font-size: 11px;
          margin-bottom: -1px;

          @media (max-width: 1360px) {
            font-size: 12px;
            margin-bottom: 4px;
          }
        }

        @media (max-width: 1360px) {
          justify-content: flex-start;
          text-align: left;
          align-content: baseline;
          align-self: center;
          padding-left: 21px;
          gap: 0;
        }
      }

      &.group {
        border-left: var(--border-style);
        border-top: var(--border-style);
        border-bottom: var(--border-style);
        border-top-left-radius: var(--border-radius);
        border-bottom-left-radius: var(--border-radius);

        .label {
          .sub-text {
            font-size: 13px;
            margin-bottom: -2px;

            @media (max-width: 1360px) {
              font-size: 12px;
              margin-bottom: unset;
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

        &.soon {
          font-style: italic;
        }
      }

      &.last {
        .value {
          border-bottom: var(--border-style);

          border-bottom-left-radius: var(--border-radius);
          border-bottom-right-radius: var(--border-radius);
        }
      }

      &.first-in-group {
        min-height: 42px;

        .label,
        .value {
          padding-top: 6px;
        }
      }

      &.last-in-group {
        min-height: 42px;

        .label,
        .value {
          padding-bottom: 6px;
        }
      }
    }
  }

  .button-row {
    margin-top: 25px;
    width: 100%;
    display: flex;
    justify-content: center;

    &.show-content {
      margin-top: 75px;

      @media (max-width: 1360px) {
        margin-top: 15px;
      }
    }
  }
}
</style>
