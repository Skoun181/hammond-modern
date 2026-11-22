<script>
import axios from 'axios'
import { mapState } from 'vuex'

export default {
  props: {
    limit: { type: Number, default: 0 },
    heading: { type: String, default: 'Anstehende Wartungen' },
    showActions: { type: Boolean, default: false },
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
    urgent() { return this.tasks.filter(t => ['overdue','due'].includes(t.severity)).length },
    soon() { return this.tasks.filter(t => t.severity === 'soon').length },
  },
  mounted() { this.load(); this.refreshTimer = setInterval(() => { if (!this.loading) this.load() }, 60000) },
  beforeDestroy() { clearInterval(this.refreshTimer) },
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
    status(task) {
      const items=[]
      if(task.daysLeft!=null)items.push(task.daysLeft<0?'Seit '+(-task.daysLeft)+' Tagen überfällig':task.daysLeft===0?'Heute fällig':'In '+task.daysLeft+' Tagen')
      if(task.dueOdo!=null)items.push(task.kmLeft==null?'km-Stand unbekannt':task.kmLeft<0?(-task.kmLeft).toLocaleString('de-DE')+' km überfällig':task.kmLeft===0?'Kilometer erreicht':'Noch '+task.kmLeft.toLocaleString('de-DE')+' km')
      return items.join(' · ')
    },
    priority(task) {
      const group={overdue:0,due:1,soon:2,planned:3}[task.severity]
      const values=[]
      if(task.daysLeft!=null)values.push(task.daysLeft)
      if(task.kmLeft!=null)values.push(task.kmLeft/50)
      return (group==null?4:group)*1000000+Math.max(-99999,Math.min(99999,values.length?Math.min(...values):99999))
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
      <b-button type="is-primary" size="is-small" tag="router-link" :to="{ name: 'maintenance', query: { neu: '1' } }">＋ Neue Wartung hinzufügen</b-button>
    </div>
    <div v-if="urgent || soon" class="maintenance-reminder"><strong>{{ urgent }} fällig / überfällig</strong> · {{ soon }} demnächst fällig</div>
    <p v-if="loading" class="help">Wartungen werden geladen…</p>
    <b-notification v-if="error" type="is-danger" :closable="false">{{ error }}</b-notification>
    <p v-if="!loading && !error && !count" class="help">Keine offenen Wartungen. Über „Neue Wartung“ kannst du TÜV, Ölwechsel oder Inspektionen planen.</p>
    <div v-for="task in upcoming" :key="task.id" class="hm-maintenance-row">
      <div class="hm-maintenance-row-text">
        <strong>{{ task.title }}</strong>
        <p class="help">{{ label(task.vehicleId) }}</p>
      </div>
      <span class="hm-maintenance-status" :class="'maintenance-'+task.severity">{{ ({overdue:'Überfällig',due:'Fällig',soon:'Demnächst',planned:'Geplant'})[task.severity] }} · {{ task.status }}</span>
      <div v-if="showActions" class="hm-maintenance-actions">
        <b-button size="is-small" @click="$emit('edit', task)">Bearbeiten</b-button>
        <b-button size="is-small" type="is-primary" @click="$emit('complete', task)">Erledigt</b-button>
        <b-button size="is-small" type="is-danger" outlined @click="$emit('remove', task)">Löschen</b-button>
      </div>
    </div>
    <router-link v-if="limit && count > limit" :to="{ name: 'maintenance' }" class="hm-maintenance-more">Alle {{ count }} Wartungen ansehen</router-link>
  </section>
</template>

<style scoped>
.maintenance-reminder{border:1px solid var(--hm-border);border-left:4px solid var(--hm-accent);background:var(--hm-input);padding:10px 12px;border-radius:8px;margin-bottom:10px;color:var(--hm-text)}
.maintenance-reminder strong{color:var(--hm-text)!important}
.maintenance-overdue,.maintenance-due{color:var(--hm-text)!important;text-decoration:underline;text-decoration-color:var(--hm-accent)}
.maintenance-planned{color:var(--hm-muted)!important}
</style>
