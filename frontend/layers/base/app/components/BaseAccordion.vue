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
    class="flex flex-col gap-2xl justify-evenly"
    type="multiple"
  >
    <template
      v-for="item in accordionItems"
      :key="item.value"
    >
      <RkAccordionItem
        v-slot="{ open }"
        class=" flex flex-col p-3xl bg-white dark:bg-gray-950 border border-gray-100 dark:border-gray-900 rounded-md"
        :value="item.value"
      >
        <RkAccordionHeader>
          <RkAccordionTrigger
            class="text-left group flex gap-md justify-between items-center w-full text-xl font-bold cursor-pointer"
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
          class="data-[state=open]:animate-slide-down data-[state=closed]:animate-slide-up overflow-clip font-medium"
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
