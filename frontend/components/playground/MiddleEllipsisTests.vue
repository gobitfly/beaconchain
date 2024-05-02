<script setup lang="ts">
let randomInt = 9129
const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
const randomTexts = ref<string[]>([])

// custom rnd because we need to start with the same seed during SSR and at the first client rendering
function rand (max: number) : number {
  randomInt = (214013 * randomInt + 2531011) & 0xFFFFFFFF
  const normalized = ((randomInt >> 16) & 0xFFFF) / 65535
  return Math.round(normalized * max)
}

function getRandomText () : string {
  let result = ''
  for (let l = 20 + rand(100); l > 0; l--) {
    result += characters[rand(characters.length - 1)]
  }
  return result
}

function reGenerateTextList () {
  randomTexts.value.length = 0
  for (let i = 0; i < 100; i++) {
    randomTexts.value.push(getRandomText())
  }
}

reGenerateTextList()

const showAllME = ref<boolean>(false)
const showAllCSSclipped = ref<boolean>(false)
</script>

<template>
  <br>
  <Button @click="reGenerateTextList">
    Change texts
  </Button>
  <p>
    For this test:<br>
    * Blue means that the width is not defined (the component finds a width by using its content and its initial-flex-grow props).
    Other colors mean that the width is defined with flex-grow or width %<br>
    * A red frame highlights a MiddleEllipsis controlling several MiddleEllipses (this is needed to make sure they don't disturb each other,
    and allows for children of <i>undefined width</i>).
  </p>
  <div style="position: relative;">
    <p><b>With 1 ellipsis:</b></p>
    <BcSearchbarMiddleEllipsis class="frame medium standalone" :text="randomTexts[0]" />
    <BcSearchbarMiddleEllipsis class="frame big nocolor parent">
      <BcSearchbarMiddleEllipsis class="flexible loose" :text="randomTexts[1]" :initial-flex-grow="1" />
      <BcSearchbarMiddleEllipsis class="flexible medium" :text="randomTexts[2]" />
      <span>Hello I am not a MiddleEllipsis*</span>
      <BcSearchbarMiddleEllipsis class="flexible big" :text="randomTexts[3]" />
      <BcSearchbarMiddleEllipsis class="flexible loose" :text="randomTexts[4]" :initial-flex-grow="1" />
    </BcSearchbarMiddleEllipsis>
    <p>* you can put anything in a parent MiddleEllipsis, he will control its children and leave the rest as it is.</p>
  </div>

  <div style="position: relative;">
    <p><b>With 2 ellipses:</b></p>
    <p>
      Sometimes, you will notice that there is 1 ellipsis only. This is not a bug, it is because 2 ellipses would not make sense:
      <br>
      * If the text looks long when it happens: there is room for n characters and the text length is n+1, so there is only 1 character to skip, not 2.
      Adding an ellipsis would hide a character that is not clipped, this would be a loss of information.
      <br>
      * If the text looks short when it happens: with 2 ellipses, there would be 1 visible character ony, for example "…C…" or "A……" which is a loss of information without any advantage, therefore "A…D" is shown.
    </p>
    <BcSearchbarMiddleEllipsis class="frame medium standalone" :text="randomTexts[0]" :ellipses="2" />
    <BcSearchbarMiddleEllipsis class="frame big nocolor parent">
      <BcSearchbarMiddleEllipsis class="flexible loose" :text="randomTexts[1]" :ellipses="2" :initial-flex-grow="1" />
      <BcSearchbarMiddleEllipsis class="flexible medium" :text="randomTexts[2]" :ellipses="2" />
      <BcSearchbarMiddleEllipsis class="flexible big" :text="randomTexts[3]" :ellipses="2" />
      <BcSearchbarMiddleEllipsis class="flexible loose" :text="randomTexts[4]" :ellipses="2" :initial-flex-grow="1" />
    </BcSearchbarMiddleEllipsis>
  </div>

  <div style="position: relative;">
    <p><b>With an adaptive number of ellipses (configurable in the props):</b></p>
    <p>
      <i>Play with the width of your window to see the number of ellipses change.</i><br><br>
      Configuration for this test:<br>
      * 1 ellipsis if there is room for up to 16 characters,<br>
      * up to 2 ellipses if there is room for up to 32 characters,<br>
      * up to 3 ellipses if there is room for up to 64 characters,<br>
      * up to 4 ellipses if there is room for more than 64 characters.<br>
    </p>
    <BcSearchbarMiddleEllipsis class="frame medium standalone" :text="randomTexts[0]" :ellipses="[16,32,64]" />
    <BcSearchbarMiddleEllipsis class="frame big nocolor parent">
      <BcSearchbarMiddleEllipsis class="flexible loose" :text="randomTexts[2]" :ellipses="[16,32,64]" :initial-flex-grow="1" />
      <BcSearchbarMiddleEllipsis class="flexible medium" :text="randomTexts[1]" :ellipses="[16,32,64]" />
      <BcSearchbarMiddleEllipsis class="flexible big" :text="randomTexts[5]" :ellipses="[16,32,64]" />
      <BcSearchbarMiddleEllipsis class="flexible loose" :text="randomTexts[4]" :ellipses="[16,32,64]" :initial-flex-grow="1" />
    </BcSearchbarMiddleEllipsis>
  </div>

  <div style="position: relative;">
    <p><b>Redistribution of unused room</b></p>
    <p>Here we have 4 children with different flex-grow values. It can happen that some of them do not need to clip (their text is displayable entirely), which means that they have spare room. Observe that instead of leaving them with unused space, the parent distributes their spare room to children needing clipping, so those can diplay more:</p>
    (enlarge your window or change the texts to see it)
    <br>
    <BcSearchbarMiddleEllipsis class="frame all nocolor parent">
      <BcSearchbarMiddleEllipsis class="flexible medium" :text="randomTexts[0]" />
      <BcSearchbarMiddleEllipsis class="flexible big" :text="randomTexts[5]" />
      <BcSearchbarMiddleEllipsis class="flexible medium" :text="randomTexts[1]" />
      <BcSearchbarMiddleEllipsis class="flexible big" :text="randomTexts[3]" />
    </BcSearchbarMiddleEllipsis>
  </div>

  <div style="position: relative;">
    <p>
      <b>{{ randomTexts.length }} MiddleEllipses to see the lower smoothness of the UI when you resize your window:<br>
        (for performance, make sure to deactivate the debug mode in the source of the component!)</br>
        <br>
        <Button @click="showAllCSSclipped=false; showAllME=!showAllME">
          show/hide
        </Button>
      </b>
    </p>
    <div v-if="showAllME">
      <BcSearchbarMiddleEllipsis
        v-for="text of randomTexts"
        :key="text"
        class="percent medium"
        :text="text"
      />
    </div>
  </div>

  <div style="position: relative;">
    <p>
      Compare the smoothness with {{ randomTexts.length }} spans clipped natively by CSS:
      <br>
      <Button @click="showAllME=false; showAllCSSclipped=!showAllCSSclipped">
        show/hide
      </Button>
    </p>
    <div v-if="showAllCSSclipped">
      <div
        v-for="text of randomTexts"
        :key="text"
        class="percent medium css-clipping"
      >
        {{ text }}
      </div>
    </div>
  </div>

  <p>
    -> With a modern computer, 100 Middle Ellipses seem to be an upper limit to keep a satisfying smoothness (and 50 are super smooth, unnoticeable).
  </p>
