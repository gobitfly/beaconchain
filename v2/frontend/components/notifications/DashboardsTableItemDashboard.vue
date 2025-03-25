<script lang="ts" setup>
import type { DashboardType } from '~/types/dashboard'

interface Props {
  dashboardId: number,
  dashboardName: string,
  type: DashboardType,
}
const props = defineProps<Props>()

const entityLink = computed(() => {
  if (props.type === 'validator') {
    return `/dashboard/${props.dashboardId}`
  }
  if (props.type === 'account') {
    return `/account-dashboard/${props.dashboardId}`
  }
  return ''
})
</script>

<template>
  <div>
    <BcLink
      :to="entityLink"
      class="link link-dashboard"
      target="_blank"
    >
      <BcIcon
        v-if="props.type === 'validator'"
        name="desktop"
      />
      <span class="truncate-text">
        {{ props.dashboardName }}
      </span>
    </BcLink>
  </div>
</template>

<style lang="scss">
$breakpoint-lg: 1024px;

.link-dashboard {
  display: flex;
  align-items: center;
  gap: var(--padding-small);
  @media (min-width: $breakpoint-lg) {
    gap: var(--padding);
  }
}
.icon-dashboard-type {
  height: 14px;
  width: 16px;
  flex-shrink: 0;
}
</style>
