<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faCaretRight
} from '@fortawesome/pro-solid-svg-icons'
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
  <div v-if="questions.length" class="faq-container">
    <h1>FAQ</h1>
    <Accordion class="accordion">
      <AccordionTab v-for="quest in questions" :key="quest.path" :header="quest.question">
        <template #headericon>
          <FontAwesomeIcon :icon="faCaretRight" />
        </template>
        <p v-for="(anser, index) in quest.answers" :key="index">
          {{ anser }}
        </p>
        <div class="footer">
          <BcLink v-if="quest.linkPath" :path="quest.linkPath" class="link">
            {{ quest.linkLabel }}
          </BcLink>
        </div>
      </AccordionTab>
    </Accordion>
  </div>
</template>

<style lang="scss" scoped>
.faq-container {
  display: flex;
  flex-direction: column;
  align-items: center;

  h1 {
    margin-top: 0;
    margin-bottom: 22px;
  }

  p {
    margin-top: var(--padding);
    margin-bottom: var(--padding);
  }

  .accordion {
    width: 100%;
  }

  .footer {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
