<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faCaretRight } from '@fortawesome/pro-solid-svg-icons'

interface Props {
  translationPath?: string,
}
const props = defineProps<Props>()
const { t: $t } = useTranslation()

const questions = computed(() => {
  const list = []
  // eslint-disable-next-line no-constant-condition
  while (true) {
    const path: string = `${props.translationPath}.${list.length}`
    const question = tD($t, `${path}.question`)
    if (!question) {
      break
    }
    else {
      list.push({
        answers: tAll($t, `${path}.answer`),
        linkLabel: tD($t, `${path}.link.label`),
        linkPath: tD($t, `${path}.link.path`),
        path,
        question,
      })
    }
  }
  return list
})
</script>

<template>
  <div
    v-if="questions.length"
    class="faq-container"
  >
    <h1>FAQ</h1>
    <Accordion class="accordion">
      <AccordionTab
        v-for="quest in questions"
        :key="quest.path"
        :header="quest.question"
      >
        <template #headericon>
          <FontAwesomeIcon :icon="faCaretRight" />
        </template>
        <div class="answer">
          <p
            v-for="(answer, index) in quest.answers"
            :key="index"
          >
            {{ answer }}
          </p>
          <div
            v-if="quest.linkPath"
            class="footer"
          >
            <BcLink
              :to="quest.linkPath"
              target="_blank"
              class="link"
            >
              {{ quest.linkLabel }}
            </BcLink>
          </div>
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
  .answer {
    padding: 16px;
  }
}
</style>
