<script setup lang="ts">

interface Props {
  validators: number[],
  groupId?: number,
  context: 'group' | 'sync' | 'propsoal'
}
const props = defineProps<Props>()

const openValidatorModal = () => {
  // TODO: replace with real modal
  // TODO: pass title and subtitle based on groupId (mapped with with group name, from dashboard overview store) and context
  alert(`${props.validators?.join(', ')} - ${props.groupId} - ${props.context}`)
}

</script>
<template>
  <div class="validator_column">
    <div class="validators">
      <NuxtLink v-for="v in props.validators" :key="v" :to="`/validator/${v}`" class="link validator_link">
        {{ v }}
      </NuxtLink>
    </div>
    <IconPopout v-if="validators?.length" class="link popout" @click="openValidatorModal" />
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';

.validator_column {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .validators {
    @include main.truncate-text;

    .validator_link:not(:last-child)::after {
      content: ", "
    }
  }

  .popout {
    width: 14px;
    height: auto;
    margin-left: var(--padding-small);
    flex-shrink: 0;
  }
}</style>
