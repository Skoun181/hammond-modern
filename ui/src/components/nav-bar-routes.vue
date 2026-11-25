<script>
export default {
  props: {
    routes: { type: Array, required: true },
  },
  methods: {
    title(route) {
      return typeof route.title === 'function' ? route.title() : route.title
    },
    badge(route) {
      return typeof route.badge === 'function' ? route.badge() : route.badge
    },
  },
}
</script>

<template>
  <BaseLink v-for="route in routes" :key="route.name" :to="route">
    {{ title(route) }}
    <b-tag v-if="badge(route) > 0" rounded type="is-danger is-light">
      {{ badge(route) }}
    </b-tag>
  </BaseLink>
</template>
