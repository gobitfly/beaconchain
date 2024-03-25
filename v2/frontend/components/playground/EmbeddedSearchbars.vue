<script setup lang="ts">
import { Category, ResultType, type Matching } from '~/types/searchbar'
import { ChainIDs } from '~/types/networks'

// TODO : implementing these examples of use

// Picks a result by default when the user presses Enter instead of clicking a result in the drop-down.
// If you return undefined, it means that either no result suits you or you want to deactivate Enter.
function pickSomethingByDefault (possibilities : Matching[]) : Matching|undefined {
  // BcSearchbarMain.vue has sorted the possible results in `possibilities` by network and type priority (the order appearing in the drop-down).
  // Now we look for the possibility that matches the best with the user input (this is known through the field `Matching.closeness`).
  // If several possibilities with this best closeness value exist, we catch the first one (so the one having the highest priority).
  let bestMatchWithHigherPriority = possibilities[0]
  for (const possibility of possibilities) {
    if (possibility.closeness < bestMatchWithHigherPriority.closeness) {
      bestMatchWithHigherPriority = possibility
    }
  }
  return bestMatchWithHigherPriority
}

function userSelectedAnAccount (wanted : string, type : ResultType, chain : ChainIDs) {
  switch (type) {
    case ResultType.Accounts :
    case ResultType.Contracts :
    case ResultType.EnsAddresses :
    case ResultType.EnsOverview :
      // to be implemented
      break
    default :
      return
  }

  return { wanted, chain } // just some dummy stuff to avoid the warning about unused parameters
}

function userSelectedValidator (wanted : string, type : ResultType, chain : ChainIDs, count : number) { // parameter `count` tells how many validators are in the batch (1 if no batch)
  switch (type) {
    case ResultType.ValidatorsByIndex :
    case ResultType.ValidatorsByPubkey :
    case ResultType.ValidatorsByDepositAddress :
    case ResultType.ValidatorsByDepositEnsName :
    case ResultType.ValidatorsByWithdrawalCredential :
    case ResultType.ValidatorsByWithdrawalAddress :
    case ResultType.ValidatorsByWithdrawalEnsName :
    case ResultType.ValidatorsByGraffiti :
      // to be implemented
      break
    default :
      return
  }

  return { wanted, chain, count } // just some dummy stuff to avoid the warning about unused parameters
}
</script>

<template>
  <br>
  This tab is used to design and implement the bars embedded in the Account and Validator pages
  <br> <br>
  <div class="container">
    Accounts:
    <br><br>
    <BcSearchbarMain
      :searchable="[Category.Addresses]"
      :unsearchable="[ResultType.EnsOverview]"
      bar-style="embedded"
      :pick-by-default="pickSomethingByDefault"
      @go="userSelectedAnAccount"
    />
  </div>
  <br>
  <div class="container">
    Validators:
    <br><br>
    <BcSearchbarMain
      :searchable="[Category.Validators]"
      :only-networks="[ChainIDs.Gnosis]"
      bar-style="embedded"
      :pick-by-default="pickSomethingByDefault"
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
