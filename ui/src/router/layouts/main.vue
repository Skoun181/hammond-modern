<script>
import NavBar from '@components/nav-bar.vue'
export default { components: { NavBar } }
</script>

<template>
  <div class="hm-shell">
    <NavBar />
    <main id="main-content" class="hm-main"><div class="hm-main-inner"><slot /></div></main>
  </div>
</template>
