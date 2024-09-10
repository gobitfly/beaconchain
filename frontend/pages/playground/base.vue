<script lang="ts" setup> 

</script>
<template>
  <div class="p-8">
    <h1>Playground BaseComponents</h1>
    <main class="mt-6 ">
      <section class="grid gap-2">
        <h2>BaseButton</h2>
        <BaseButton>Primary</BaseButton>
        <BaseButton variant="secondary">Secondary</BaseButton>
        <BaseButton isDisabled> Disabled</BaseButton>
      </section>
    </main>
    <div  class="absolute bottom-0 w-full" >
      <BcFooterMainFooter></BcFooterMainFooter>  
    </div>
  </div>
</template>