<script setup lang="ts">
import { Category, ResultType, type Matching } from '~/types/searchbar'
import { ChainIDs } from '~/types/networks'

// TODO : implementing this component!

// picks a result by default when the user presses Enter instead of clicking a result in the drop-down
function pickSomethingByDefault (possibilities : Matching[]) : Matching {
  // BcSearchbarMainComponent.vue has sorted the possible results in `possibilities` by network and type priority (the order appearing in the drop-down).
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
  <BcSearchbarMainComponent
    :searchable="[Category.Validators]"
    bar-style="embedded"
    :pick-by-default="pickSomethingByDefault"
    @go="userSelectedAValidator"
  />
</template>
