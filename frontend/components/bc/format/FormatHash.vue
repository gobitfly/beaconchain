<script setup lang="ts">

interface Props {
  hash?: string,
  ens?: string,
  type?: 'address' | 'withdrawal_credentials' | 'public_key' | 'tx' | 'block_hash' | 'root' // if none is provided the default format will be applied
  full?: boolean, // if true the hash will not be truncated
  noLink?: boolean, // most of the time we want to render it as a link (if possible), but there might be cases where we don't
  noCopy?: boolean, // same as for the link
}
const props = defineProps<Props>()

const data = computed(() => {
  if (!props.hash) {
    return
  }
  const hash = props.hash
  const className = props.full ? 'full' : props.ens ? 'truncate-text' : 'parts'
  let parts: { value: string, className?: string }[] = []
  let link: string = ''
  if (props.ens) {
    parts.push({ value: props.ens, className: !props.full ? 'truncate-text' : '' })
  } else if (props.type === 'withdrawal_credentials') {
    const isSet = hash.startsWith('0x01')
    const color = isSet ? 'green' : 'orange'
    parts.push({ value: hash.substring(0, 4), className: color })
    if (props.full) {
      parts.push({ value: hash.substring(4) })
    } else {
      parts = parts.concat([{ value: hash.substring(26, 30), className: 'dots-before' }, { value: hash.substring(hash.length - 4), className: 'dots-before' }])
    }
    if (isSet && !props.noLink) {
      link = `/address/0x${props.hash.substring(26)}`
    }
  } else {
    const color = props.full ? 'prime' : undefined
    const middle = props.full ? { value: hash.substring(6, hash.length - 4) } : { value: '', className: 'dots-before' }
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

</script>
<template>
  <div v-if="data" class="format-hash">
    <BcTooltip class="tt-container">
      <template v-if="!full || ens" #tooltip>
        <div v-if="ens" class="tt ens-name full">
          {{ ens }}
        </div>
        <div class="tt">
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
    <BcCopyToClipboard v-if="!props.noCopy" :value="props.hash" class="copy" />
  </div>
</template>

<style lang="scss" scoped>
.format-hash {
  &:has(.truncate-text) {
    display: flex;
  }

  .tt-container {
    &:has(.truncate-text) {
      display: flex;
      overflow: hidden;
    }
  }

  .prime {
    color: var(--primary-color);
  }

  .green {
    color: var(--positive-color);
  }

  .orange {
    color: var(--orange-color);
  }

  .copy {
    margin-left: var(--padding);
    line-height: 100%;
  }

}

.full {
  word-wrap: break-word;
}

.tt {
  max-width: 300px;
  text-align: left;

  :deep(.format-hash) {
    .green {
      color: var(--light-green);
    }

    .orange {
      color: var(--light-orange);
    }
  }
}

.ens-name {
  margin-bottom: var(--padding-small);
}
</style>
