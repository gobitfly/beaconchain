<script setup lang="ts">
const placeholder = 'Here is an example of translation that would be too complicated to implement with `formatMultiPartSpan` or by hard-coding spans in the template:\n\n' +
                    '- As we like both, the [chocolate](chocolink) that you bring _*may*_ be dark or with milk.\n' +
                    '- Da wir beide mögen, _*darf*_ die [Schokolade](chocolink), die du bringst, dunkel oder mit Milch sein.\n\n' +
                    'This component is called `MiniParser` because it has been programmed with the objective to minimize the overhead (it fits in 177 lines of code). *In the medium term, it will save more than 177 lines of code* across [the project](besturl), will save time (no layout to create for each text) and will avoid pain (when the layout depends on the language).\n\n' +
                    'For example, in _SubscriptionRow.vue_, we simply have one line:\n\n`✨ <BcMiniParser :input="tAll(t, tPath)" /> ✨`\n\n*instead of*\n\n_(shortening the following would keep it longer anyway, would make it unclear and invalid the day the text changes)_\n`💀 <div v-if="tPath.includes(\'offline_validator\')">`\n`💀   {{ tOf(t, tPath, 0) }}`\n`💀   <ul>`\n`💀     <li>{{ tOf(t, tPath, 1) }}</li>`\n`💀     <li>{{ tOf(t, tPath, 2) }}</li>`\n`💀     <li>{{ tOf(t, tPath, 3) }}</li>`\n`💀   </ul>`\n`💀 </div>`\n`💀 <div v-else-if="tPath.includes(\'offline_group\')">`\n`💀   {{ tOf(t, tPath, 0) }}`\n`💀   <ul>`\n`💀     <li>{{ tOf(t, tPath, 1) }}</li>`\n`💀     <li>{{ tOf(t, tPath, 2) }}</li>`\n`💀     <li>{{ tOf(t, tPath, 3) }}</li>`\n`💀     <li>{{ tOf(t, tPath, 4) }}</li>`\n`💀   </ul>`\n`💀   <b>{{ tOf(t, tPath, 5) }}</b> {{ tOf(t, tPath, 6) }}`\n`💀 </div>`\n`💀 <div v-else-if="tPath.includes(\'ignore_spam\')">`\n`💀   {{ tOf(t, tPath, 0) }}`\n`💀   <b>{{ tOf(t, tPath, 1) }}</b>`\n`💀   {{ tOf(t, tPath, 2) }}`\n`💀   <b>{{ tOf(t, tPath, 3) }}</b>`\n`💀 </div>`\n`💀 <div v-else>`\n`💀   {{ tOf(t, tPath, 0) }}`\n`💀 </div>`\n\n' +
                    'There are *several places* in the codebase where long hard-coded layouts like that one (potentially *invalid in other languages*) can be replaced with `<MiniParser/>` (shorter, simpler and language-safe).\n\n' +
                    '# Additional features\n' +
                    '- You can also give urls [directly](http://bitfly.at).\n- You can escape the tags \\*.\n- You can pass numbers that the parser will format, for example $myYearlySalary \\*.\n- ...'

const exampleOfLinks = {
  chocolink: 'https://en.wikipedia.org/wiki/Chocolate',
  besturl: '/dashboard'
}

const exampleOfNumbers = {
  myYearlySalary: 1420999
}

const input = ref<string>(placeholder)
</script>

<template>
  <div class="test-area">
    <textarea v-model="input" class="input" autocorrect="off" spellcheck="false" />
    <div class="magic">
      ➡️
    </div>
    <BcMiniParser :input="input" :links="exampleOfLinks" :numbers="exampleOfNumbers" class="output" />
  </div>
</template>

<style scoped lang="scss">
.test-area {
  display: flex;
  margin: 16px;
  .input {
    width: 47%;
    height: 550px;
    font-size: 16px;
  }
  .magic {
    text-align: center;
    padding-top: 150px;
    font-size: 30px;
    margin-left: auto;
    margin-right: auto;
  }
  .output {
    width: 47%;
    border: 1px solid grey;
    padding: 4px;
    font-size: 16px;
  }
}
</style>
