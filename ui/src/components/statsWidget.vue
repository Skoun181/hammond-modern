<script>
import { addDays, addMonths } from 'date-fns'
import { formatMoney } from '@utils/format-modern'
import { mapState } from 'vuex'

import axios from 'axios'
export default {
  props: {
    user: {
      type: Object,
      required: true,
    },
  },
  data: function() {
    return {
      dateRangeOptions: [
        { label: 'Diese Woche', value: 'this_week' },
        { label: 'Dieser Monat', value: 'this_month' },
        { label: 'Letzte 30 Tage', value: 'past_30_days' },
        { label: 'Letzte 3 Monate', value: 'past_3_months' },
        { label: 'Dieses Jahr', value: 'this_year' },
        { label: 'Gesamter Zeitraum', value: 'all_time' },
      ],
      dateRangeOption: 'past_30_days',
      stats: [],
      loadError: false,
    }
  },
  computed: {
    ...mapState('utils', ['isMobile']),
    summaryObject() {
      if (this.stats == null) {
        return [
          [
            {
              label: 'Gesamtausgaben',
              value: this.formatCurrency(0, this.user.currency),
            },
            {
              label: 'Tankkosten',
              value: `${this.formatCurrency(0, this.user.currency)} (0)`,
            },
            {
              label: 'Sonstige Ausgaben',
              value: `${this.formatCurrency(0, this.user.currency)} (0)`,
            },
          ],
        ]
      }
      return this.stats.map((x) => {
        return [
          {
            label: 'Gesamtausgaben',
            value: this.formatCurrency(x.expenditureTotal, x.currency),
          },
          {
            label: 'Tankkosten',
            value: `${this.formatCurrency(x.expenditureFillups, x.currency)} (${x.countFillups})`,
          },
          {
            label: 'Sonstige Ausgaben',
            value: `${this.formatCurrency(x.expenditureExpenses, x.currency)} (${x.countExpenses})`,
          },
        ]
      })
    },
  },
  watch: {
    dateRangeOption(newOne, old) {
      if (newOne === old) {
        return
      }
      this.getStats()
    },
  },
  mounted() {
    this.getStats()
  },
  methods: {
    formatCurrency(number, currencyCode) {
      if (!currencyCode) {
        currencyCode = (this.user && this.user.currency) || 'EUR'
      }
      return formatMoney(number, currencyCode)
    },
    getStats() {
      this.loadError = false
      axios
        .get('/api/me/stats', {
          params: {
            start: this.getStartDate(),
            end: new Date(),
          },
        })
        .then((data) => {
          this.stats = Array.isArray(data.data) ? data.data : []
        })
        .catch(() => { this.loadError = true; this.stats = [] })
    },
    getStartDate() {
      const toDate = new Date()
      switch (this.dateRangeOption) {
        case 'this_week':
          var currentDayOfWeek = toDate.getDay()
          var toSubtract = 0
          if (currentDayOfWeek === 0) {
            toSubtract = -6
          }
          if (currentDayOfWeek > 1) {
            toSubtract = -1 * (currentDayOfWeek - 1)
          }
          toDate.setHours(0, 0, 0, 0)
          return addDays(toDate, toSubtract)
        case 'this_month':
          return new Date(toDate.getFullYear(), toDate.getMonth(), 1)
        case 'past_30_days':
          return addDays(toDate, -30)
        case 'past_3_months':
          return addMonths(toDate, -3)
        case 'this_year':
          return new Date(toDate.getFullYear(), 0, 1)
        case 'all_time':
          return new Date(1969, 4, 20)
        default:
          return new Date(1969, 4, 20)
      }
    },
  },
}
</script>

<template>
  <div>
    <div class="columns">
      <div class="column" :class="isMobile ? 'has-text-centered' : ''"> <h2 class="title">Kostenübersicht – alle meine Fahrzeuge</h2></div>
      <div class="column">
        <b-select v-model="dateRangeOption" class="is-pulled-right is-medium">
          <option v-for="option in dateRangeOptions" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </b-select></div
      >
    </div>
    <b-notification v-if="loadError" type="is-danger" :closable="false">Statistiken konnten nicht geladen werden.</b-notification>
    <div v-for="(currencyLevel, index) in summaryObject" :key="index" class="modern-stat-grid">
      <div v-for="item in currencyLevel" :key="item.label" class="modern-stat-card">
        <div>
          <p class="heading">{{ item.label }}</p>
          <p class="title is-4">{{ item.value }}</p>
        </div>
      </div>
    </div>
  </div>
</template>
