<script>
import Layout from '@layouts/main.vue'
import axios from 'axios'
import { formatDecimal } from '@utils/format-modern'

const fuelTypes = ['Benzin', 'Diesel', 'Ethanol', 'Erdgas (CNG)', 'Elektro', 'Autogas (LPG)']
const seasons = { summer: 'Sommer', winter: 'Winter', 'all-season': 'Ganzjahr', other: 'Sonstige' }
const wheelPositions = ['front-left', 'front-right', 'rear-left', 'rear-right']

export default {
  page: { title: 'Fahrzeuge' },
  components: { Layout },
  data() { return { vehicles: [], tires: [], tireError: false, loading: true, error: '', search: '' } },
  computed: {
    visibleVehicles() {
      const q = this.search.trim().toLowerCase()
      return this.vehicles.filter(v => [v.nickname, v.registration, v.make, v.model].join(' ').toLowerCase().includes(q))
    },
    tiresByVehicle() {
      const grouped = {}
      for (const tire of this.tires) {
        if (!tire.vehicleId) continue
        if (!grouped[tire.vehicleId]) grouped[tire.vehicleId] = []
        grouped[tire.vehicleId].push(tire)
      }
      return grouped
    },
  },
  mounted() { this.load() },
  methods: {
    async load() {
      this.loading = true; this.error = ''; this.tireError = false
      try {
        const [vehicles, tires] = await Promise.all([
          this.$store.dispatch('vehicles/fetchVehicles'),
          axios.get('/api/tires').then(res => res.data).catch(() => { this.tireError = true; return [] }),
        ])
        this.vehicles = vehicles || []
        this.tires = Array.isArray(tires) ? tires : []
      } catch (_) { this.error = 'Fahrzeuge konnten nicht geladen werden.' }
      finally { this.loading = false }
    },
    fuelType(v) {
      return fuelTypes[v.fuelType] || (v.fuelTypeDetail && v.fuelTypeDetail.long) || 'Nicht angegeben'
    },
    formatDate(value) {
      const date = new Date(value)
      return value && Number.isFinite(date.getTime()) ? date.toLocaleDateString('de-DE') : '–'
    },
    history(v) {
      return [...(v.fillups || []).map(item => ({ ...item, source: 'Tankung' })),
        ...(v.expenses || []).map(item => ({ ...item, source: 'Ausgabe' }))]
        .filter(item => item.date && Number.isFinite(new Date(item.date).getTime()))
        .sort((a,b) => new Date(b.date) - new Date(a.date))
    },
    latestOdo(v) {
      // Use the most recent recorded reading, not an inferred live mileage.
      // Never label miles as kilometres or include the default zero of expenses.
      const item = this.history(v).find(item => item.odoReading !== null &&
        Number.isFinite(Number(item.odoReading)) && Number(item.odoReading) > 0 &&
        (item.distanceUnit === 0 || item.distanceUnit === 1))
      if (!item) return null
      return { value: formatDecimal(item.odoReading, 0) + (item.distanceUnit === 0 ? ' mi' : ' km'),
        detail: item.source + ' · ' + this.formatDate(item.date) }
    },
    latestActivity(v) {
      const item = this.history(v)[0]
      return item ? this.formatDate(item.date) : 'Noch keine Einträge'
    },
    engine(v) { return Number(v.engineSize) > 0 ? formatDecimal(v.engineSize, 0) + ' cm³' : '–' },
    tireSummary(v) {
      const items = this.tiresByVehicle[v.id] || []
      const mounted = items.filter(t => t.status === 'mounted')
      const counts = { summer: 0, winter: 0, 'all-season': 0, other: 0 }
      for (const tire of items) if (tire.status !== 'retired' && Object.prototype.hasOwnProperty.call(counts, tire.season)) counts[tire.season]++
      return {
        total: items.length,
        mounted: mounted.length,
        stored: items.filter(t => t.status === 'stored').length,
        retired: items.filter(t => t.status === 'retired').length,
        fittedPositions: Array.from(new Set(mounted.map(t => t.position).filter(p => wheelPositions.includes(p)))).length,
        types: Object.keys(counts).filter(key => counts[key]).map(key => seasons[key] + ': ' + counts[key]).join(' · '),
        sizes: Array.from(new Set(mounted.map(t => t.size).filter(Boolean))).join(' / '),
      }
    },
  },
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading">
      <div><p class="hm-eyebrow">FUHRPARK</p><h1 class="title hm-page-title">Fahrzeuge</h1><p class="hm-page-description">Deine Fahrzeuge, Kilometerstände, Fahrzeughistorie und zugeordneten Reifen auf einen Blick.</p></div>
      <b-button type="is-primary" tag="router-link" :to="{ name: 'vehicle-create' }">＋ Fahrzeug hinzufügen</b-button>
    </div>
    <b-notification v-if="error" type="is-danger" :closable="false">{{ error }} <b-button size="is-small" @click="load">Erneut laden</b-button></b-notification>
    <p v-if="loading" role="status">Fahrzeuge werden geladen …</p>
    <template v-else-if="!error">
      <div class="vehicle-list-toolbar"><b-field label="Fahrzeug suchen"><b-input v-model="search" placeholder="Name, Kennzeichen, Marke oder Modell" /></b-field><p class="help">{{ visibleVehicles.length }} von {{ vehicles.length }} Fahrzeugen</p></div>
      <p v-if="tireError" class="box vehicle-hint">Reifendaten konnten nicht geladen werden. Die übrigen Fahrzeugdaten sind trotzdem verfügbar.</p>
      <p v-if="!visibleVehicles.length" class="box">{{ vehicles.length ? 'Keine passenden Fahrzeuge.' : 'Noch keine Fahrzeuge angelegt. Klicke auf Fahrzeug hinzufügen.' }}</p>
      <div class="columns is-multiline">
        <div v-for="v in visibleVehicles" :key="v.id" class="column is-half">
          <article class="box fleet-card">
            <div class="fleet-header">
              <div class="fleet-symbol" aria-hidden="true">🚘</div>
              <div class="fleet-identity">
                <p class="fleet-label">{{ v.isOwner ? 'EIGENES FAHRZEUG' : 'FREIGEGEBEN' }}</p>
                <router-link class="fleet-name" :to="{ name: 'vehicle-detail', params: { vehicleId: v.id } }">{{ v.nickname || [v.make,v.model].filter(Boolean).join(' ') || 'Fahrzeug' }}</router-link>
                <p class="fleet-model">{{ [v.make,v.model].filter(Boolean).join(' ') || 'Marke / Modell nicht hinterlegt' }}</p>
              </div>
              <span v-if="v.registration" class="fleet-plate">{{ v.registration }}</span>
            </div>

            <div class="fleet-odo">
              <p class="fleet-label">LETZTER ERFASSTER KILOMETERSTAND</p>
              <strong>{{ latestOdo(v) ? latestOdo(v).value : 'Noch nicht erfasst' }}</strong>
              <p class="fleet-secondary">{{ latestOdo(v) ? latestOdo(v).detail : 'Ergänze eine Tankung oder Ausgabe mit Kilometerstand.' }}</p>
            </div>

            <div class="fleet-facts">
              <div><p class="fleet-label">Kraftstoff</p><strong>{{ fuelType(v) }}</strong></div>
              <div><p class="fleet-label">Baujahr</p><strong>{{ v.yearOfManufacture > 0 ? v.yearOfManufacture : '–' }}</strong></div>
              <div><p class="fleet-label">Hubraum</p><strong>{{ engine(v) }}</strong></div>
              <div><p class="fleet-label">Tankungen</p><strong>{{ (v.fillups || []).length }}</strong></div>
              <div><p class="fleet-label">Ausgaben</p><strong>{{ (v.expenses || []).length }}</strong></div>
              <div><p class="fleet-label">Letzter Eintrag</p><strong>{{ latestActivity(v) }}</strong></div>
            </div>

            <section class="fleet-tires">
              <div class="fleet-section-header">
                <h3>Reifen</h3>
                <router-link :to="{ name: 'tires', query: { vehicleId: v.id } }">Verwalten ›</router-link>
              </div>
              <p class="fleet-secondary">Meine zugeordneten Reifen</p>
              <p v-if="tireError" class="fleet-secondary">Reifendaten nicht verfügbar.</p>
              <template v-else-if="tireSummary(v).total">
                <div class="fleet-tire-counts">
                  <div><strong>{{ tireSummary(v).mounted }}</strong><span>Montiert</span></div>
                  <div><strong>{{ tireSummary(v).stored }}</strong><span>Eingelagert</span></div>
                  <div><strong>{{ tireSummary(v).retired }}</strong><span>Ausgemustert</span></div>
                </div>
                <p v-if="tireSummary(v).sizes"><span class="fleet-secondary">Montierte Größen:</span> <strong>{{ tireSummary(v).sizes }}</strong></p>
                <p v-if="tireSummary(v).types" class="fleet-secondary">{{ tireSummary(v).types }}</p>
                <p class="fleet-secondary">Erfasste Radpositionen: {{ tireSummary(v).fittedPositions }}/4</p>
              </template>
              <p v-else class="fleet-secondary">Noch keine Reifen zugeordnet.</p>
            </section>

            <div class="buttons fleet-actions">
              <b-button type="is-primary" tag="router-link" :to="{ name: 'vehicle-detail', params: { vehicleId: v.id } }">Details</b-button>
              <b-button class="fleet-secondary-button" tag="router-link" :to="{ name: 'tires', query: { vehicleId: v.id } }">Reifen</b-button>
              <b-button v-if="v.isOwner" class="fleet-secondary-button" tag="router-link" :to="{ name: 'vehicle-edit', params: { vehicleId: v.id } }">Bearbeiten</b-button>
            </div>
          </article>
        </div>
      </div>
    </template>
  </Layout>
</template>

<style scoped>
.vehicle-list-toolbar { margin-bottom:1rem; }
.vehicle-hint { color:var(--hm-text)!important; }
.fleet-card { height:100%; display:flex; flex-direction:column; gap:14px; overflow-wrap:anywhere; }
.fleet-card strong,.fleet-card h3 { color:var(--hm-text)!important; }
.fleet-header { display:flex; align-items:center; flex-wrap:wrap; gap:12px; }
.fleet-symbol { display:grid; place-items:center; flex:0 0 46px; height:46px; font-size:1.5rem; background:var(--hm-input); border:1px solid var(--hm-border); border-radius:12px; }
.fleet-identity { flex:1 1 120px; min-width:0; }
.fleet-label { color:var(--hm-muted)!important; font-size:.73rem; font-weight:700; letter-spacing:.035em; line-height:1.4; margin-bottom:4px; }
.fleet-name { font-size:clamp(1.25rem,2.2vw,1.65rem); line-height:1.2; font-weight:800; color:var(--hm-text)!important; display:block; }
.fleet-name:hover { text-decoration:underline; }
.fleet-model { color:var(--hm-muted); font-size:.88rem; margin-top:4px; }
.fleet-plate { border:1px solid var(--hm-border); border-radius:6px; padding:6px 10px; font-size:.91rem; font-weight:800; letter-spacing:.08em; color:var(--hm-text); background:var(--hm-input); }
.fleet-odo { padding:14px; border:1px solid var(--hm-border); border-left:4px solid var(--hm-accent); background:var(--hm-input); border-radius:10px; }
.fleet-odo > strong { display:block; font-size:clamp(1.35rem,2.5vw,1.9rem); line-height:1.25; font-variant-numeric:tabular-nums; }
.fleet-secondary { color:var(--hm-muted)!important; font-size:.82rem; line-height:1.5; }
.fleet-facts { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:8px; }
.fleet-facts > div { padding:10px; min-width:0; border:1px solid var(--hm-border); border-radius:8px; }
.fleet-facts strong { line-height:1.35; font-size:.95rem; }
.fleet-tires { background:var(--hm-input); border:1px solid var(--hm-border); border-radius:10px; padding:12px; }
.fleet-section-header { display:flex; flex-wrap:wrap; align-items:center; justify-content:space-between; gap:6px; }
.fleet-section-header h3 { font-size:1.05rem; font-weight:750; margin:0; }
.fleet-section-header a { color:var(--hm-text)!important; font-size:.82rem; text-decoration:underline; font-weight:650; }
.fleet-tire-counts { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:8px; margin:9px 0; }
.fleet-tire-counts > div { border-right:1px solid var(--hm-border); min-width:0; }
.fleet-tire-counts > div:last-child { border-right:0; }
.fleet-tire-counts strong { font-size:1.3rem; display:block; line-height:1.2; }
.fleet-tire-counts span { color:var(--hm-muted); font-size:.76rem; }
.fleet-actions { margin-top:auto; margin-bottom:0; padding-top:2px; }
.fleet-secondary-button { background:var(--hm-input)!important; border-color:var(--hm-border)!important; color:var(--hm-text)!important; }
@media(max-width:480px) { .fleet-facts { grid-template-columns:repeat(2,minmax(0,1fr)); } .fleet-header { align-items:flex-start; } }
</style>
