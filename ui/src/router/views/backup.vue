<script>
import Layout from '@layouts/main.vue'
import store from '@state/store'

const MAX_BACKUP_SIZE = 512 * 1024 * 1024

export default {
  page: { title: 'Backup & Serverumzug' },
  components: { Layout },
  data() {
    return {
      file: null, preview: null, result: null,
      exporting: false, checking: false, restoring: false,
      confirmed: false, confirmWord: '', error: '', success: '',
    }
  },
  computed: {
    canRestore() {
      return this.preview && this.file && this.confirmed &&
        this.confirmWord === 'ERSETZEN' && !this.restoring && !this.checking
    },
  },
  methods: {
    token() {
      return store.state.auth.currentUser && store.state.auth.currentUser.token
    },
    headers() {
      return this.token() ? { Authorization: this.token() } : {}
    },
    setFile(event) {
      const files = event.target.files
      this.file = files && files.length ? files[0] : null
      this.preview = null
      this.result = null
      this.confirmed = false
      this.confirmWord = ''
      this.error = ''
      this.success = ''
    },
    async readError(response) {
      const data = await response.json().catch(() => ({}))
      return new Error(data.error || 'Serverfehler (HTTP ' + response.status + ')')
    },
    async exportBackup() {
      this.exporting = true
      this.error = ''
      this.success = ''
      try {
        const response = await fetch('/api/backup/export', {
          headers: this.headers(), credentials: 'same-origin',
          cache: 'no-store',
        })
        if (!response.ok) throw await this.readError(response)
        const blob = await response.blob()
        const matches = (response.headers.get('Content-Disposition') || '')
          .match(/filename="?([^";]+)"?/)
        const filename = matches ? matches[1] : 'hammond-modern-backup.zip'
        const url = window.URL.createObjectURL(blob)
        const anchor = document.createElement('a')
        anchor.href = url
        anchor.download = filename
        document.body.appendChild(anchor)
        anchor.click()
        document.body.removeChild(anchor)
        window.URL.revokeObjectURL(url)
        this.success = 'Backup wurde erstellt. Speichere die ZIP-Datei an einem sicheren Ort.'
        this.$buefy.toast.open({ message: 'Backup erstellt!', type: 'is-success' })
      } catch (e) { this.error = e.message || 'Export fehlgeschlagen.' }
      finally { this.exporting = false }
    },
    validateFile() {
      if (!this.file) throw new Error('Bitte zuerst eine ZIP-Datei auswählen.')
      if (!this.file.name.toLowerCase().endsWith('.zip')) throw new Error('Nur ZIP-Dateien werden akzeptiert.')
      if (this.file.size === 0 || this.file.size > MAX_BACKUP_SIZE) throw new Error('Die ZIP-Datei muss zwischen 1 Byte und 512 MiB groß sein.')
    },
    async sendBackup(path, confirm) {
      this.validateFile()
      const form = new FormData()
      form.append('file', this.file, this.file.name)
      if (confirm) form.append('confirm', 'ERSETZEN')
      const response = await fetch(path, {
        method: 'POST', headers: this.headers(), body: form,
        credentials: 'same-origin', cache: 'no-store',
      })
      if (!response.ok) throw await this.readError(response)
      return response.json()
    },
    async inspectBackup() {
      this.error = ''
      this.success = ''
      this.preview = null
      this.confirmed = false
      this.confirmWord = ''
      this.checking = true
      try { this.preview = await this.sendBackup('/api/backup/preview', false) }
      catch (e) { this.error = e.message }
      finally { this.checking = false }
    },
    async restoreBackup() {
      if (!this.canRestore) return
      this.restoring = true
      this.error = ''
      this.success = ''
      try {
        this.result = await this.sendBackup('/api/backup/restore', true)
        this.preview = null
        this.success = 'Wiederherstellung erfolgreich! Bitte neu anmelden. Das Passwort des Ziel-Administrators bleibt gültig.'
        this.$buefy.toast.open({ message: 'Daten wiederhergestellt!', type: 'is-success' })
      } catch (e) { this.error = e.message || 'Wiederherstellung fehlgeschlagen.' }
      finally { this.restoring = false }
    },
    async logoutAfterRestore() {
      await store.dispatch('auth/logOut')
      this.$router.push({ name: 'login' })
    },
  },
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading">
      <div>
        <p class="hm-eyebrow">BACKUP & UMZUG</p>
        <h1 class="title hm-page-title">Daten sichern & wiederherstellen</h1>
        <p class="hm-page-description">Übertrage Hammond Modern samt Fahrzeugen, Tankungen, Rechnungen und Einstellungen auf eine andere Installation.</p>
      </div>
    </div>

    <b-notification v-if="error" type="is-danger" :closable="true" @close="error = ''">{{ error }}</b-notification>
    <b-notification v-if="success" type="is-success" :closable="true" @close="success = ''">{{ success }}</b-notification>

    <div class="box">
      <h2 class="title is-5">1. Backup erstellen</h2>
      <p>Exportiert eine konsistente SQLite-Datenbank, Fahrzeughistorie, Reifen, Benutzerprofile, Wartungen und Uploads in eine ZIP-Datei mit Prüfsummen.</p>
      <p class="help">Nicht enthalten: Passwörter und Passwort-Hashes, API-Keys, JWT-Secrets, .env, Docker-Compose-Dateien. Die ZIP enthält personenbezogene Fahrzeugdaten – sicher aufbewahren.</p>
      <br />
      <b-button type="is-primary" icon-left="download" :loading="exporting" :disabled="checking || restoring" @click="exportBackup">
        {{ exporting ? 'Backup wird erstellt …' : 'Backup herunterladen (.zip)' }}
      </b-button>
    </div>

    <div class="box">
      <h2 class="title is-5">2. Backup auswählen und prüfen</h2>
      <p>Wähle ein zuvor aus Hammond Modern exportiertes ZIP. Die Integrität, das Datenbankschema und die Admin-Zuordnung werden vor der Wiederherstellung geprüft.</p>
      <p class="help">Maximal 512 MiB ZIP und 1 GiB entpackte Daten. Auf der Zielinstallation muss bereits ein Administrator mit derselben E-Mail-Adresse wie im Backup eingerichtet sein.</p>
      <br />
      <div class="file has-name is-fullwidth">
        <label class="file-label">
          <input class="file-input" type="file" accept=".zip,application/zip" :disabled="checking || restoring" @change="setFile" />
          <span class="file-cta"><span class="file-label">ZIP-Datei auswählen</span></span>
          <span class="file-name">{{ file ? file.name : 'Noch keine Datei ausgewählt' }}</span>
        </label>
      </div>
      <br />
      <b-button type="is-primary" :disabled="!file || checking || restoring" :loading="checking" @click="inspectBackup">
        {{ checking ? 'Backup wird geprüft …' : 'Backup prüfen' }}
      </b-button>

      <div v-if="preview" class="notification is-info is-light" style="margin-top: 1rem">
        <h3 class="title is-6">Backup ist kompatibel</h3>
        <p>Erstellt: {{ new Date(preview.manifest.createdAt).toLocaleString('de-DE') }}</p>
        <p>App-Version: {{ preview.manifest.appVersion }} · Backup-Format: {{ preview.manifest.formatVersion }}</p>
        <p>Fahrzeuge: {{ preview.manifest.counts.vehicles || 0 }} · Tankungen: {{ preview.manifest.counts.fillups || 0 }} · Ausgaben: {{ preview.manifest.counts.expenses || 0 }}</p>
        <p>Wartungen: {{ preview.manifest.counts.maintenance_tasks || 0 }} · Reifen: {{ preview.manifest.counts.tires || 0 }} · Benutzer: {{ preview.manifest.counts.users || 0 }} · Dateien: {{ preview.assetCount }}</p>
        <p v-if="!Object.prototype.hasOwnProperty.call(preview.manifest.counts, 'tires')" class="has-text-warning-dark">Älteres Backup ohne Reifendaten. Bei der Wiederherstellung werden vorhandene Reifen ebenfalls ersetzt (Reifenliste danach leer).</p>
      </div>
    </div>

    <div class="box">
      <h2 class="title is-5">3. Wiederherstellung bestätigen</h2>
      <b-notification type="is-warning" :closable="false">
        <strong>ACHTUNG: Der Import ersetzt alle vorhandenen Anwendungsdaten und Uploads!</strong>
        Der Server erstellt vorher automatisch eine lokale Sicherheitskopie.
        Das bestehende Ziel-Admin-Passwort bleibt erhalten; weitere Konten benötigen neue Passwörter.
        Sichere wichtige Daten zusätzlich außerhalb dieses Containers.
      </b-notification>
      <b-checkbox v-model="confirmed" :disabled="!preview || restoring">
        Mir ist bewusst, dass vorhandene Daten ersetzt werden.
      </b-checkbox>
      <b-field label='Zur Bestätigung "ERSETZEN" eintippen' style="margin-top: 1rem">
        <b-input v-model="confirmWord" placeholder="ERSETZEN" :disabled="!preview || restoring" />
      </b-field>
      <b-button type="is-danger" :loading="restoring" :disabled="!canRestore" @click="restoreBackup">
        {{ restoring ? 'Wiederherstellung läuft …' : 'Daten endgültig wiederherstellen' }}
      </b-button>
      <div v-if="result" class="content" style="margin-top: 1.5rem">
        <h3 class="title is-6">Import abgeschlossen</h3>
        <p>Importiert: {{ result.manifest.counts.vehicles || 0 }} Fahrzeuge,
          {{ result.manifest.counts.fillups || 0 }} Tankungen,
          {{ result.manifest.counts.expenses || 0 }} Ausgaben und
          {{ result.assetCount }} Dateien.</p>
        <p>Reifen: {{ result.manifest.counts.tires || 0 }}</p>
        <p>Weitere Benutzerpasswörter können unter <router-link to="/admin/users">Benutzer</router-link> neu vergeben werden.</p>
        <b-button type="is-primary" @click="logoutAfterRestore">Abmelden und neu anmelden</b-button>
      </div>
    </div>
    <div class="box">
      <p class="hm-eyebrow">DATENIMPORT</p>
      <h2 class="title is-5">4. Daten aus anderen Anwendungen übernehmen</h2>
      <p>Diese Importe unterscheiden sich vom ZIP-Restore: Sie übernehmen ausgewählte Daten aus einem anderen Format.</p>
      <div class="columns" style="margin-top:.6rem">
        <div class="column">
          <h3 class="title is-6">Original-Hammond (SQLite)</h3>
          <p>Fahrzeuge, Tankungen und Ausgaben aus einer hammond.db mit Vorschau und Duplikatprüfung.</p>
          <b-button style="margin-top:.6rem" type="is-primary" tag="router-link" :to="{ name: 'import-hammond' }">Hammond-Migration starten</b-button>
        </div>
        <div class="column">
          <h3 class="title is-6">Fuelly (CSV)</h3>
          <p>Tankungen aus einer CSV-Datei importieren. Fahrzeuge müssen vorher angelegt sein.</p>
          <b-button style="margin-top:.6rem" tag="router-link" :to="{ name: 'import-fuelly' }">CSV importieren</b-button>
        </div>
      </div>
    </div>
  </Layout>
</template>
