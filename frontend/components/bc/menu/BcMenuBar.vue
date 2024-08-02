<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { MenuBarEntry } from '~/types/menuBar'

interface Props {
  buttons?: MenuBarEntry[],
  alignRight?: boolean
}
defineProps<Props>()

</script>

<template>
  <Menubar v-if="buttons?.length" :model="buttons" breakpoint="0px" :class="{ 'right-aligned-submenu': alignRight }">
    <template #item="{ item }">
      <BcTooltip
        v-if="item.disabledTooltip"
        :text="item.disabledTooltip"
        class="button-content"
        @click.stop.prevent="() => undefined"
      >
        <span class="text-disabled">{{ item.label }}</span>
      </BcTooltip>
      <BcLink
        v-else-if="item.route && !item.command"
        :to="item.route"
        class="pointer"
        :class="{ 'p-active': item.active }"
      >
        <span class="button-content" :class="[item.class]">
          <span class="text">{{ item.label }}</span>
          <IconChevron v-if="item.dropdown" class="toggle" direction="bottom" />
        </span>
      </BcLink>
      <span
        v-else
        class="button-content pointer"
        :class="[item.class, { 'p-active': item.active }]"
        :highlight="item.highlight || null"
      >
        <FontAwesomeIcon v-if="item.faIcon" :icon="item.faIcon" class="icon" />
        <span v-if="item.label" class="text">{{ item.label }}</span>
        <IconChevron v-if="item.dropdown && (!item.faIcon || item.label)" class="toggle" direction="bottom" />
      </span>
    </template>
  </Menubar>
</template>

<style lang="scss" scoped>
</style>
