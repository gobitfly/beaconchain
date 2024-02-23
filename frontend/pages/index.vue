<template>
  <BcPageWrapper :is-home-page="true">
    Index page
    <div id="searchbar">
      <BcMainSearchBar location="page" />
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
#searchbar {
  display: flex;
  justify-content: center;
}
</style>
