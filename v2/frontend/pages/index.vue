<script setup lang="ts">
import { SearchbarShape, SearchbarColors } from '~/types/searchbar'

const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

useBcSeo()
</script>

<template>
  <BcPageWrapper :is-home-page="true">
    <h1>Index page</h1>
    <div class="search-bar">
      <BcSearchbarGeneral
        v-if="showInDevelopment"
        :bar-shape="SearchbarShape.Big"
        :color-theme="SearchbarColors.Default"
        :screen-width-causing-sudden-change="
          0
        /*if you introduce a media query (or similar) changing the width of the
        bar, give the threshold here to avoid visual bugs in the list of results */
        "
        class="bar"
      />
    </div>
    <br><br><br>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
.search-bar {
  display: flex;
  justify-content: center;
  .bar {
    width: 730px;
  }
}
</style>
