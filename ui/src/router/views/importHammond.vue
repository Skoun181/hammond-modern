<script>
import { authenticatedFetch } from '@state/modules/auth'
import Layout from '@layouts/main.vue'

export default {
  page: { title: 'Hammond-Daten übernehmen' },
  components: { Layout },
  data() {
    return { file: null, preview: null, loading: false, importing: false, error: '', done: null, confirmed: false, selectedVehicles: [] }
  },
  computed: {
    canImport() { return this.preview && this.preview.vehicles && this.selectedVehicles.length > 0 && this.confirmed && !this.importing }
  },
  methods: {
    onFileChange(event) {
      const files = event.target.files
      this.file = files && files.length ? files[0] : null
      this.preview = null
      this.done = null
      this.confirmed = false
      this.selectedVehicles = []
      this.error = ''
    },
    formData(execute) {
      const form = new FormData()
      form.append('file', this.file, this.file.name)
      if (execute) {
        form.append('confirm', 'IMPORTIEREN')
        this.selectedVehicles.forEach(id => form.append('vehicleIds', id))
      }
      return form
    },
    showError(error) {
      this.error = error.message || 'Import nicht möglich.'
    },
    async sendUpload(endpoint, execute) {
      // The browser must generate the multipart boundary itself. Global Axios
      // defaults may force application/json and make Gin c.FormFile fail.
      const response = await authenticatedFetch(endpoint, {
        method: 'POST',
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
        this.selectedVehicles = []
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
        this.selectedVehicles = []
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
      <p>Gefunden: <strong>{{ preview.sourceVehicleCount || preview.vehicles.length }} Datenbankeinträge, {{ preview.totalFillups }} Tankungen und {{ preview.totalExpenses }} Ausgaben.</strong></p>
      <p v-if="preview.groupedDuplicates" class="hm-import-duplicate-note">
        {{ preview.groupedDuplicates }} vermutlich doppelte Fahrzeugeinträge zusammengefasst.
        Es bleiben <strong>{{ preview.vehicles.length }} unterschiedliche Fahrzeuge</strong> zur Auswahl.
      </p>
      <p class="help">Wähle per Häkchen aus, welche Fahrzeuge und deren gesamte Historie übernommen werden sollen. Standardmäßig ist nichts ausgewählt.</p>
      <div class="hm-import-selection-summary">{{ selectedVehicles.length }} von {{ preview.vehicles.length }} Fahrzeugen ausgewählt</div>
      <label v-for="vehicle in preview.vehicles" :key="vehicle.id" class="hm-import-vehicle-option">
        <input v-model="selectedVehicles" type="checkbox" :value="vehicle.id" :disabled="importing">
        <span class="hm-import-vehicle-content">
          <strong>{{ vehicle.name || '(Ohne Namen)' }} · {{ vehicle.registration || 'Ohne Kennzeichen' }}</strong>
          <span class="hm-import-vehicle-meta">{{ vehicle.fillups }} Tankungen · {{ vehicle.expenses }} Ausgaben</span>
          <span v-if="vehicle.duplicateCount" class="hm-import-duplicate-info">
            {{ vehicle.duplicateCount }} weitere identische Fahrzeugdatensätze erkannt – Historie wird zusammengeführt
          </span>
        </span>
        <span class="hm-import-vehicle-status">{{ vehicle.matched ? 'Bereits vorhanden' : 'Neues Fahrzeug' }}</span>
      </label>
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
