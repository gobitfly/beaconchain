<script setup lang="ts">
import { SearchbarStyle, SearchbarPurpose, ResultType, pickHighestPriorityAmongMostRelevantMatchings } from '~/types/searchbar'
import { ChainIDs } from '~/types/networks'

const selectedAccount = ref<string>('')
const selectedValidator = ref<string>('')

function userSelectedAnAccount (wanted : string, type : ResultType, chain : ChainIDs) {
  switch (type) {
    case ResultType.Accounts :
    case ResultType.Contracts :
    case ResultType.EnsAddresses :
    case ResultType.EnsOverview :
      break
    default :
      return
  }

  selectedAccount.value = 'You selected ' + wanted + ' on chain ' + chain + '.'
}

function userSelectedValidator (wanted : string, type : ResultType, chain : ChainIDs, count : number) { // parameter `count` tells how many validators are in the batch (1 if no batch)
  switch (type) {
    case ResultType.ValidatorsByIndex : // `wanted` contains the index of the validator
    case ResultType.ValidatorsByPubkey : // `wanted` contains the pubkey of the validator
      break
    // The following types can correspond to several validators. The search bar doesn't know the list of indices and pubkeys :
    case ResultType.ValidatorsByDepositAddress : // `wanted` contains the address that was used to deposit the 32 ETH
    case ResultType.ValidatorsByDepositEnsName : // `wanted` contains the ENS name that was used to deposit the 32 ETH
    case ResultType.ValidatorsByWithdrawalCredential : // `wanted` contains the withdrawal credential
    case ResultType.ValidatorsByWithdrawalAddress : // `wanted` contains the withdrawal address
    case ResultType.ValidatorsByWithdrawalEnsName : // `wanted` contains the ENS name of the withdrawal address
    case ResultType.ValidatorsByGraffiti : // `wanted` contains the graffiti used to sign blocks
      break
    default :
      return
  }

  selectedValidator.value = 'You selected ' + wanted + ' on chain ' + chain + '. Number of validators: ' + count
}
</script>

<template>
  <br>
  This tab is used to design and implement the bars embedded in the Account and Validator pages
  <br> <br>
  <div class="container">
    Accounts: <span>{{ selectedAccount }}</span>
    <br><br>
    <BcSearchbarMain
      :bar-style="SearchbarStyle.Embedded"
      :bar-purpose="SearchbarPurpose.Accounts"
      :pick-by-default="pickHighestPriorityAmongMostRelevantMatchings"
      @go="userSelectedAnAccount"
    />
  </div>
  <br>
  <div class="container">
    Validators: <span>{{ selectedValidator }}</span>
    <br><br>
    <BcSearchbarMain
      :bar-style="SearchbarStyle.Embedded"
      :bar-purpose="SearchbarPurpose.Validators"
      :only-networks="[ChainIDs.Ethereum, ChainIDs.Gnosis]"
      :pick-by-default="pickHighestPriorityAmongMostRelevantMatchings"
      @go="userSelectedValidator"
    />
  </div>
</template>

<style scoped lang="scss">
.container {
  max-width: 500px;
  height: 200px;
  padding: 16px;
}
</style>