</template>

<style scoped lang="scss">
.fixed {
  position: relative;
  display: inline-flex;
  &.medium {
    width: 200px;
    background-color: rgba(149, 218, 132, 0.63);
  }
  &.big {
    width: 400px;
    background-color: rgba(218, 132, 132, 0.603);
  }
}

.flexible {
  position: relative;
  display: inline-flex;
  &.medium {
    flex-grow: 1;
    background-color: rgba(149, 218, 132, 0.404);
  }
  &.big {
    flex-grow: 2;
    background-color: rgba(218, 132, 132, 0.404);
  }
  &.loose {
    background-color: rgba(87, 174, 255, 0.404);
  }
}

.percent {
  position: relative;
  display: inline-block;
  margin: 4px;
  &.small {
    width: 5%;
  }
  &.medium {
    width: 10%;
  }
  &.big {
    width: 20%;
  }
  &.all {
    width: 100%;
  }
}

.frame {
  position: relative;
  display: inline-flex;
  &.medium {
    width: 25%;
  }
  &.big {
    width: 65%;
  }
  &.all {
    width: 100%;
  }
  &.standalone {
    background-color: rgba(252, 204, 45, 0.397);
  }
}

.nocolor {
  &.medium,
  &.big,
  &.loose {
    background-color: transparent;
  }
}

.parent {
  border: 1px solid red;
}

.css-clipping {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
