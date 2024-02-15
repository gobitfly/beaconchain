<template>
  <BcAdControl />
  <div class="ad_test_container">
    No blue box should be left here
    <div id="replace_me" class="ad_test">
      Ok come on and replace me
    </div>
  </div>
  <div class="ad_test_container">
    Ad should be within the box
    <div id="inside_me" class="ad_test">
      Should be iniside
    </div>
  </div>
  <div class="ad_test_container">
    Ad should be after this text, but before the box
    <div id="before_me" class="ad_test">
      Should come before me
    </div>
  </div>
  <div class="ad_test_container">
    <div id="after_me" class="ad_test">
      Should come after me
    </div>
    Ad should be before this text, but after the box
  </div>
  <div class="ad_test_container">
    Ad should be after this text, but before the box
    <div id="around_me" class="ad_test">
      Should come around me
    </div>
    Ad should be before this text, but after the box
  </div>
</template>

<style lang="scss">
.ad_test_container{
  color: red;
  border: 1px solid red;
  background-color: aqua;
  padding: 10px;
}
.ad_test{
  color: pink;
  border: 1px solid pink;
  background-color: darkblue;
  padding: 10px;
}
</style>
