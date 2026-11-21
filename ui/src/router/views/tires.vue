<script>
import Layout from '@layouts/main.vue'
import axios from 'axios'

const seasons = { summer: 'Sommerreifen', winter: 'Winterreifen', 'all-season': 'Ganzjahresreifen', other: 'Sonstige' }
const statuses = { stored: 'Eingelagert', mounted: 'Montiert', retired: 'Ausgemustert' }
const wheelPositions = ['front-left', 'front-right', 'rear-left', 'rear-right']
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
      form: emptyTire(), quantity: 1, batchPositions: [], wheelPositions, initialFilterApplied: false, editingId: null, showForm: false, loading: true, busy: false, error: '', success: '' }
  },
  computed: {
    visibleTires() {
      const q = this.search.trim().toLowerCase()
      return this.tires.filter(t => (!this.filter || (this.filter === 'unassigned' ? !t.vehicleId : t.vehicleId === this.filter)) &&
        [t.manufacturer,t.model,t.size,t.dot,t.setName,t.storage,this.vehicleName(t.vehicleId)].join(' ').toLowerCase().includes(q))
    },
    batchMounted() { return !this.editingId && this.quantity > 1 && this.form.status === 'mounted' },
    canSave() { return !this.busy && (!this.batchMounted || (this.form.vehicleId && this.batchPositions.length === this.quantity)) },
    unavailableLink() { return this.form.vehicleId && !this.vehicles.some(v => v.id === this.form.vehicleId) },
  },
  watch: {
    '$route.query.vehicleId'() { this.applyRouteFilter() },
    quantity(value) { this.batchPositions = wheelPositions.slice(0, Number(value)) },
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
    vehicle(id) { return this.vehicles.find(v => v.id === id) },
    vehicleName(id) {
      if (!id) return 'Nicht zugeordnet'
      const v = this.vehicle(id)
      return v ? [v.nickname || [v.make, v.model].filter(Boolean).join(' ') || 'Fahrzeug', v.registration].filter(Boolean).join(' · ') : 'Fahrzeug nicht mehr verfügbar'
    },
    add() {
      this.form = emptyTire(this.vehicles.some(v => v.id === this.filter) ? this.filter : '')
      this.quantity = 1; this.batchPositions = []; this.editingId = null; this.showForm = true; this.error = ''; this.success = ''
    },
    edit(t) {
      this.form = Object.assign(emptyTire(), t)
      delete this.form.id; delete this.form.createdAt; delete this.form.updatedAt
      this.quantity = 1; this.batchPositions = []; this.editingId = t.id; this.showForm = true; this.error = ''; this.success = ''
    },
    cancel() { this.showForm = false; this.editingId = null; this.quantity = 1; this.batchPositions = []; this.form = emptyTire() },
    updateVehicle() { if (!this.form.vehicleId && this.form.status === 'mounted') { this.form.status = 'stored'; this.form.position = '' } },
    async save() {
      if (!this.canSave) return
      this.busy = true; this.error = ''; this.success = ''
      try {
        const payload = {...this.form}
        for (const field of ['treadDepth','pressure','purchasePrice','mileage']) {
          const value = payload[field]
          payload[field] = value === '' || value === null || value === undefined ? null : Number(value)
        }
        if (payload.status !== 'mounted') payload.position = ''
        if (this.editingId) await axios.put('/api/tires/' + this.editingId, payload)
        else {
          payload.quantity = Number(this.quantity)
          if (this.batchMounted) {
            payload.position = ''
            payload.positions = [...this.batchPositions]
          }
          await axios.post('/api/tires', payload)
        }
        const message = this.editingId ? 'Reifen aktualisiert.' : this.quantity + (this.quantity === 1 ? ' Reifen angelegt.' : ' Reifen erfolgreich angelegt.')
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
    <b-notification v-if="error" type="is-danger" :closable="true" @close="error = ''">{{ error }}</b-notification>
    <b-notification v-if="success" type="is-success" :closable="true" @close="success = ''">{{ success }}</b-notification>
    <div v-if="showForm" class="box tire-form">
      <h2 class="title is-5">{{ editingId ? 'Reifen bearbeiten' : 'Reifen hinzufügen' }}</h2>
      <form @submit.prevent="save">
        <div v-if="!editingId" class="tire-batch-panel">
          <b-field label="Anzahl der Reifen">
            <b-select v-model.number="quantity" expanded>
              <option v-for="n in [1,2,3,4]" :key="n" :value="n">{{ n }} {{ n === 1 ? 'Reifen' : 'Reifen gleichzeitig' }}</option>
            </b-select>
          </b-field>
          <p class="help">Alle Angaben gelten pro Reifen. Jeder Eintrag bleibt anschließend einzeln bearbeitbar. Kaufpreis, Profil, Druck und Laufleistung werden übernommen.</p>
        </div>
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
          <div v-if="form.status === 'mounted' && !batchMounted" class="column is-half"><b-field label="Einbauposition *"><b-select v-model="form.position" expanded required><option v-for="(label,key) in positions" :key="key" :value="key">{{ label }}</option></b-select></b-field></div>
          <div v-if="batchMounted" class="column is-full">
            <div class="tire-batch-panel">
              <p class="label">Radpositionen auswählen ({{ batchPositions.length }}/{{ quantity }})</p>
              <div class="tire-position-grid">
                <b-checkbox v-for="position in wheelPositions" :key="position" v-model="batchPositions" :native-value="position">{{ positions[position] }}</b-checkbox>
              </div>
              <p class="help">Wähle genau {{ quantity }} unterschiedliche Positionen. Bereits montierte Reifen an diesen Positionen zuerst einlagern.</p>
            </div>
          </div>
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
        <div class="buttons"><b-button type="is-primary" native-type="submit" :loading="busy" :disabled="!canSave">{{ editingId ? 'Änderungen speichern' : quantity + ' Reifen speichern' }}</b-button><b-button :disabled="busy" @click="cancel">Abbrechen</b-button></div>
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
          <div class="tire-vehicle">
            <p class="tire-label">ZUGEHÖRIGES FAHRZEUG</p>
            <router-link v-if="vehicle(t.vehicleId)" class="tire-vehicle-name" :to="{ name: 'vehicle-detail', params: { vehicleId: t.vehicleId } }">{{ vehicle(t.vehicleId).nickname || [vehicle(t.vehicleId).make, vehicle(t.vehicleId).model].filter(Boolean).join(' ') || 'Fahrzeug' }}</router-link>
            <p v-else class="tire-vehicle-name">{{ vehicleName(t.vehicleId) }}</p>
            <div v-if="vehicle(t.vehicleId)" class="tire-vehicle-meta">
              <span v-if="vehicle(t.vehicleId).registration" class="tire-plate">{{ vehicle(t.vehicleId).registration }}</span>
              <span>{{ [vehicle(t.vehicleId).make, vehicle(t.vehicleId).model].filter(Boolean).join(' ') }}</span>
            </div>
          </div>
          <div class="tire-status-row"><span class="tire-badge">{{ seasons[t.season] }}</span><span class="tire-badge">{{ statuses[t.status] }}</span><span v-if="t.position" class="tire-badge">{{ positions[t.position] }}</span></div>
          <h2 class="title is-5 tire-title">{{ t.manufacturer }} {{ t.model }}</h2>
          <p><strong>{{ t.size }}</strong> {{ [t.loadIndex,t.speedIndex].filter(Boolean).join(' ') }} <span v-if="t.runFlat">· Runflat</span></p>
          <p v-if="t.setName">Satz: {{ t.setName }}</p>
          <p>DOT: {{ dotText(t.dot) }}</p>
          <div class="tire-metrics">
            <div><p class="tire-label">Profil</p><strong>{{ measure(t.treadDepth,'mm') }}</strong></div>
            <div><p class="tire-label">Druck</p><strong>{{ measure(t.pressure,'bar') }}</strong></div>
            <div><p class="tire-label">Laufleistung</p><strong>{{ measure(t.mileage,'km') }}</strong></div>
          </div>
          <p v-if="t.purchaseDate">Gekauft: {{ t.purchaseDate }}</p><p v-if="t.purchasePrice !== null">Preis: {{ measure(t.purchasePrice,t.currency) }}</p>
          <p v-if="t.storage">Lagerort: {{ t.storage }}</p><p v-if="t.notes" class="tire-notes">{{ t.notes }}</p>
          <div class="buttons tire-actions"><b-button type="is-primary" size="is-small" :disabled="busy" @click="edit(t)">Bearbeiten</b-button><b-button size="is-small" type="is-danger is-light" :disabled="busy" @click="remove(t)">Löschen</b-button></div>
        </article>
      </div>
    </div>
  </Layout>
</template>

<style scoped>
.tire-card { height:100%; display:flex; flex-direction:column; gap:.45rem; overflow-wrap:anywhere; color:var(--hm-text)!important; }
.tire-card strong, .tire-card .tire-title { color:var(--hm-text)!important; }
.tire-card p:not(.tire-label):not(.tire-vehicle-name) { color:var(--hm-text); }
.tire-vehicle { background:var(--hm-input); border:1px solid var(--hm-border); border-left:4px solid var(--hm-accent); border-radius:10px; padding:12px 14px; margin-bottom:8px; }
.tire-label { color:var(--hm-muted)!important; font-size:.75rem; font-weight:650; letter-spacing:.035em; line-height:1.4; margin-bottom:4px; }
.tire-vehicle-name { color:var(--hm-text)!important; display:block; font-size:clamp(1.2rem,2.3vw,1.6rem); font-weight:800; line-height:1.25; }
a.tire-vehicle-name:hover { text-decoration:underline; }
.tire-vehicle-meta { display:flex; flex-wrap:wrap; align-items:center; gap:8px; margin-top:8px; color:var(--hm-text); font-size:.86rem; }
.tire-plate { border:1px solid var(--hm-border); border-radius:5px; color:var(--hm-text); background:var(--hm-surface); font-size:.9rem; font-weight:750; padding:3px 9px; letter-spacing:.06em; }
.tire-status-row { display:flex; flex-wrap:wrap; gap:6px; margin-bottom:3px; }
.tire-badge { background:var(--hm-input); color:var(--hm-text); border:1px solid var(--hm-border); padding:3px 8px; border-radius:6px; font-size:.78rem; font-weight:650; }
.tire-title { margin:4px 0!important; }
.tire-metrics { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:7px; margin:8px 0; }
.tire-metrics > div { background:var(--hm-input); border:1px solid var(--hm-border); padding:10px 8px; border-radius:8px; min-width:0; }
.tire-metrics strong { font-size:1.05rem; }
.tire-actions { margin-top:auto; padding-top:12px; }
.tire-notes { white-space:pre-wrap; margin-top:.6rem; }
.tire-batch-panel { background:var(--hm-input); border:1px solid var(--hm-border); border-radius:10px; padding:14px; margin-bottom:12px; }
.tire-position-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:8px; margin:10px 0; }
.tire-form .checkbox { color:var(--hm-text)!important; }
.tire-form .help { color:var(--hm-muted)!important; }
@media(max-width:420px) { .tire-metrics { grid-template-columns:1fr; } .tire-position-grid { grid-template-columns:1fr; } }
</style>
