<template>
  <BcPageWrapper :is-home-page="true">
    <h1>Index page</h1>
    <div class="searchbar">
      <BcSearchbarGeneral bar-style="gaudy" />
    </div>
    <br><br><br>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
.searchbar {
  display: flex;
  justify-content: center;
}
</style>
