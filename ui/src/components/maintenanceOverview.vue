<script>
import axios from 'axios'
import { mapState } from 'vuex'

export default {
  props: {
    limit: { type: Number, default: 0 },
    heading: { type: String, default: 'Anstehende Wartungen' },
  },
  data() { return { tasks: [], loading: false, error: '' } },
  computed: {
    ...mapState('vehicles', ['vehicles']),
    upcoming() {
      const items = this.tasks.filter(task => !task.completedAt).map(task => ({
        ...task,
        status: this.status(task),
        sortValue: this.priority(task),
      }))
      items.sort((a, b) => a.sortValue - b.sortValue)
      return this.limit > 0 ? items.slice(0, this.limit) : items
    },
    count() { return this.tasks.filter(task => !task.completedAt).length },
  },
  mounted() { this.load() },
  methods: {
    async load() {
      this.loading = true
      this.error = ''
      try { const { data } = await axios.get('/api/maintenance'); this.tasks = data || [] }
      catch (error) { this.error = 'Wartungsdaten konnten nicht geladen werden.' }
      finally { this.loading = false }
    },
    label(id) {
      const vehicle = this.vehicles.find(v => v.id === id)
      return vehicle ? vehicle.nickname + ' – ' + vehicle.registration : 'Fahrzeug'
    },
    odo(id) {
      const vehicle = this.vehicles.find(v => v.id === id)
      if (!vehicle) return null
      const values = (vehicle.fillups || []).map(f => Number(f.odoReading)).filter(n => n > 0 && Number.isFinite(n))
      return values.length ? Math.max(...values) : null
    },
    status(task) {
      const items = []
      if (task.dueDate) {
        const date = new Date(task.dueDate)
        const days = Math.ceil((date.getTime() - Date.now()) / 86400000)
        items.push(days < 0 ? 'Termin überfällig' : days === 0 ? 'Heute fällig' : 'In ' + days + ' Tagen')
      }
      if (task.dueOdo !== null && task.dueOdo !== undefined) {
        const current = this.odo(task.vehicleId)
        items.push(current == null ? 'Kilometerstand noch unbekannt' : current >= task.dueOdo ? 'Kilometer erreicht' : 'Noch ' + (task.dueOdo - current).toLocaleString('de-DE') + ' km')
      }
      return items.join(' · ')
    },
    priority(task) {
      const dates = []
      if (task.dueDate) dates.push((new Date(task.dueDate).getTime() - Date.now()) / 86400000)
      if (task.dueOdo != null && this.odo(task.vehicleId) != null) dates.push((task.dueOdo - this.odo(task.vehicleId)) / 50)
      return dates.length ? Math.min(...dates) : 99999
    },
  },
}
</script>

<template>
  <section class="box hm-maintenance-overview">
    <div class="hm-page-heading hm-maintenance-heading">
      <div>
        <p class="hm-eyebrow">WARTUNGSPLANUNG</p>
        <h2 class="title is-5">{{ heading }}</h2>
        <p class="help">TÜV, Inspektion und Service – übersichtlich nach Fälligkeit.</p>
      </div>
      <b-button type="is-primary" size="is-small" tag="router-link" :to="{ name: 'maintenance', query: { neu: '1' } }">＋ Neue Wartung</b-button>
    </div>
    <p v-if="loading" class="help">Wartungen werden geladen…</p>
    <b-notification v-if="error" type="is-danger" :closable="false">{{ error }}</b-notification>
    <p v-if="!loading && !error && !count" class="help">Keine offenen Wartungen. Über „Neue Wartung“ kannst du TÜV, Ölwechsel oder Inspektionen planen.</p>
    <div v-for="task in upcoming" :key="task.id" class="hm-maintenance-row">
      <div class="hm-maintenance-row-text">
        <strong>{{ task.title }}</strong>
        <p class="help">{{ label(task.vehicleId) }}</p>
      </div>
      <span class="hm-maintenance-status">{{ task.status }}</span>
    </div>
    <router-link v-if="limit && count > limit" :to="{ name: 'maintenance' }" class="hm-maintenance-more">Alle {{ count }} Wartungen ansehen</router-link>
  </section>
</template>
