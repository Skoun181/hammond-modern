<script>
import Layout from '@layouts/main.vue'

export default {
  page: {
    title: 'Loading page...',
    meta: [{ name: 'description', content: 'Loading page...' }],
  },
  components: { Layout },
}
</script>

<template>
  <Layout>
    <Transition appear>
      <BaseIcon :class="$style.loadingIcon" name="sync" spin />
    </Transition>
  </Layout>
</template>

<style module>

.loadingIcon {
  font-size: 2rem;

  display: block;
  margin: 0 auto;

  &:global(.v-enter-active) {
    transition: opacity 1s;
  }
  &:global(.v-enter-from) {
    opacity: 0;
  }
}
</style>
