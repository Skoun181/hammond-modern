<script>
import Layout from '@layouts/main.vue'
import axios from 'axios'

export default {
  page: { title: 'Hammond-Daten übernehmen' },
  components: { Layout },
  data() {
    return { file: null, preview: null, loading: false, importing: false, error: '', done: null, confirmed: false }
  },
  computed: {
    canImport() { return this.preview && this.preview.vehicles && this.preview.vehicles.length && this.confirmed && !this.importing }
  },
  methods: {
    onFileChange(event) {
      const files = event.target.files
      this.file = files && files.length ? files[0] : null
      this.preview = null
      this.done = null
      this.confirmed = false
      this.error = ''
    },
    formData(execute) {
      const form = new FormData()
      form.append('file', this.file, this.file.name)
      if (execute) form.append('confirm', 'IMPORTIEREN')
      return form
    },
    showError(error) {
      this.error = (error.response && error.response.data && error.response.data.error) ||
        error.message || 'Import nicht möglich.'
    },
    async inspect() {
      if (!this.file) { this.error = 'Bitte zuerst eine hammond.db auswählen.'; return }
      if (this.file.size > 50 * 1024 * 1024) { this.error = 'Die Datei ist größer als 50 MB.'; return }
      this.error = ''
      this.preview = null
      this.loading = true
      try {
        const result = await axios.post('/api/import/hammond/preview', this.formData(false), { timeout: 120000 })
        this.preview = result.data
        this.confirmed = false
      } catch (error) { this.showError(error) }
      finally { this.loading = false }
    },
    async runImport() {
      if (!this.canImport) return
      this.error = ''
      this.importing = true
      try {
        const result = await axios.post('/api/import/hammond/execute', this.formData(true), { timeout: 180000 })
        this.done = result.data
        this.preview = null
        this.confirmed = false
      } catch (error) { this.showError(error) }
      finally { this.importing = false }
    },
  },
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading">
      <div>
        <p class="hm-eyebrow">DATENÜBERNAHME</p>
        <h1 class="title hm-page-title">Vom Original-Hammond wechseln</h1>
        <p class="hm-page-description">Fahrzeuge, Tankungen und Ausgaben aus einer Hammond-Sicherung importieren.</p>
      </div>
    </div>

    <div class="box hm-import-panel">
      <h2 class="title is-5">1. Original-Datenbank auswählen</h2>
      <p>Erstelle eine Sicherung deiner Original-Hammond-Installation und wähle die Datei <strong>hammond.db</strong> aus dem bisherigen <code>/config</code>-Verzeichnis.</p>
      <p class="help">Tipp: Original-Hammond für die Sicherung kurz anhalten, damit die SQLite-Datei konsistent ist. Den laufenden Hammond-Modern-Datenordner niemals überschreiben.</p>
      <div class="file has-name is-fullwidth">
        <label class="file-label">
          <input class="file-input" type="file" accept=".db,.sqlite,.sqlite3,application/vnd.sqlite3,application/octet-stream" @change="onFileChange">
          <span class="file-cta"><span class="file-label">Datenbank auswählen</span></span>
          <span class="file-name">{{ file ? file.name : 'Keine Datei ausgewählt' }}</span>
        </label>
      </div>
      <b-button type="is-primary" :loading="loading" :disabled="!file || loading || importing" class="hm-import-action" @click="inspect">Sicherung prüfen</b-button>
    </div>

    <b-notification v-if="error" type="is-danger" :closable="false">{{ error }}</b-notification>
    <b-notification v-if="done" type="is-success" :closable="false">
      <strong>Import abgeschlossen!</strong>
      <p>Neu übernommen: {{ done.vehicles }} Fahrzeuge, {{ done.fillups }} Tankungen und {{ done.expenses }} Ausgaben.</p>
      <p>Bereits importiert und übersprungen: {{ done.skippedFillups }} Tankungen, {{ done.skippedExpenses }} Ausgaben.</p>
      <router-link to="/">Zum Dashboard</router-link>
    </b-notification>

    <div v-if="preview" class="box hm-import-panel">
      <h2 class="title is-5">2. Vorschau prüfen</h2>
      <p>Gefunden: <strong>{{ preview.vehicles.length }} Fahrzeuge, {{ preview.totalFillups }} Tankungen und {{ preview.totalExpenses }} Ausgaben.</strong></p>
      <div v-for="vehicle in preview.vehicles" :key="vehicle.id" class="hm-maintenance-row">
        <div class="hm-maintenance-row-text">
          <strong>{{ vehicle.name || '(Ohne Namen)' }}</strong>
          <p class="help">{{ vehicle.registration || 'Ohne Kennzeichen' }} · {{ vehicle.fillups }} Tankungen · {{ vehicle.expenses }} Ausgaben</p>
        </div>
        <span class="hm-maintenance-status">{{ vehicle.matched ? 'Bestehend, wird zugeordnet' : 'Wird neu angelegt' }}</span>
      </div>
      <b-notification type="is-warning" :closable="false">
        <strong>Was nicht übernommen wird:</strong> Benutzerkonten, Passwörter, Fahrzeugfreigaben, Anhänge/Rechnungsdateien und alte Erinnerungen. Importierte Einträge werden deinem aktuellen Konto zugeordnet. Vorher bitte auch Hammond Modern sichern.
      </b-notification>
      <label class="checkbox hm-import-confirm">
        <input v-model="confirmed" type="checkbox">
        Ich habe meine Hammond-Modern-Daten gesichert und möchte die angezeigten Daten importieren.
      </label>
      <div class="hm-import-actions">
        <b-button type="is-primary" :disabled="!canImport" :loading="importing" @click="runImport">3. Import jetzt durchführen</b-button>
      </div>
    </div>
  </Layout>
</template>
