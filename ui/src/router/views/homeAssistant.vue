<script>
import Layout from '@layouts/main.vue'
import axios from 'axios'

export default {
  page: { title: 'Home Assistant' },
  components: { Layout },
  data() {
    return {
      loading: true,
      busy: false,
      configured: false,
      createdAt: '',
      token: '',
      error: '',
      success: '',
      baseUrl: '',
    }
  },
  mounted() {
    this.baseUrl = window.location.origin
    this.load()
  },
  beforeDestroy() {
    this.token = ''
  },
  methods: {
    async load() {
      this.loading = true
      try {
        const response = await axios.get('/api/integrations/homeassistant/key')
        this.configured = !!response.data.configured
        this.createdAt = response.data.createdAt || ''
      } catch (e) {
        this.error = 'Der Home-Assistant-Zugang konnte nicht abgefragt werden.'
      } finally {
        this.loading = false
      }
    },
    async generate() {
      if (this.busy) return
      if (this.configured && !window.confirm('Der bisherige Home-Assistant-Schlüssel wird sofort ungültig. Wirklich einen neuen erstellen?')) return
      this.busy = true
      this.error = ''
      this.success = ''
      this.token = ''
      try {
        const response = await axios.post('/api/integrations/homeassistant/key')
        this.token = response.data.token
        this.configured = true
        this.createdAt = response.data.createdAt || ''
        this.success = 'Schlüssel erstellt. Bitte jetzt kopieren – er wird nur einmal angezeigt.'
      } catch (e) {
        this.error = 'Der Schlüssel konnte nicht erzeugt werden.'
      } finally {
        this.busy = false
      }
    },
    async revoke() {
      if (this.busy || !this.configured || !window.confirm('Home Assistant verliert sofort den Zugriff auf deine Wartungen. Schlüssel widerrufen?')) return
      this.busy = true
      this.error = ''
      this.success = ''
      try {
        await axios.delete('/api/integrations/homeassistant/key')
        this.configured = false
        this.createdAt = ''
        this.token = ''
        this.success = 'Home-Assistant-Schlüssel widerrufen.'
      } catch (e) {
        this.error = 'Der Schlüssel konnte nicht widerrufen werden.'
      } finally {
        this.busy = false
      }
    },
    async copyToken() {
      if (!this.token) return
      try {
        await navigator.clipboard.writeText(this.token)
        this.success = 'Schlüssel in die Zwischenablage kopiert.'
      } catch (e) {
        this.error = 'Automatisches Kopieren nicht möglich. Bitte den Schlüssel im Feld markieren und manuell kopieren.'
      }
    },
  },
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading">
      <div>
        <p class="hm-eyebrow">SMART HOME</p>
        <h1 class="title hm-page-title">Home Assistant</h1>
        <p class="hm-page-description">Offene Wartungen in Home Assistant als Sensoren anzeigen.</p>
      </div>
    </div>
    <b-notification v-if="error" type="is-danger" @close="error = ''">{{ error }}</b-notification>
    <b-notification v-if="success" type="is-success" @close="success = ''">{{ success }}</b-notification>
    <div class="box ha-panel">
      <h2 class="title is-5">Persönlicher Lesezugriff</h2>
      <p>Die Integration darf nur deine offenen Wartungen und deren Fälligkeit lesen. Sie kann keine Wartungen abschließen oder Ausgaben, Tankdaten und Passwörter abrufen.</p>
      <p v-if="loading">Verbindung wird geprüft …</p>
      <p v-else><strong>Status:</strong> {{ configured ? 'Schlüssel aktiv' : 'Kein Schlüssel angelegt' }}<span v-if="createdAt"> · Erstellt am {{ new Date(createdAt).toLocaleDateString('de-DE') }}</span></p>
      <div class="buttons">
        <b-button type="is-primary" :disabled="busy || loading" :loading="busy" @click="generate">{{ configured ? 'Schlüssel erneuern' : 'Schlüssel erstellen' }}</b-button>
        <b-button v-if="configured" type="is-danger" outlined :disabled="busy" @click="revoke">Schlüssel widerrufen</b-button>
      </div>
      <div v-if="token" class="ha-secret">
        <b-field label="Neuer API-Schlüssel – wird nur jetzt angezeigt">
          <b-input :value="token" readonly expanded />
        </b-field>
        <b-button @click="copyToken">Schlüssel kopieren</b-button>
        <p class="help">Speichere den Schlüssel im Passwortmanager oder direkt in Home Assistant. Beim nächsten Öffnen kann Hammond ihn nicht erneut anzeigen.</p>
      </div>
    </div>
    <div class="box ha-panel">
      <h2 class="title is-5">Integration einrichten</h2>
      <p>1. Kopiere den Ordner <code>custom_components/hammond_modern</code> aus dem GitHub-Repository in den Ordner <code>config/custom_components</code> deiner Home-Assistant-Installation.</p>
      <p>2. Starte Home Assistant neu und öffne <strong>Einstellungen → Geräte &amp; Dienste → Integration hinzufügen → Hammond Modern</strong>.</p>
      <p>3. Gib die von Home Assistant erreichbare Hammond-Adresse und deinen persönlichen API-Schlüssel ein.</p>
      <b-field label="Beispieladresse (aus deinem Browser)"><b-input :value="baseUrl" readonly /></b-field>
      <p class="help">Wichtig: Home Assistant muss die Adresse selbst erreichen können. Bei Dockhand/macvlan gegebenenfalls die IP deines Hammond-Containers verwenden, z. B. <code>http://192.168.0.36:3000</code> nur falls auf dieser IP der Dienst tatsächlich auf Port 3000 erreichbar ist. Eine HTTPS-Adresse ist für Zugriff außerhalb des vertrauenswürdigen LANs nötig.</p>
      <p>Danach siehst du die Sensoren <strong>Wartungen offen, Überfällig, Heute fällig, Bald fällig, Nächste Wartung</strong> und für jede Wartung einen eigenen Statussensor.</p>
      <p>Eine ausführliche Anleitung steht in <a href="https://github.com/Skoun181/hammond-modern/blob/develop/hammond-modern-next/docs/HOME-ASSISTANT.md" target="_blank" rel="noopener noreferrer">docs/HOME-ASSISTANT.md</a>.</p>
    </div>
  </Layout>
</template>

<style scoped>
.ha-panel { max-width: 900px; display: flex; flex-direction: column; gap: 12px; overflow-wrap: anywhere; }
.ha-panel strong { color: var(--hm-text); }
.ha-panel code { overflow-wrap: anywhere; }
.ha-secret { border: 1px solid var(--hm-border); background: var(--hm-input); border-radius: 10px; padding: 14px; }
.ha-secret .help { margin-top: 10px; }
</style>
