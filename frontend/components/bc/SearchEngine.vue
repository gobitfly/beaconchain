<script setup lang="ts">
import { warn } from 'vue'
import { SearchTypes } from '~/types/searchtypes'

const props = defineProps({ searchable: { type: Array, required: true } })
const searchable = props.searchable as SearchTypes[]
const emit = defineEmits(['enter', 'select'])

const inputField = ref<any>(null)
const proposal1 = ref<any>(null)

function userPressedEnter (input: string) {
  emit('enter', input, SearchTypes.addresses)
}

function userClickedProposal (selection: string) {
  emit('select', selection, SearchTypes.validators)
}

onMounted(() => {
  for (let i = 0; i < searchable.length; i++) {
    warn(searchable[i].toString())
  }
})
</script>

<template>
  <div>
    <input
      ref="inputField"
      type="text"
      @keydown.enter="userPressedEnter(inputField.value)"
    >
    <div
      ref="proposal1"
      @click="userClickedProposal(proposal1.innerText)"
    >
      CLICK ME
    </div>
  </div>
</template>
