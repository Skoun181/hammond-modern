<script>
import Layout from '@layouts/main.vue'
import axios from 'axios'
export default {
  page: { title: 'Fahrer' },
  components: { Layout },
  data() { return { drivers: [], name: '', notes: '', expiry: '', editingId: null, showForm: false, loading: true, saving: false, search: '', error: '', success: '' } },
  computed: {
    filteredDrivers() {
      const q = this.search.trim().toLowerCase()
      return this.drivers.filter(d => [d.name, d.notes].join(' ').toLowerCase().includes(q))
    },
  },
  mounted() { this.load() },
  methods: {
    async load() {
      this.loading = true
      try { const response = await axios.get('/api/drivers'); this.drivers = response.data || [] }
      catch (_) { this.error = 'Fahrer konnten nicht geladen werden.' }
      finally { this.loading = false }
    },
    startEdit(driver) { this.editingId = driver.id; this.name = driver.name; this.notes = driver.notes || ''; this.expiry = driver.licenseExpiry ? driver.licenseExpiry.substring(0, 10) : ''; this.showForm = true; this.error = ''; this.success = '' },
    add() { this.reset(); this.showForm = true; this.error = ''; this.success = '' },
    reset() { this.name = ''; this.notes = ''; this.expiry = ''; this.editingId = null; this.showForm = false },
    async save() {
      if (this.saving) return
      this.error = ''; this.success = ''; this.saving = true
      try {
        const data = { name: this.name.trim(), notes: this.notes, licenseExpiry: this.expiry ? new Date(this.expiry + 'T12:00:00Z').toISOString() : null }
        if (this.editingId) await axios.put('/api/drivers/' + this.editingId, data)
        else await axios.post('/api/drivers', data)
        this.success = this.editingId ? 'Fahrer aktualisiert.' : 'Fahrer angelegt.'
        this.reset(); await this.load()
      } catch (e) { this.error = (e.response && e.response.data && e.response.data.error) || 'Fahrer konnte nicht gespeichert werden.' }
      finally { this.saving = false }
    },
    async remove(driver) {
      if (this.saving || !window.confirm('Fahrer wirklich löschen?')) return
      this.saving = true; this.error = ''; this.success = ''
      try { await axios.delete('/api/drivers/' + driver.id); if (this.editingId===driver.id)this.reset(); await this.load(); this.success = 'Fahrer gelöscht.' }
      catch (_) { this.error = 'Löschen fehlgeschlagen.' }
      finally { this.saving = false }
    },
    dateText(value) { return value ? new Date(value).toLocaleDateString('de-DE') : 'Nicht hinterlegt' },
    isExpired(value) { if(!value)return false; const d=new Date(value);d.setHours(23,59,59,999);return d<new Date() },
  },
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading">
      <div><p class="hm-eyebrow">VERWALTUNG</p><h1 class="title hm-page-title">Fahrer</h1><p class="hm-page-description">Deine angelegten Fahrer bearbeiten oder neue Fahrer erfassen. Keine Führerscheinnummer nötig.</p></div>
      <b-button type="is-primary" :disabled="saving" @click="add">＋ Fahrer hinzufügen</b-button>
    </div>
    <b-notification v-if="error" type="is-danger" :closable="true" @close="error=''">{{ error }}</b-notification>
    <b-notification v-if="success" type="is-success" :closable="true" @close="success=''">{{ success }}</b-notification>
    <div v-if="showForm" class="box">
      <h2 class="title is-5">{{ editingId ? 'Fahrer bearbeiten' : 'Fahrer hinzufügen' }}</h2>
      <form @submit.prevent="save">
        <b-field label="Name"><b-input v-model="name" required maxlength="150" expanded /></b-field>
        <b-field label="Führerschein gültig bis (optional)"><input v-model="expiry" class="input" type="date" /></b-field>
        <b-field label="Notizen (optional)"><b-input v-model="notes" type="textarea" maxlength="2000" expanded /></b-field>
        <div class="buttons"><b-button native-type="submit" type="is-primary" :loading="saving">{{ editingId ? 'Änderungen speichern' : 'Fahrer speichern' }}</b-button>
        <b-button :disabled="saving" @click="reset">Abbrechen</b-button></div>
      </form>
    </div>
    <b-field label="Fahrer suchen"><b-input v-model="search" placeholder="Name oder Notizen" /></b-field>
    <p v-if="loading" role="status">Fahrer werden geladen …</p>
    <p v-else-if="!filteredDrivers.length" class="box">{{ drivers.length ? 'Keine passenden Fahrer.' : 'Noch keine Fahrer angelegt.' }}</p>
    <div class="columns is-multiline">
      <div v-for="driver in filteredDrivers" :key="driver.id" class="column is-half">
        <div class="box">
          <h2 class="title is-5">{{ driver.name }}</h2><p>Führerschein gültig bis: <span :class="{ 'has-text-danger': isExpired(driver.licenseExpiry) }">{{ dateText(driver.licenseExpiry) }}</span></p>
          <p v-if="driver.notes" style="white-space:pre-wrap;overflow-wrap:anywhere">{{ driver.notes }}</p>
          <div class="buttons" style="margin-top:1rem"><b-button size="is-small" type="is-primary" :disabled="saving" @click="startEdit(driver)">Bearbeiten</b-button>
          <b-button size="is-small" type="is-danger is-light" :disabled="saving" @click="remove(driver)">Löschen</b-button></div>
        </div>
      </div>
    </div>
  </Layout>
</template>
