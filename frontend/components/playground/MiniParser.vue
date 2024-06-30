<script setup lang="ts">
const placeholder = 'Here is an example of translation that would be a pain to implement by hard-coding spans in the template or with `formatMultiPartSpan`:\n\n' +
                    '- The chocolate *can* be very dark or with milk because we _like_ [both](chocolink).\n' +
                    '- Da wir [beide](chocolink) _mögen_, *kann* die Schokolade sehr dunkel or mit Milch sein.\n\n' +
                    'The best website [is _this one_](besturl).\n\n' +
                    'It is called `MiniParser` because it fits _in *200* lines of code_.\n' +
                    'Additional features: you can also give urls [directly](http://bitfly.at), you can escape tags \\*, ...'

const exampleOfLinks = {
  chocolink: 'https://en.wikipedia.org/wiki/Chocolate',
  besturl: '/register'
}

const input = ref<string>(placeholder)
</script>

<template>
  <div class="test-area">
    A bug bounty with a prize of 2 chocolate bars is open for anyone willing to torture the parser:<br><br>
    <textarea v-model="input" class="input" autocorrect="off" spellcheck="false" />
    <div class="magic">
      🪄
    </div>
    <BcMiniParser :input="input" :links="exampleOfLinks" class="output" />
  </div>
</template>

<style scoped lang="scss">
.test-area {
  width: 620px;
  height: 360;
  margin: 16px;
  margin-left: auto;
  margin-right: auto;
  .input {
    width: 100%;
    height: 350px;
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
