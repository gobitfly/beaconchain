<script lang="ts" setup>
import { get } from 'lodash-es'

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

const data = computed(() => {
  console.log('products.value?.premium_products', products.value?.premium_products)
  const sorted = products.value?.premium_products?.sort((a, b) => a.price_per_month_eur - b.price_per_month_eur) ?? []
  const rows: CompareRow[] = []
  const addRow = (type: RowType, property?: string, comingSoon = false) => {
    const row: CompareRow = { type, comingSoon }
    switch (type) {
      case 'header':
        row.values = sorted.map(p => ({ value: p.product_name }))
        break
      case 'group':
        row.label = $t(`pricing.groups.${property}`)
        row.values = sorted.map(p => ({ }))
        break
      case 'perc':
        row.label = $t(`pricing.percs.${property}`)
        row.values = sorted.map((p) => {
          if (!property) {
            return {}
          }
          if (property === 'support_us') {
            return { value: p.price_per_month_eur > 0 }
          }
          const value = get(p.premium_perks, property)
          return {
            value: get(p.premium_perks, property)
          }
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

  addRow('group', 'notification', true)

  addRow('group', 'mobille_app')

  /**
   ad_free
configure_notifications_via_api
custom_machine_alerts
email_notifications_per_day
heatmap_history_seconds
machine_monitoring_history_seconds
manage_dashboard_via_api
mobile_app_custom_themes
mobile_app_widget
monitor_machines
share_custom_dashboards
summary_chart_history_seconds
validator_dashboards
validator_group_notifications
validator_groups_per_dashboard
validators_per_dashboard
webhook_endpoints
   */

  console.log('rows', rows)
  return rows
})

</script>

<template>
  <div class="compare-plans-container">
    {{ data }}
  </div>
</template>

<script lang="ts" setup>

</script>

<style lang="scss" scoped>
.compare-plans-container {
  width: 100%;
  height: 500px;

  background-color: var(--container-background);
  border: 2px solid var(--container-border-color);
  border-radius: 7px;
  font-size: 50px;

  display: flex;
  justify-content: center;
  align-items: center;

  margin-bottom: 43px;
}
</style>
