<script>
import Layout from '@layouts/main.vue'
export default {
  page: { title: 'Datenimport' },
  components: { Layout },
  computed: {
    isAdmin() { return this.$store.state.auth.currentUser && this.$store.state.auth.currentUser.role === 'ADMIN' },
  },
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading hm-import-heading">
      <div>
        <p class="hm-eyebrow">DATENÜBERNAHME</p>
        <h1 class="title hm-page-title">Datenimport</h1>
        <p class="hm-page-description">Übernimm deine Fahrzeughistorie sicher in Hammond Modern.</p>
      </div>
    </div>
    <div class="hm-import-options">
      <div v-if="isAdmin" class="box hm-import-option">
        <p class="hm-eyebrow">SERVERUMZUG</p>
        <h2 class="title is-5">Hammond Modern Backup (.zip)</h2>
        <p>Erstelle eine vollständige Datensicherung oder stelle deine Daten aus einem ZIP-Backup auf einer neuen Installation wieder her.</p>
        <b-button type="is-primary" tag="router-link" to="/admin/backup">Backup &amp; Umzug öffnen</b-button>
      </div>
      <div class="box hm-import-option hm-import-option-featured">
        <p class="hm-eyebrow">EMPFOHLEN</p>
        <h2 class="title is-5">Original-Hammond</h2>
        <p>Importiere Fahrzeuge, Tankungen und Ausgaben aus einer Sicherung deiner bisherigen Hammond-Installation.</p>
        <p class="hm-import-hint">SQLite-Datenbank <code>hammond.db</code> · mit Vorschau und Duplikatprüfung</p>
        <b-button type="is-primary" tag="router-link" to="/import/hammond">Hammond-Migration starten</b-button>
      </div>
      <div class="box hm-import-option">
        <p class="hm-eyebrow">WEITERE QUELLE</p>
        <h2 class="title is-5">Fuelly (CSV)</h2>
        <p>Du hast Tankungen bei Fuelly gespeichert? Exportiere dort deine CSV-Datei und importiere sie hier.</p>
        <p class="hm-import-hint">CSV-Import · Fahrzeuge müssen bereits angelegt sein</p>
        <b-button type="is-primary" tag="router-link" to="/import/fuelly">CSV importieren</b-button>
      </div>
    </div>
  </Layout>
</template>
