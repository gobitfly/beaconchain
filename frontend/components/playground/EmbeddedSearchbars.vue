<script setup lang="ts">
import { type SearchBar, SearchbarStyle, SearchbarPurpose, ResultType, type ResultSuggestion, pickHighestPriorityAmongBestMatchings } from '~/types/searchbar'
import { ChainIDs } from '~/types/networks'

const selectedAccount = ref<string>('')
const selectedValidator = ref<string>('')
const accountSearchBar = ref<SearchBar>()
const validatorSearchBar = ref<SearchBar>()

function userSelectedAnAccount (result : ResultSuggestion) {
  switch (result.type) {
    case ResultType.Accounts :
    case ResultType.Contracts :
    case ResultType.EnsAddresses :
    case ResultType.EnsOverview :
      break
    default :
      return
  }
  selectedAccount.value = 'You selected ' + result.queryParam + ' on chain ' + result.chainId + '.'
  accountSearchBar.value!.hideResult(result)
}

function userSelectedValidator (result : ResultSuggestion) {
  switch (result.type) {
    case ResultType.ValidatorsByIndex :
    case ResultType.ValidatorsByPubkey :
      break
    case ResultType.ValidatorsByDepositAddress :
    case ResultType.ValidatorsByDepositEnsName :
    case ResultType.ValidatorsByWithdrawalCredential :
    case ResultType.ValidatorsByWithdrawalAddress :
    case ResultType.ValidatorsByWithdrawalEnsName :
    case ResultType.ValidatorsByGraffiti :
      break
    default :
      return
  }
  selectedValidator.value = 'You selected ' + result.queryParam + ' on chain ' + result.chainId + '. Number of validators: ' + result.count
  validatorSearchBar.value!.hideResult(result)
  validatorSearchBar.value!.closeDropdown()
}
</script>

<template>
  <br>
  This tab is used to design and implement the bars embedded in the Account and Validator pages
  <br> <br>
  <div class="example container">
    Accounts: <span>{{ selectedAccount }}</span>
    <div class="bar-container">
      <BcSearchbarMain
        ref="accountSearchBar"
        :bar-style="SearchbarStyle.Embedded"
        :bar-purpose="SearchbarPurpose.AccountAddition"
        :pick-by-default="pickHighestPriorityAmongBestMatchings"
        :keep-dropdown-open="true"
        @go="userSelectedAnAccount"
      />
    </div>
  </div>
  <br>
  <div class="example container">
    Validators: <span>{{ selectedValidator }}</span>
    <div class="bar-container">
      <BcSearchbarMain
        ref="validatorSearchBar"
        :bar-style="SearchbarStyle.Embedded"
        :bar-purpose="SearchbarPurpose.ValidatorAddition"
        :only-networks="[ChainIDs.Ethereum, ChainIDs.Gnosis]"
        :pick-by-default="pickHighestPriorityAmongBestMatchings"
        :keep-dropdown-open="true"
        @go="userSelectedValidator"
      />
    </div>
  </div>
</template>

<style scoped lang="scss">
.example {
  position: relative;
  max-width: 500px;
  height: 200px;
  padding: 16px;

  .bar-container {
    position: relative;
  }
}
</style>
