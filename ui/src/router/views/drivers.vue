<script>
import Layout from '@layouts/main.vue'
import axios from 'axios'
export default {
  page: { title: 'Fahrer' },
  components: { Layout },
  data() { return { drivers: [], name: '', notes: '', expiry: '', editingId: null, error: '' } },
  mounted() { this.load() },
  methods: {
    async load() { try { const response = await axios.get('/api/drivers'); this.drivers = response.data } catch (e) { this.error = 'Fahrer konnten nicht geladen werden.' } },
    startEdit(driver) { this.editingId = driver.id; this.name = driver.name; this.notes = driver.notes || ''; this.expiry = driver.licenseExpiry ? driver.licenseExpiry.substring(0, 10) : '' },
    reset() { this.name = ''; this.notes = ''; this.expiry = ''; this.editingId = null },
    async save() {
      this.error = ''
      try {
        const data = { name: this.name, notes: this.notes, licenseExpiry: this.expiry ? new Date(this.expiry + 'T12:00:00Z').toISOString() : null }
        if (this.editingId) await axios.put('/api/drivers/' + this.editingId, data)
        else await axios.post('/api/drivers', data)
        this.reset(); await this.load()
      } catch (e) { this.error = 'Fahrer konnte nicht gespeichert werden.' }
    },
    async remove(driver) { if (!window.confirm('Fahrer wirklich löschen?')) return; try { await axios.delete('/api/drivers/' + driver.id); await this.load() } catch (e) { this.error = 'Löschen fehlgeschlagen.' } },
    dateText(value) { return value ? new Date(value).toLocaleDateString('de-DE') : 'Nicht hinterlegt' },
    isExpired(value) { return value && new Date(value) < new Date() },
  },
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading"><div><p class="hm-eyebrow">VERWALTUNG</p><h1 class="title hm-page-title">Fahrer</h1><p class="hm-page-description">Fahrer unabhängig von Benutzerkonten verwalten. Keine Führerscheinnummer nötig.</p></div></div>
    <b-notification v-if="error" type="is-danger" :closable="false">{{ error }}</b-notification>
    <div class="box">
      <h2 class="title is-5">{{ editingId ? 'Fahrer bearbeiten' : 'Fahrer hinzufügen' }}</h2>
      <form @submit.prevent="save">
        <b-field label="Name"><b-input v-model="name" required maxlength="150" expanded /></b-field>
        <b-field label="Führerschein gültig bis (optional)"><input v-model="expiry" class="input" type="date" /></b-field>
        <b-field label="Notizen (optional)"><b-input v-model="notes" type="textarea" maxlength="2000" expanded /></b-field>
        <b-button native-type="submit" type="is-primary">{{ editingId ? 'Änderungen speichern' : 'Fahrer speichern' }}</b-button>
        <b-button v-if="editingId" type="is-light" @click="reset">Abbrechen</b-button>
      </form>
    </div>
    <div v-for="driver in drivers" :key="driver.id" class="box">
      <strong>{{ driver.name }}</strong><p>Führerschein gültig bis: <span :class="{ 'has-text-danger': isExpired(driver.licenseExpiry) }">{{ dateText(driver.licenseExpiry) }}</span></p>
      <p v-if="driver.notes">{{ driver.notes }}</p>
      <b-button size="is-small" @click="startEdit(driver)">Bearbeiten</b-button>
      <b-button size="is-small" type="is-danger is-light" @click="remove(driver)">Löschen</b-button>
    </div>
  </Layout>
</template>
