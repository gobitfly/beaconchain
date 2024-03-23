<script setup lang="ts">
import { Category, ResultType, type Matching } from '~/types/searchbar'
import { ChainIDs } from '~/types/networks'

// TODO : implementing this toy component

// picks a result by default when the user presses Enter instead of clicking a result in the drop-down
function pickSomethingByDefault (possibilities : Matching[]) : Matching {
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

function userSelectedAValidator (wanted : string, type : ResultType, chain : ChainIDs) {
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

  return { wanted, chain } // just some dummy stuff to avoid the warning about unused parameters
}
</script>

<template>
  <br>
  Not implemented yet. This tab will be used to design and implement the bars embedded in the Account and Validator pages
  <br> <br>
  <div class="container">
    Accounts:
    <br>
    <BcSearchbarMain
      :searchable="[Category.Addresses]"
      bar-style="embedded"
      :pick-by-default="pickSomethingByDefault"
      @go="userSelectedAnAccount"
    />
  </div>
  <br>
  <div class="container">
    Validators:
    <br>
    <BcSearchbarMain
      :searchable="[Category.Validators]"
      bar-style="embedded"
      :pick-by-default="pickSomethingByDefault"
      @go="userSelectedAValidator"
    />
  </div>
</template>

<style scoped lang="scss">
.container {
  width: 400px;
  height: 200px;
}
</style>
