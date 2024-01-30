<script setup lang="ts">
const year = new Date().getFullYear()
const version = '2'
// TODO: query the real version number
// A possibility: https://stackoverflow.com/questions/53962583/how-to-add-git-hash-to-vue-js-component
// from https://discord.com/channels/808317185873018881/1184805267196624966/1201533326503202866
</script>

<template>
  <span>© Bitfly GmbH {{ year }} | v{{ version }}</span>
</template>
