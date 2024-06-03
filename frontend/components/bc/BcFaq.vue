<script setup lang="ts">
interface Props {
  translationPath?: string
}
const props = defineProps<Props>()
const { t: $t } = useI18n()

const questions = computed(() => {
  const list = []
  let notFound = false
  let index = 0
  while (!notFound) {
    const path = `${props.translationPath}.${index}`
    const question = tD($t, `${path}.question`)
    if (!question) {
      notFound = true
    } else {
      list.push({
        path,
        question,
        answers: tAll($t, `${path}.answer`),
        linkPath: tD($t, `${path}.link.path`),
        linkLabel: tD($t, `${path}.link.label`)
      })
      index++
    }
  }
  return list
})

</script>

<template>
  <div v-if="questions.length">
    <h1>FAQ</h1>
    <Accordion>
      <AccordionTab v-for="quest in questions" :key="quest.path" :header="quest.question">
        <p v-for="(anser, index) in quest.answers" :key="index">
          {{ anser }}
        </p>
        <BcLink v-if="quest.linkPath" :path="quest.linkPath" class="link">
          {{ quest.linkLabel }}
        </BcLink>
      </AccordionTab>
    </Accordion>
  </div>
</template>

<style lang="scss" scoped>
</style>
