<script setup lang="ts">
const route = useRoute()
// const id = computed(() => route.params.id as string)
const id = ref(Number(route.params.id as string))
// const id = computed(() => undefined)
const url = computed(() => `/todos/${id.value}`)
const {
  data,
  error,
} = useFetch(url, {
// } = useFetch(`/todos/${id.value}`, {
  baseURL: 'https://jsonplaceholder.typicode.com',
  // immediate: !!id.value,
  // key: 'test',
  // watch: [ id ],
})
const onClick = () => {
  // This is just a test to see if the page reloads
  // It should not, because the id is undefined
  // route.params.id = 'new-id'
  id.value++
  navigateTo({
    name: 'test-id', params: { id: id.value },
  })
}
</script>

<template>
  <div>
    {{ data }}
    {{ error }}
    <div>
      <button @click="onClick">
        click
      </button>
    </div>
  </div>
</template>
