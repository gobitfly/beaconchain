<script setup lang="ts">
export type AccordionItem = {
  content: string,
  title: string,
}
const { items } = defineProps<{
  items: AccordionItem[],
}>()
const accordionItems = computed(() => {
  return items.map((item, index) => ({
    ...item,
    value: `item-${index + 1}`,
  }))
})
</script>

<template>
  <RkAccordionRoot
    class="flex flex-col justify-evenly gap-2xl"
    type="multiple"
  >
    <template
      v-for="item in accordionItems"
      :key="item.value"
    >
      <RkAccordionItem
        v-slot="{ open }"
        class=" flex flex-col rounded-md border border-gray-100 bg-white p-3xl dark:border-gray-900 dark:bg-gray-950"
        :value="item.value"
      >
        <RkAccordionHeader>
          <RkAccordionTrigger
            class="group flex w-full cursor-pointer items-center justify-between gap-md text-left text-xl font-bold"
          >
            <span>{{ item.title }}</span>
            <BaseIcon
              :name="open ? 'minus' : 'plus'"
              class="transition-transform duration-300 group-data-[state=open]:-rotate-180"
              aria-hidden="true"
            />
          </RkAccordionTrigger>
        </RkAccordionHeader>
        <RkAccordionContent
          style="--slide-height: var(--reka-accordion-content-height);"
          class="overflow-clip font-medium data-[state=closed]:animate-slide-up data-[state=open]:animate-slide-down"
        >
          <BaseText
            is="div"
            variant="secondary"
            class="mt-2xl"
          >
            {{ item.content }}
          </BaseText>
        </RkAccordionContent>
      </RkAccordionItem>
    </template>
  </RkAccordionRoot>
</template>

<style scoped></style>
