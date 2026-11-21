<script>
import Layout from '@layouts/main.vue'
export default {
  page: { title: 'Fahrzeuge' },
  components: { Layout },
  data() { return { vehicles: [], loading: true, error: '', search: '' } },
  computed: {
    visibleVehicles() {
      const q = this.search.trim().toLowerCase()
      return this.vehicles.filter(v => [v.nickname, v.registration, v.make, v.model].join(' ').toLowerCase().includes(q))
    },
  },
  mounted() { this.load() },
  methods: {
    async load() {
      this.loading = true; this.error = ''
      try { this.vehicles = (await this.$store.dispatch('vehicles/fetchVehicles')) || [] }
      catch (_) { this.error = 'Fahrzeuge konnten nicht geladen werden.' }
      finally { this.loading = false }
    },
  },
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading">
      <div><p class="hm-eyebrow">VERWALTUNG</p><h1 class="title hm-page-title">Fahrzeuge</h1><p class="hm-page-description">Eigene und freigegebene Fahrzeuge auf einen Blick.</p></div>
      <b-button type="is-primary" tag="router-link" :to="{ name: 'vehicle-create' }">＋ Fahrzeug hinzufügen</b-button>
    </div>
    <b-notification v-if="error" type="is-danger" :closable="false">{{ error }} <b-button size="is-small" @click="load">Erneut laden</b-button></b-notification>
    <p v-if="loading" role="status">Fahrzeuge werden geladen …</p>
    <template v-else-if="!error">
      <b-field label="Fahrzeug suchen"><b-input v-model="search" placeholder="Name, Kennzeichen, Marke oder Modell" /></b-field>
      <p v-if="!visibleVehicles.length" class="box">{{ vehicles.length ? 'Keine passenden Fahrzeuge.' : 'Noch keine Fahrzeuge angelegt. Klicke auf Fahrzeug hinzufügen.' }}</p>
      <div class="columns is-multiline">
        <div v-for="v in visibleVehicles" :key="v.id" class="column is-half">
          <div class="box">
            <p class="hm-eyebrow">{{ v.isOwner ? 'EIGENES FAHRZEUG' : 'FREIGEGEBEN' }}</p>
            <h2 class="title is-5">{{ v.nickname }} · {{ v.registration }}</h2>
            <p>{{ [v.make, v.model, v.yearOfManufacture].filter(Boolean).join(' · ') || 'Keine weiteren Stammdaten' }}</p>
            <div class="buttons" style="margin-top:1rem">
              <b-button tag="router-link" :to="{ name: 'vehicle-detail', params: { vehicleId: v.id } }">Details</b-button>
              <b-button v-if="v.isOwner" type="is-primary" tag="router-link" :to="{ name: 'vehicle-edit', params: { vehicleId: v.id } }">Bearbeiten</b-button>
              <b-button tag="router-link" :to="{ name: 'tires', query: { vehicleId: v.id } }">Reifen</b-button>
            </div>
          </div>
        </div>
      </div>
    </template>
  </Layout>
</template>
