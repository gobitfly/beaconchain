<script setup lang="ts">
import type { MenuBarEntry } from '~/types/menuBar'

interface Props {
  alignRight?: boolean,
  buttons?: MenuBarEntry[],
}
defineProps<Props>()
</script>

<template>
  <PvMenubar
    v-if="buttons?.length"
    :model="buttons"
    breakpoint="0px"
    :class="{ 'right-aligned-submenu': alignRight }"
  >
    <template #item="{ item }">
      <component
        :is="item.component"
        v-if="item.component"
        class="button-content"
      />
      <BcTooltip
        v-else-if="item.disabledTooltip"
        :text="item.disabledTooltip"
        class="button-content"
        @click.stop.prevent="() => undefined"
      >
        <span class="text-disabled text">{{ item.label }}</span>
      </BcTooltip>
      <BcLink
        v-else-if="item.route && !item.command"
        :to="item.route"
        class="pointer"
        :class="{ 'p-active': item.active }"
      >
        <span
          class="button-content"
          :class="[item.class]"
        >
          <span class="text">{{ item.label }}</span>
          <BcIcon
            v-if="item.dropdown"
            name="chevron-down"
            class="toggle"
          />
        </span>
      </BcLink>
      <span
        v-else
        class="button-content pointer"
        :class="[item.class, { 'p-active': item.active }]"
        :highlight="item.highlight || null"
      >
        <BcIcon
          v-if="item.faIcon"
          :name="item.faIcon"
          class="icon"
        />
        <span
          v-if="item.label"
          class="text"
        >{{ item.label }}</span>
        <BcIcon
          v-if="item.dropdown && (!item.faIcon || item.label)"
          name="chevron-down"
          class="toggle"
        />
      </span>
    </template>
  </PvMenubar>
</template>

<style lang="scss" scoped>
</style>
