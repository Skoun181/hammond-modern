<script>
import Layout from '@layouts/main.vue'
import axios from 'axios'

const seasons = { summer: 'Sommerreifen', winter: 'Winterreifen', 'all-season': 'Ganzjahresreifen', other: 'Sonstige' }
const statuses = { stored: 'Eingelagert', mounted: 'Montiert', retired: 'Ausgemustert' }
const positions = { '': 'Nicht angegeben', 'front-left': 'Vorne links', 'front-right': 'Vorne rechts', 'rear-left': 'Hinten links', 'rear-right': 'Hinten rechts', spare: 'Ersatzrad', other: 'Sonstige' }
function emptyTire(vehicleId = '') {
  return { vehicleId, manufacturer: '', model: '', season: 'summer', size: '', dot: '', setName: '',
    loadIndex: '', speedIndex: '', runFlat: false, status: 'stored', position: '', treadDepth: null,
    pressure: null, purchaseDate: '', purchasePrice: null, currency: 'EUR', mileage: null, storage: '', notes: '' }
}
export default {
  page: { title: 'Reifen' }, components: { Layout },
  data() {
    return { tires: [], vehicles: [], seasons, statuses, positions, filter: '', search: '',
      form: emptyTire(), initialFilterApplied: false, editingId: null, showForm: false, loading: true, busy: false, error: '', success: '' }
  },
  computed: {
    visibleTires() {
      const q = this.search.trim().toLowerCase()
      return this.tires.filter(t => (!this.filter || (this.filter === 'unassigned' ? !t.vehicleId : t.vehicleId === this.filter)) &&
        [t.manufacturer,t.model,t.size,t.dot,t.setName,t.storage].join(' ').toLowerCase().includes(q))
    },
    unavailableLink() { return this.form.vehicleId && !this.vehicles.some(v => v.id === this.form.vehicleId) },
  },
  watch: {
    '$route.query.vehicleId'() { this.applyRouteFilter() },
  },
  mounted() { this.load() },
  methods: {
    applyRouteFilter() {
      const id = this.$route.query.vehicleId
      this.filter = this.vehicles.some(v => v.id === id) ? id : ''
    },
    async load() {
      this.loading = true; this.error = ''
      try {
        const [res, vehicles] = await Promise.all([axios.get('/api/tires'), this.$store.dispatch('vehicles/fetchVehicles')])
        this.tires = res.data || []; this.vehicles = vehicles || []
        if (!this.initialFilterApplied) { this.applyRouteFilter(); this.initialFilterApplied = true }
      } catch (_) { this.error = 'Reifen oder Fahrzeuge konnten nicht geladen werden.' }
      finally { this.loading = false }
    },
    vehicleName(id) {
      if (!id) return 'Nicht zugeordnet'
      const v = this.vehicles.find(x => x.id === id)
      return v ? [v.nickname, v.registration].filter(Boolean).join(' · ') : 'Fahrzeug nicht mehr verfügbar'
    },
    add() {
      this.form = emptyTire(this.vehicles.some(v => v.id === this.filter) ? this.filter : '')
      this.editingId = null; this.showForm = true; this.error = ''; this.success = ''
    },
    edit(t) {
      this.form = Object.assign(emptyTire(), t)
      delete this.form.id; delete this.form.createdAt; delete this.form.updatedAt
      this.editingId = t.id; this.showForm = true; this.error = ''; this.success = ''
    },
    cancel() { this.showForm = false; this.editingId = null; this.form = emptyTire() },
    updateVehicle() { if (!this.form.vehicleId && this.form.status === 'mounted') { this.form.status = 'stored'; this.form.position = '' } },
    async save() {
      if (this.busy) return
      this.busy = true; this.error = ''; this.success = ''
      try {
        const payload = {...this.form}
        for (const field of ['treadDepth','pressure','purchasePrice','mileage']) {
          const value = payload[field]
          payload[field] = value === '' || value === null || value === undefined ? null : Number(value)
        }
        if (payload.status !== 'mounted') payload.position = ''
        if (this.editingId) await axios.put('/api/tires/' + this.editingId, payload)
        else await axios.post('/api/tires', payload)
        const message = this.editingId ? 'Reifen aktualisiert.' : 'Reifen angelegt.'
        this.cancel(); await this.load(); this.success = message
      } catch (e) { this.error = (e.response && e.response.data && e.response.data.error) || 'Speichern fehlgeschlagen.' }
      finally { this.busy = false }
    },
    async remove(t) {
      if (this.busy || !window.confirm('Reifen endgültig löschen?')) return
      this.busy = true; this.error = ''; this.success = ''
      try { await axios.delete('/api/tires/' + t.id); if(this.editingId===t.id)this.cancel(); await this.load(); this.success = 'Reifen gelöscht.' }
      catch (_) { this.error = 'Löschen fehlgeschlagen.' }
      finally { this.busy = false }
    },
    measure(value, unit) { return value === null || value === undefined ? '–' : value.toLocaleString('de-DE') + ' ' + unit },
    dotText(dot) { return dot ? dot + ' (KW ' + dot.slice(0,2) + '/' + dot.slice(2) + ')' : '–' },
  },
}
</script>
<template>
  <Layout>
    <div class="hm-page-heading">
      <div><p class="hm-eyebrow">VERWALTUNG</p><h1 class="title hm-page-title">Reifen</h1><p class="hm-page-description">Verwalte einzelne Reifen, Reifensätze, Montage und Einlagerung. Verknüpfe Reifen mit deinen Fahrzeugen.</p></div>
      <b-button type="is-primary" :disabled="busy || loading" @click="add">＋ Reifen hinzufügen</b-button>
    </div>
    <b-notification v-if="error" type="is-danger" :closable="true" @close="error=''">{{ error }}</b-notification>
    <b-notification v-if="success" type="is-success" :closable="true" @close="success=''">{{ success }}</b-notification>
    <div v-if="showForm" class="box">
      <h2 class="title is-5">{{ editingId ? 'Reifen bearbeiten' : 'Reifen hinzufügen' }}</h2>
      <form @submit.prevent="save">
        <div class="columns is-multiline">
          <div class="column is-half"><b-field label="Hersteller *"><b-input v-model="form.manufacturer" maxlength="150" required placeholder="Michelin, Continental …" /></b-field></div>
          <div class="column is-half"><b-field label="Modell"><b-input v-model="form.model" maxlength="150" placeholder="Reifenmodell" /></b-field></div>
          <div class="column is-half"><b-field label="Reifengröße *"><b-input v-model="form.size" maxlength="150" required placeholder="225/45 R17" /></b-field></div>
          <div class="column is-half"><b-field label="Reifenart"><b-select v-model="form.season" expanded><option v-for="(label,key) in seasons" :key="key" :value="key">{{ label }}</option></b-select></b-field></div>
          <div class="column is-half"><b-field label="DOT (Produktionswoche/Jahr)" message="Vier Ziffern, z. B. 2325 = KW 23/2025"><b-input v-model="form.dot" maxlength="4" placeholder="2325" /></b-field></div>
          <div class="column is-half"><b-field label="Reifensatz / Bezeichnung"><b-input v-model="form.setName" maxlength="150" placeholder="Winterradsatz 2025" /></b-field></div>
          <div class="column is-half"><b-field label="Traglastindex"><b-input v-model="form.loadIndex" maxlength="12" placeholder="91" /></b-field></div>
          <div class="column is-half"><b-field label="Geschwindigkeitsindex"><b-input v-model="form.speedIndex" maxlength="12" placeholder="V" /></b-field></div>
          <div class="column is-half"><b-field label="Fahrzeug"><b-select v-model="form.vehicleId" expanded @input="updateVehicle"><option value="">Nicht zugeordnet</option><option v-if="unavailableLink" :value="form.vehicleId" disabled>Nicht mehr verfügbar – bitte neu zuordnen</option><option v-for="v in vehicles" :key="v.id" :value="v.id">{{ vehicleName(v.id) }}</option></b-select></b-field></div>
          <div class="column is-half"><b-field label="Status"><b-select v-model="form.status" expanded><option v-for="(label,key) in statuses" :key="key" :value="key">{{ label }}</option></b-select></b-field></div>
          <div v-if="form.status==='mounted'" class="column is-half"><b-field label="Einbauposition *"><b-select v-model="form.position" expanded required><option v-for="(label,key) in positions" :key="key" :value="key">{{ label }}</option></b-select></b-field></div>
          <div class="column is-half"><b-field label="Profiltiefe (mm)"><b-input v-model="form.treadDepth" type="number" min="0" max="30" step="0.1" /></b-field></div>
          <div class="column is-half"><b-field label="Reifendruck (bar)"><b-input v-model="form.pressure" type="number" min="0" max="20" step="0.1" /></b-field></div>
          <div class="column is-half"><b-field label="Laufleistung des Reifens (km)"><b-input v-model="form.mileage" type="number" min="0" max="10000000" step="1" /></b-field></div>
          <div class="column is-half"><b-field label="Kaufdatum"><input v-model="form.purchaseDate" class="input" type="date" /></b-field></div>
          <div class="column is-half"><b-field label="Kaufpreis pro Reifen"><b-input v-model="form.purchasePrice" type="number" min="0" max="1000000" step="0.01" /></b-field></div>
          <div class="column is-half"><b-field label="Währung"><b-input v-model="form.currency" maxlength="3" placeholder="EUR" /></b-field></div>
          <div class="column is-half"><b-field label="Einlagerungsort"><b-input v-model="form.storage" maxlength="150" placeholder="Garage, Reifenhotel …" /></b-field></div>
          <div class="column is-full"><b-checkbox v-model="form.runFlat">Runflat / Notlaufeigenschaften</b-checkbox></div>
          <div class="column is-full"><b-field label="Notizen"><b-input v-model="form.notes" type="textarea" maxlength="2000" /></b-field></div>
        </div>
        <div class="buttons"><b-button type="is-primary" native-type="submit" :loading="busy">{{ editingId ? 'Änderungen speichern' : 'Reifen speichern' }}</b-button><b-button :disabled="busy" @click="cancel">Abbrechen</b-button></div>
      </form>
    </div>
    <div class="box">
      <div class="columns">
        <div class="column"><b-field label="Fahrzeug filtern"><b-select v-model="filter" expanded><option value="">Alle meine Reifen</option><option value="unassigned">Nicht zugeordnet</option><option v-for="v in vehicles" :key="v.id" :value="v.id">{{ vehicleName(v.id) }}</option></b-select></b-field></div>
        <div class="column"><b-field label="Suche"><b-input v-model="search" placeholder="Hersteller, Modell, DOT, Satz …" /></b-field></div>
      </div>
      <p class="help">{{ visibleTires.length }} Reifen · Zuordnungen sind auf freigegebene Fahrzeuge beschränkt.</p>
    </div>
    <p v-if="loading" role="status">Reifen werden geladen …</p>
    <p v-else-if="!visibleTires.length" class="box">{{ tires.length ? 'Keine passenden Reifen.' : 'Noch keine Reifen angelegt.' }}</p>
    <div class="columns is-multiline">
      <div v-for="t in visibleTires" :key="t.id" class="column is-half">
        <article class="box tire-card">
          <p class="hm-eyebrow">{{ seasons[t.season] }} · {{ statuses[t.status] }}</p>
          <h2 class="title is-5">{{ t.manufacturer }} {{ t.model }}</h2>
          <p><strong>{{ t.size }}</strong> {{ [t.loadIndex,t.speedIndex].filter(Boolean).join(' ') }} <span v-if="t.runFlat">· Runflat</span></p>
          <p v-if="t.setName">Satz: {{ t.setName }}</p>
          <p>Fahrzeug: <strong>{{ vehicleName(t.vehicleId) }}</strong><span v-if="t.position"> · {{ positions[t.position] }}</span></p>
          <p>DOT: {{ dotText(t.dot) }}</p>
          <div class="columns is-mobile" style="margin: .6rem 0">
            <div class="column"><p class="heading">Profil</p><strong>{{ measure(t.treadDepth,'mm') }}</strong></div>
            <div class="column"><p class="heading">Druck</p><strong>{{ measure(t.pressure,'bar') }}</strong></div>
            <div class="column"><p class="heading">Laufleistung</p><strong>{{ measure(t.mileage,'km') }}</strong></div>
          </div>
          <p v-if="t.purchaseDate">Gekauft: {{ t.purchaseDate }}</p><p v-if="t.purchasePrice!==null">Preis: {{ measure(t.purchasePrice,t.currency) }}</p>
          <p v-if="t.storage">Lagerort: {{ t.storage }}</p><p v-if="t.notes" class="tire-notes">{{ t.notes }}</p>
          <div class="buttons" style="margin-top:1rem"><b-button type="is-primary" size="is-small" :disabled="busy" @click="edit(t)">Bearbeiten</b-button><b-button size="is-small" type="is-danger is-light" :disabled="busy" @click="remove(t)">Löschen</b-button></div>
        </article>
      </div>
    </div>
  </Layout>
</template>
<style scoped>
.tire-card { height: 100%; overflow-wrap: anywhere; }
.tire-notes { white-space: pre-wrap; margin-top: .6rem; }
</style>
