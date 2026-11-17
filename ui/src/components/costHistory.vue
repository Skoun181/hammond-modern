<script>
import { mapState } from 'vuex'
import { formatMoney, formatDecimal } from '@utils/format-modern'

export default {
  computed: {
    ...mapState('vehicles', ['vehicles']),
    ...mapState('users', ['me']),
    months() {
      const today = new Date()
      const months = []
      for (let i = 5; i >= 0; i--) {
        const date = new Date(today.getFullYear(), today.getMonth() - i, 1)
        months.push({
          key: date.getFullYear() + '-' + String(date.getMonth() + 1).padStart(2, '0'),
          label: date.toLocaleDateString('de-DE', { month: 'short', year: '2-digit' }),
          fuel: 0, expenses: 0,
        })
      }
      ;(this.vehicles || []).forEach(vehicle => {
        ;(vehicle.fillups || []).forEach(item => {
          const date = new Date(item.date)
          if (isNaN(date.getTime())) return
          const key = date.getFullYear() + '-' + String(date.getMonth() + 1).padStart(2, '0')
          const month = months.find(m => m.key === key)
          if (month) month.fuel += Number(item.totalAmount) || 0
        })
        ;(vehicle.expenses || []).forEach(item => {
          const date = new Date(item.date)
          if (isNaN(date.getTime())) return
          const key = date.getFullYear() + '-' + String(date.getMonth() + 1).padStart(2, '0')
          const month = months.find(m => m.key === key)
          if (month) month.expenses += Number(item.amount) || 0
        })
      })
      return months.map(month => ({ ...month, total: month.fuel + month.expenses }))
    },
    maxTotal() { return Math.max(1, ...this.months.map(m => m.total)) },
    totalCost() { return this.months.reduce((sum, m) => sum + m.total, 0) },
    perVehicle() {
      return (this.vehicles || []).map(v => {
        const fillups = v.fillups || []
        const expenses = v.expenses || []
        const total = fillups.reduce((s, f) => s + (Number(f.totalAmount) || 0), 0) +
          expenses.reduce((s, e) => s + (Number(e.amount) || 0), 0)
        const readings = fillups.concat(expenses).map(f => Number(f.odoReading))
          .filter(n => Number.isFinite(n) && n > 0)
        const distance = readings.length >= 2 ? Math.max(...readings) - Math.min(...readings) : 0
        return { id: v.id, nickname: v.nickname, total, distance, costPerKm: distance > 0 ? total / distance : null }
      })
    },
  },
  methods: {
    money(n) { return formatMoney(n, (this.me && this.me.currency) || 'EUR') },
    decimal(n) { return formatDecimal(n) },
    height(value) { return Math.max(0, value / this.maxTotal * 100) + '%' },
  },
}
</script>

<template>
  <section class="box">
    <div class="hm-page-heading">
      <div>
        <h2 class="title is-5">Kostenentwicklung – letzte 6 Monate</h2>
        <p class="hm-page-description">Tankkosten und sonstige Ausgaben aller angezeigten Fahrzeuge.</p>
      </div>
      <strong>{{ money(totalCost) }}</strong>
    </div>
    <div class="hm-cost-chart" role="img" aria-label="Monatliche Tankkosten und sonstige Ausgaben">
      <div v-for="month in months" :key="month.key" class="hm-cost-column">
        <div class="hm-cost-value">{{ money(month.total) }}</div>
        <div class="hm-cost-bar-area">
          <div class="hm-cost-bar" :style="{ height: height(month.total) }">
            <div class="hm-cost-expense" :style="{ flexGrow: month.expenses }"></div>
            <div class="hm-cost-fuel" :style="{ flexGrow: month.fuel }"></div>
          </div>
        </div>
        <div class="hm-cost-month">{{ month.label }}</div>
      </div>
    </div>
    <p class="help">Orange = Tankungen · Violett = sonstige Ausgaben. Angezeigt werden gespeicherte Beträge im jeweiligen Monat.</p>
    <h3 class="title is-6">Kosten je Fahrzeug</h3>
    <div class="columns is-multiline">
      <div v-for="v in perVehicle" :key="v.id" class="column is-6">
        <div class="modern-stat-card">
          <strong>{{ v.nickname }}</strong>
          <p>Erfasste Gesamtkosten: {{ money(v.total) }}</p>
          <p v-if="v.costPerKm !== null">Kosten je erfasstem km: {{ money(v.costPerKm) }}/km</p>
          <p v-else class="help">€/km benötigt mindestens zwei unterschiedliche Kilometerstände.</p>
          <p class="help">Grundlage: alle gespeicherten Kosten, geteilt durch die Spanne der erfassten Kilometerstände. Keine vollständigen Lebenszykluskosten.</p>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hm-cost-chart { display:flex; gap:8px; min-height:245px; margin:18px 0; align-items:stretch; overflow-x:auto; }
.hm-cost-column { flex:1; min-width:65px; display:flex; flex-direction:column; align-items:center; }
.hm-cost-value { font-size:.72rem; color:var(--hm-text); white-space:nowrap; }
.hm-cost-bar-area { height:170px; width:100%; display:flex; justify-content:center; align-items:flex-end; border-bottom:1px solid var(--hm-border); }
.hm-cost-bar { width:65%; min-height:2px; display:flex; flex-direction:column; justify-content:flex-end; overflow:hidden; border-radius:5px 5px 0 0; }
.hm-cost-expense { background:#8b5cf6; min-height:0; }
.hm-cost-fuel { background:#f59e0b; min-height:0; }
.hm-cost-month { margin-top:8px; font-size:.72rem; color:var(--hm-muted); }
</style>
