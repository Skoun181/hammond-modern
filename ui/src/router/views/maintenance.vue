<script>
import Layout from '@layouts/main.vue'
import MaintenanceOverview from '@components/maintenanceOverview.vue'
import axios from 'axios'
import { mapState } from 'vuex'

export default {
  page: { title: 'Wartungscenter' },
  components: { Layout, MaintenanceOverview },
  data() {
    return {
      tasks: [],
      loading: false,
      saving: false,
      error: '',
      editingId: '',
      showForm: false,
      overviewKey: 0,
      form: { vehicleId: '', title: '', notes: '', dueDate: '', dueOdo: '', completed: false },
    }
  },
  computed: {
    ...mapState('vehicles', ['vehicles']),
    activeTasks() { return this.tasks.filter(t => !t.completedAt) },
    completedTasks() { return this.tasks.filter(t => !!t.completedAt) },
  },
  watch: {
    '$route.query.neu': function(value) {
      if (value === '1') {
        this.editingId = ''
        this.showForm = true
      }
    },
  },
  mounted() {
    this.showForm = this.$route.query.neu === '1'
    this.$store.dispatch('vehicles/fetchVehicles').then(() => {
      if (this.vehicles.length && !this.form.vehicleId) this.form.vehicleId = this.vehicles[0].id
    }).catch(() => { this.error = 'Fahrzeuge konnten nicht geladen werden.' })
    this.loadTasks()
  },
  methods: {
    async loadTasks() {
      this.loading = true
      try { const response = await axios.get('/api/maintenance'); this.tasks = response.data || [] }
      catch (error) { this.error = 'Wartungsplanung konnte nicht geladen werden.' }
      finally { this.loading = false }
    },
    vehicleName(id) {
      const vehicle = this.vehicles.find(v => v.id === id)
      return vehicle ? vehicle.nickname : 'Fahrzeug'
    },
    currentOdo(id) {
      const vehicle = this.vehicles.find(v => v.id === id)
      const readings = vehicle ? (vehicle.fillups || []).map(f => Number(f.odoReading)).filter(n => Number.isFinite(n) && n > 0) : []
      return readings.length ? Math.max(...readings) : null
    },
    status(task) {
      if (task.completedAt) return 'Erledigt'
      const states = []
      if (task.dueDate) {
        const days = Math.ceil((new Date(task.dueDate).getTime() - Date.now()) / 86400000)
        if (days < 0) states.push('Datum überfällig')
        else states.push('in ' + days + ' Tagen')
      }
      if (task.dueOdo !== null && task.dueOdo !== undefined) {
        const odo = this.currentOdo(task.vehicleId)
        if (odo === null) states.push('km-Stand noch unbekannt')
        else if (odo >= task.dueOdo) states.push('Kilometer erreicht')
        else states.push('noch ' + (task.dueOdo - odo).toLocaleString('de-DE') + ' km')
      }
      return states.join(' · ')
    },
    resetForm() {
      this.showForm = false
      if (this.$route.query.neu) this.$router.replace({ name: 'maintenance', query: {} })
      this.editingId = ''
      this.form = { vehicleId: this.vehicles.length ? this.vehicles[0].id : '', title: '', notes: '', dueDate: '', dueOdo: '', completed: false }
    },
    edit(task) {
      this.showForm = true
      this.editingId = task.id
      this.form = {
        vehicleId: task.vehicleId, title: task.title, notes: task.notes || '',
        dueDate: task.dueDate ? task.dueDate.slice(0, 10) : '',
        dueOdo: task.dueOdo === null || task.dueOdo === undefined ? '' : task.dueOdo,
        completed: !!task.completedAt,
      }
    },
    async save() {
      this.error = ''
      if (!this.form.vehicleId || !this.form.title.trim() || (!this.form.dueDate && this.form.dueOdo === '')) {
        this.error = 'Bitte Fahrzeug, Titel und mindestens ein Fälligkeitskriterium angeben.'
        return
      }
      this.saving = true
      const payload = {
        vehicleId: this.form.vehicleId, title: this.form.title.trim(), notes: this.form.notes,
        dueDate: this.form.dueDate ? this.form.dueDate + 'T12:00:00Z' : null,
        dueOdo: this.form.dueOdo === '' ? null : Number(this.form.dueOdo),
        completed: !!this.form.completed,
      }
      try {
        if (this.editingId) await axios.put('/api/maintenance/' + this.editingId, payload)
        else await axios.post('/api/maintenance', payload)
        this.resetForm()
        await this.loadTasks()
        this.overviewKey += 1
      } catch (error) {
        this.error = (error.response && error.response.data && error.response.data.error) || 'Wartung konnte nicht gespeichert werden.'
      } finally { this.saving = false }
    },
    async complete(task) {
      this.edit(task)
      this.form.completed = true
      await this.save()
    },
    async remove(task) {
      if (!window.confirm('Wartung wirklich löschen?')) return
      try { await axios.delete('/api/maintenance/' + task.id); await this.loadTasks(); this.overviewKey += 1 }
      catch (error) { this.error = 'Wartung konnte nicht gelöscht werden.' }
    },
  },
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading">
      <div>
        <p class="hm-eyebrow">FAHRZEUGPFLEGE</p>
        <h1 class="title hm-page-title">Wartungscenter</h1>
        <p class="hm-page-description">TÜV, Ölwechsel, Inspektion und andere Termine planen – noch ohne Benachrichtigungen.</p>
      </div>
    </div>
    <b-notification v-if="error" type="is-danger" :closable="false">{{ error }}</b-notification>
    <MaintenanceOverview :key="overviewKey" heading="Was steht als Nächstes an?" />
    <div v-if="showForm" id="maintenance-form" class="box">
      <h2 class="title is-5">{{ editingId ? 'Wartung bearbeiten' : 'Neue Wartung planen' }}</h2>
      <form @submit.prevent="save">
        <b-field label="Fahrzeug">
          <b-select v-model="form.vehicleId" expanded required>
            <option v-for="v in vehicles" :key="v.id" :value="v.id">{{ v.nickname }} – {{ v.registration }}</option>
          </b-select>
        </b-field>
        <b-field label="Wartung / Termin">
          <b-input v-model="form.title" maxlength="120" required placeholder="z. B. HU / TÜV, Ölwechsel, Inspektion" expanded />
        </b-field>
        <div class="columns">
          <div class="column">
            <b-field label="Fällig am (optional)">
              <b-input v-model="form.dueDate" type="date" expanded />
            </b-field>
          </div>
          <div class="column">
            <b-field label="Fällig bei Kilometerstand (optional)">
              <b-input v-model="form.dueOdo" type="number" min="0" expanded placeholder="z. B. 75000" />
            </b-field>
          </div>
        </div>
        <b-field label="Notizen">
          <b-input v-model="form.notes" type="textarea" maxlength="2000" expanded />
        </b-field>
        <b-field v-if="editingId">
          <b-checkbox v-model="form.completed">Als erledigt markieren</b-checkbox>
        </b-field>
        <b-button type="is-primary" native-type="submit" :loading="saving">{{ editingId ? 'Änderungen speichern' : 'Wartung speichern' }}</b-button>
        <b-button v-if="editingId" @click="resetForm">Abbrechen</b-button>
      </form>
    </div>
    <h2 class="title is-5">Offene Wartungen bearbeiten ({{ activeTasks.length }})</h2>
    <p v-if="loading">Wartungen werden geladen…</p>
    <p v-else-if="!activeTasks.length" class="help">Noch keine offenen Wartungen vorhanden.</p>
    <div v-for="task in activeTasks" :key="task.id" class="box">
      <div class="columns is-vcentered">
        <div class="column">
          <strong>{{ task.title }}</strong> · {{ vehicleName(task.vehicleId) }}
          <p class="help">{{ status(task) }}</p>
          <p v-if="task.notes">{{ task.notes }}</p>
        </div>
        <div class="column is-narrow">
          <b-button size="is-small" @click="edit(task)">Bearbeiten</b-button>
          <b-button size="is-small" type="is-primary" @click="complete(task)">Erledigt</b-button>
          <b-button size="is-small" type="is-danger" outlined @click="remove(task)">Löschen</b-button>
        </div>
      </div>
    </div>
    <h2 v-if="completedTasks.length" class="title is-5">Erledigt ({{ completedTasks.length }})</h2>
    <div v-for="task in completedTasks" :key="task.id" class="box">
      <strong>{{ task.title }}</strong> · {{ vehicleName(task.vehicleId) }}
      <p class="help">Erledigt</p>
      <b-button size="is-small" @click="edit(task)">Bearbeiten</b-button>
      <b-button size="is-small" type="is-danger" outlined @click="remove(task)">Löschen</b-button>
    </div>
  </Layout>
</template>
