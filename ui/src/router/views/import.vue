<script>
import Layout from '@layouts/main.vue'

export default {
  page: {
    title: 'Import',
    meta: [{ name: 'Import from other apps', content: 'The Import page.' }],
  },
  components: { Layout },
  props: {
    user: {
      type: Object,
      required: true,
    },
  },
}
</script>

<template>
  <Layout>
    <div class="columns box"
      ><div class="column">
        <h1 class="title">Datenimport</h1>
        <p class="subtitle">Übernimm Daten aus deiner bisherigen Fahrzeugverwaltung.</p>
      </div></div
    >
    <br />
    <div class="box hm-import-panel">
      <h2 class="title is-5">Original-Hammond übernehmen</h2>
      <p>Einfach die alte <code>hammond.db</code> auswählen, Vorschau prüfen und Fahrzeuge, Tankungen sowie Ausgaben importieren.</p>
      <br />
      <b-button type="is-primary" tag="router-link" to="/import/hammond">Hammond-Migration starten</b-button>
    </div>
    <div class="columns">
      <div class="box column is-one-third" to="/import-fuelly">
        <h1 class="title is-5">Fuelly (CSV)</h1>
        <p>If you have been using Fuelly to store your vehicle data, export the CSV file from Fuelly and click here to import.</p>
        <br />
        <b-button type="is-primary" tag="router-link" to="/import/fuelly">Import</b-button>
      </div>
    </div>
  </Layout>
</template>
