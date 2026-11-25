<script>
import axios from 'axios'
import { Chart, CategoryScale, LinearScale, LineElement, PointElement, Filler, Legend, Tooltip } from 'chart.js'
import { Line } from 'vue-chartjs'

Chart.register(CategoryScale, LinearScale, LineElement, PointElement, Filler, Legend, Tooltip)

export default {
  components: { Line },
  props: {
    vehicle: { type: Object, required: true },
    since: { type: [Date, String], default: '' },
    user: { type: Object, required: true },
  },
  data() {
    return { rows: [], request: null, options: { maintainAspectRatio: false } }
  },
  computed: {
    data() {
      return {
        labels: this.rows.map(row => row.date.slice(0, 10)),
        datasets: [{
          stepped: true,
          label: `Verbrauch (${this.user.distanceUnitDetail.short}/${this.vehicle.fuelUnitDetail.short})`,
          fill: true,
          data: this.rows.map(row => row.mileage),
        }],
      }
    },
  },
  watch: {
    since() { this.fetchMileage() },
  },
  mounted() {
    this.fetchMileage()
  },
  beforeUnmount() {
    this.request?.abort()
  },
  methods: {
    async fetchMileage() {
      this.request?.abort()
      this.request = new AbortController()
      try {
        const response = await axios.get(`/api/vehicles/${this.vehicle.id}/mileage`, {
          params: { since: this.since },
          signal: this.request.signal,
        })
        this.rows = response.data || []
      } catch (error) {
        if (!axios.isCancel(error)) this.rows = []
      }
    },
  },
}
</script>

<template>
  <Line :data="data" :options="options" />
</template>
