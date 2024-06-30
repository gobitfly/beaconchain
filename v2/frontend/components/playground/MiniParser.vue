<script setup lang="ts">
const placeholder = 'Here is an example of translation that would be too complicated to implement with `formatMultiPartSpan` or by hard-coding spans in the template:\n\n' +
                    '- The [chocolate](chocolink) can be dark or with milk *because we like both*.\n' +
                    '- *Da wir beide mögen*, kann die [Schokolade](chocolink) dunkel oder mit Milch sein.\n\n' +
                    'This component is called `MiniParser` because it has been programmed with the objective to minimize the overhead (it fits in 237 lines of code). *In the medium term, it will save more than 237 lines of code* across [the project](besturl), will save time (no layout to create for each text) and will save from pain (when the layout depends on the language).\n\n' +
                    'For example, in _SubscriptionRow.vue_, we simply have one line:\n\n`✨ <BcMiniParser :input="tAll(t, tPath)" class="tt-content" />` ✨\n\n*instead of*\n\n_(shortening the following would keep it longer anyway, would make it unclear and invalid the day the text changes)_\n`💀 <div v-if="tPath.includes(\'offline_validator\')" class="tt-content">`\n`💀   {{ tOf(t, tPath, 0) }}`\n`💀   <ul>`\n`💀     <li>{{ tOf(t, tPath, 1) }}</li>`\n`💀     <li>{{ tOf(t, tPath, 2) }}</li>`\n`💀     <li>{{ tOf(t, tPath, 3) }}</li>`\n`💀   </ul>`\n`💀 </div>`\n`💀 <div v-else-if="tPath.includes(\'offline_group\')" class="tt-content">`\n`💀   {{ tOf(t, tPath, 0) }}`\n`💀   <ul>`\n`💀     <li>{{ tOf(t, tPath, 1) }}</li>`\n`💀     <li>{{ tOf(t, tPath, 2) }}</li>`\n`💀     <li>{{ tOf(t, tPath, 3) }}</li>`\n`💀     <li>{{ tOf(t, tPath, 4) }}</li>`\n`💀   </ul>`\n`💀   <b>{{ tOf(t, tPath, 5) }}</b> {{ tOf(t, tPath, 6) }}`\n`💀 </div>`\n`💀 <div v-else-if="tPath.includes(\'ignore_spam\')" class="tt-content">`\n`💀   {{ tOf(t, tPath, 0) }}`\n`💀   <b>{{ tOf(t, tPath, 1) }}</b>`\n`💀   {{ tOf(t, tPath, 2) }}`\n`💀   <b>{{ tOf(t, tPath, 3) }}</b>`\n`💀 </div>`\n`💀 <div v-else class="tt-content">`\n`💀   {{ tOf(t, tPath, 0) }}`\n`💀 </div>`\n\n' +
                    'There are *several places* in the codebase where long hard-coded layouts like that one (potentially *invalid in other languages*) can be replaced with `<MiniParser/>` (shorter, simpler and language-safe).\n\n' +
                    '# Additional features\n' +
                    '- You can also give urls [directly](http://bitfly.at).\n- You can escape the tags \\*.\n- ...'

const exampleOfLinks = {
  chocolink: 'https://en.wikipedia.org/wiki/Chocolate',
  besturl: '/dashboard'
}

const input = ref<string>(placeholder)
</script>

<template>
  <div class="test-area">
    A bug bounty with a prize of 2 chocolate bars is open for anyone torturing the parser:<br><br>
    <textarea v-model="input" class="input" autocorrect="off" spellcheck="false" />
    <div class="magic">
      🪄
    </div>
    <BcMiniParser :input="input" :links="exampleOfLinks" class="output" />
  </div>
</template>

<style scoped lang="scss">
.test-area {
  width: 720px;
  margin: 16px;
  margin-left: auto;
  margin-right: auto;
  .input {
    width: 100%;
    height: 330px;
    font-size: 16px;
  }
  .magic {
    height:24px;
    text-align: center;
    padding-top: 6px;
    font-size: 24px;
  }
  .output {
    margin-top: 20px;
    width: 100%;
    border: 1px solid grey;
    padding: 4px;
    font-size: 16px;
  }
}
</style>
