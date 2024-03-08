<script setup lang="ts">import {
  faCopy
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { warn } from 'vue'

interface Props {
  hash?: string,
  ens?: string,
  type?: 'address' | 'withdrawal_address' | 'public_key' | 'tx' | 'block_hash' | 'root'
  full?: boolean,
  noLink?: boolean, // most of the time we want to render it as a link (if possible), but there might be cases where we don't
  noCopy?: boolean, // same as for the link
}
const props = defineProps<Props>()

const dots = { value: '...' }

const data = computed(() => {
  if (!props.hash) {
    return
  }
  const hash = props.hash
  const className = props.ens ? 'truncate-text' : props.full ? 'full' : 'parts'
  let parts: { value: string, className?: string }[] = [] // [{ value: props.ens ? props.ens : '0x' }]
  let link: string = ''
  if (props.ens) {
    parts.push({ value: props.ens })
  } else if (props.type === 'withdrawal_address') {
    const isSet = hash.startsWith('0x01')
    const color = isSet ? 'green' : 'orange'
    parts.push({ value: hash.substring(0, 4), className: color })
    if (props.full) {
      parts.push({ value: hash.substring(4) })
    } else {
      parts = parts.concat([dots, { value: hash.substring(26, 30) }, dots, { value: hash.substring(hash.length - 4) }])
    }
    if (isSet && !props.noLink) {
      link = `/address/0x${props.hash.substring(26)}`
    }
  } else {
    const color = props.full ? 'prime' : undefined
    const middle = props.full ? { value: hash.substring(6, hash.length - 4) } : dots
    parts = [{ value: '0x' }, { value: hash.substring(2, 6), className: color }, middle, { value: hash.substring(hash.length - 4), className: color }]
  }
  if (!props.noLink) {
    switch (props.type) {
      case 'address':
        link = `/address/${props.hash}`
        break
      case 'block_hash':
        link = `/block/${props.hash}`
        break
      case 'root':
        link = `/slot/${props.hash}`
        break
      case 'public_key':
        link = `/validator/${props.hash}`
        break
    }
  }

  return {
    parts,
    link,
    className
  }
})

function copyToClipboard (): void {
  if (!props.hash) {
    return
  }

  navigator.clipboard.writeText(props.hash)
    .catch((error) => {
      warn('Error copying text to clipboard:', error)
    })
}

</script>
<template>
  <BcTooltip v-if="data">
    <template v-if="!full || ens" #tooltip>
      <div v-if="ens">
        {{ ens }}
      </div>
      <div class="tt-hash">
        <BcFormatHash :hash="hash" :full="true" :no-link="true" :no-copy="true" :type="type" />
      </div>
    </template>
    <span v-if="!data.link" :class="data.className">
      <span v-for="(part, index) in data.parts" :key="index" :class="part.className">
        {{ part.value }}
      </span>
    </span>
    <NuxtLink v-else :to="data.link" class="link" :class="data.className">
      <span v-for="(part, index) in data.parts" :key="index" :class="part.className">
        {{ part.value }}
      </span>
    </NuxtLink>
  </BcTooltip>
  <FontAwesomeIcon v-if="!props.noCopy" :icon="faCopy" class="copy" @click="copyToClipboard" />
</template>

<style lang="scss" scoped>
.tt-hash {
  max-width: 300px;
}

.full {
  word-wrap: break-word;
}

.prime {
  color: var(--primary-color);
}

.green {
  color: var(--green);
}

.orange {
  color: var(--orange);
}

.copy {
  margin-left: var(--padding);
  cursor: pointer;
}
</style>
