<script>
import Layout from '@layouts/main.vue'
import store from '@state/store'

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
      this.error = error.message || 'Import nicht möglich.'
    },
    async sendUpload(endpoint, execute) {
      // The browser must generate the multipart boundary itself. Global Axios
      // defaults may force application/json and make Gin c.FormFile fail.
      const token = store.state.auth.currentUser && store.state.auth.currentUser.token
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: token ? { Authorization: token } : {},
        body: this.formData(execute),
        credentials: 'same-origin',
      })
      const payload = await response.json().catch(() => ({}))
      if (!response.ok) throw new Error(payload.error || 'Upload fehlgeschlagen (HTTP ' + response.status + ').')
      return payload
    },
    async inspect() {
      if (!this.file) { this.error = 'Bitte zuerst eine hammond.db auswählen.'; return }
      if (this.file.size > 50 * 1024 * 1024) { this.error = 'Die Datei ist größer als 50 MB.'; return }
      this.error = ''
      this.preview = null
      this.loading = true
      try {
        this.preview = await this.sendUpload('/api/import/hammond/preview', false)
        this.confirmed = false
      } catch (error) { this.showError(error) }
      finally { this.loading = false }
    },
    async runImport() {
      if (!this.canImport) return
      this.error = ''
      this.importing = true
      try {
        this.done = await this.sendUpload('/api/import/hammond/execute', true)
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
    <div class="hm-page-heading hm-import-heading">
      <div>
        <p class="hm-eyebrow">DATENÜBERNAHME</p>
        <h1 class="title hm-page-title">Vom Original-Hammond wechseln</h1>
        <p class="hm-page-description">Fahrzeuge, Tankungen und Ausgaben aus einer Hammond-Sicherung importieren.</p>
      </div>
    </div>

    <div class="box hm-import-panel hm-import-step">
      <h2 class="title is-5">1. Original-Datenbank auswählen</h2>
      <p>Wähle eine <strong>konsistente Sicherung</strong> deiner alten Hammond-Datenbank aus. Die Datei heißt normalerweise <code>hammond.db</code> und liegt im ursprünglichen <code>/config</code>-Ordner.</p>
      <p class="help">Tipp: Original-Hammond für die Sicherung kurz anhalten, damit die SQLite-Datei konsistent ist. Den laufenden Hammond-Modern-Datenordner niemals überschreiben.</p>
      <div class="file has-name is-fullwidth hm-import-picker">
        <label class="file-label">
          <input class="file-input" type="file" accept=".db,.sqlite,.sqlite3,application/vnd.sqlite3,application/octet-stream" @change="onFileChange">
          <span class="file-cta"><span class="file-label">Datenbank auswählen</span></span>
          <span class="file-name">{{ file ? file.name : 'Keine Datei ausgewählt' }}</span>
        </label>
      </div>
      <b-button type="is-primary" :loading="loading" :disabled="!file || loading || importing" class="hm-import-action" @click="inspect">Sicherung prüfen</b-button>
    </div>

    <b-notification v-if="error" class="hm-import-feedback" type="is-danger" :closable="false">{{ error }}</b-notification>
    <b-notification v-if="done" class="hm-import-feedback" type="is-success" :closable="false">
      <strong>Import abgeschlossen!</strong>
      <p>Neu übernommen: {{ done.vehicles }} Fahrzeuge, {{ done.fillups }} Tankungen und {{ done.expenses }} Ausgaben.</p>
      <p>Bereits importiert und übersprungen: {{ done.skippedFillups }} Tankungen, {{ done.skippedExpenses }} Ausgaben.</p>
      <router-link to="/">Zum Dashboard</router-link>
    </b-notification>

    <div v-if="preview" class="box hm-import-panel hm-import-step">
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
