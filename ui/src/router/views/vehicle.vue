<script>
import Layout from '@layouts/main.vue'
import { parseAndFormatDate } from '@utils/format-date'
import { mapState } from 'vuex'
import { addDays, addMonths } from 'date-fns'
import axios from 'axios'
import { formatMoney, formatDecimal, normalizeCurrency } from '@utils/format-modern'
import store from '@state/store'
import ShareVehicle from '@components/shareVehicle.vue'
import MileageChart from '@components/mileageChart.vue'

export default {
  page() {
    return {
      title: this.vehicle.nickname,
      meta: [
        {
          name: 'description',
          content: `The vehicle profile for ${this.vehicle.nickname}.`,
        },
      ],
    }
  },
  components: { Layout, MileageChart },
  props: {
    vehicle: {
      type: Object,
      required: true,
    },
  },
  data: function() {
    return {
      fillups: [],
      expenses: [],
      attachments: [],
      showAttachmentForm: false,
      file: null,
      tryingToUpload: false,
      title: '',
      stats: null,
      users: [],
      drivers: [],
      dateRangeOptions: [
        { label: 'Diese Woche', value: 'this_week' },
        { label: 'Dieser Monat', value: 'this_month' },
        { label: 'Letzte 30 Tage', value: 'past_30_days' },
        { label: 'Letzte 3 Monate', value: 'past_3_months' },
        { label: 'Dieses Jahr', value: 'this_year' },
        { label: 'Gesamter Zeitraum', value: 'all_time' },
      ],
      dateRangeOption: 'past_30_days',
    }
  },
  computed: {
    ...mapState('users', ['me']),
    ...mapState('auth', ['currentUser']),
    ...mapState('utils', ['isMobile']),
    summaryObject() {
      if (this.stats == null) {
        return []
      }
      return this.stats.map((x) => {
        return [
          {
            label: 'Währung',
            value: normalizeCurrency(x.currency || (this.me && this.me.currency)),
          },
          {
            label: 'Gesamtausgaben',
            value: this.formatCurrency(x.expenditureTotal, x.currency),
          },
          {
            label: 'Tankkosten',
            value: `${this.formatCurrency(x.expenditureFillups, x.currency)} (${x.countFillups})`,
          },
          {
            label: 'Weitere Ausgaben',
            value: `${this.formatCurrency(x.expenditureExpenses, x.currency)} (${x.countExpenses})`,
          },
          {
            label: 'Ø Kosten je Tankung',
            value: `${this.formatCurrency(x.avgFillupCost, x.currency)}`,
          },
          {
            label: 'Ø Tankmenge',
            value: `${formatDecimal(x.avgFuelQty, 1)} ${this.vehicle.fuelUnitDetail.short}`,
          },
          {
            label: 'Ø Kraftstoffpreis',
            value: `${this.formatCurrency(x.avgFuelPrice, x.currency)} / ${this.vehicle.fuelUnitDetail.short}`,
          },
        ]
      })
    },
  },

  mounted() {
    this.fetchFillups()
    this.fetchExpenses()

    this.fetchAttachments()
    this.fetchVehicleStats()
    this.fetchVehicleUsers()
    axios.get('/api/drivers').then(response => { this.drivers = response.data }).catch(() => { this.drivers = [] })
  },
  methods: {
    driverName(id) { const driver = this.drivers.find(d => d.id === id); return driver ? driver.name : (id ? 'Gelöschter / fremder Fahrer' : 'Nicht angegeben') },
    async downloadInvoice(id) {
      if (!id) return
      try {
        const response = await axios.get('/api/invoices/' + id + '/file', { responseType: 'blob' })
        const url = window.URL.createObjectURL(response.data)
        const link = document.createElement('a'); link.href = url; link.download = 'werkstattrechnung'; document.body.appendChild(link); link.click(); link.remove(); window.URL.revokeObjectURL(url)
      } catch (error) { this.$buefy.toast.open({ message: 'Rechnung nicht verfügbar.', type: 'is-danger' }) }
    },
    fetchAttachments() {
      store
        .dispatch('vehicles/fetchAttachmentsByVehicleId', { vehicleId: this.vehicle.id })
        .then((data) => {
          this.attachments = data
        })
        .catch((err) => console.log(err))
    },
    fetchFillups() {
      axios
        .get(`/api/vehicles/${this.vehicle.id}/fillups`)
        .then((response) => {
          this.fillups = response.data
        })
        .catch((err) => console.log(err))
    },

    fetchExpenses() {
      axios
        .get(`/api/vehicles/${this.vehicle.id}/expenses`)
        .then((response) => {
          this.expenses = response.data
        })
        .catch((err) => console.log(err))
    },

    fetchVehicleStats() {
      store
        .dispatch('vehicles/fetchStatsByVehicleId', { vehicleId: this.vehicle.id })
        .then((data) => {
          this.stats = data
        })
        .catch((err) => console.log(err))
    },
    fetchVehicleUsers() {
      store
        .dispatch('vehicles/fetchUsersByVehicleId', { vehicleId: this.vehicle.id })
        .then((data) => {
          this.users = data
        })
        .catch((err) => console.log(err))
    },
    deleteFillup(fillupId) {
      var sure = confirm('This will delete this fillup. This step cannot be reversed. Are you sure?')
      if (sure) {
        store
          .dispatch('vehicles/deleteFillupById', { vehicleId: this.vehicle.id, fillupId: fillupId })
          .then((data) => {
            this.fetchVehicleStats()
            this.fetchFillups()
          })
          .catch((err) => console.log(err))
      }
    },
    deleteExpense(expenseId) {
      var sure = confirm('This will delete this expense. This step cannot be reversed. Are you sure?')
      if (sure) {
        store
          .dispatch('vehicles/deleteExpenseById', { vehicleId: this.vehicle.id, expenseId: expenseId })
          .then((data) => {
            this.fetchVehicleStats()
            this.fetchExpenses()
          })
          .catch((err) => console.log(err))
      }
    },
    deleteVehicle() {
      var sure = confirm(
        'This will delete all the expenses and fillups related with this vehicle as well. This step cannot be reversed. Are you sure?'
      )
      if (sure) {
        axios
          .delete(`/api/vehicles/${this.vehicle.id}`)
          .then((data) => {
            this.$buefy.toast.open({
              message: 'Vehicle Deleted Successfully',
              type: 'is-success',
              duration: 3000,
            })
            this.$router.push('/')
          })
          .catch((ex) => {
            this.$buefy.toast.open({
              duration: 5000,
              message: ex.message,
              position: 'is-bottom',
              type: 'is-danger',
            })
          })
          .finally(() => {})
      }
    },
    addAttachment() {
      if (this.file == null) {
        return
      }
      this.tryingToUpload = true

      const formData = new FormData()
      formData.append('file', this.file, this.file.name)
      formData.append('title', this.title)
      // const config = { headers: { 'Content-Type': 'multipart/form-data; boundary=' + formData._boundary } }
      fetch(`/api/vehicles/${this.vehicle.id}/attachments`, {
        method: 'POST',
        body: formData,
        headers: {
          Authorization: this.currentUser.token,
        },
      })
        .then((data) => {
          this.$buefy.toast.open({
            message: 'File uploaded Successfully',
            type: 'is-success',
            duration: 3000,
          })
          this.file = null
          this.title = ''
          this.fetchAttachments()
        })
        .catch((ex) => {
          this.$buefy.toast.open({
            duration: 5000,
            message: ex.message,
            position: 'is-bottom',
            type: 'is-danger',
          })
        })
        .finally(() => {
          this.tryingToUpload = false
        })
    },
    formatDate(date) {
      return parseAndFormatDate(date)
    },
    formatCurrency(number, currencyCode) {
      if (!currencyCode) {
        currencyCode = this.me.currency
      }
      return formatMoney(number, currencyCode)
    },
    formatLiter(value) { return formatDecimal(value, 1) },
    formatDecimal(value, digits = 0) { return formatDecimal(value, digits) },
    formatPrice(value, currency) {
      return formatDecimal(value, 3) + ' ' + (normalizeCurrency(currency || (this.me && this.me.currency)) === 'EUR' ? '€' : normalizeCurrency(currency || (this.me && this.me.currency)))
    },
    columnTdAttrs(row, column) {
      return null
    },
    hiddenDesktop(row, column) {
      return {
        class: 'is-hidden-desktop',
      }
    },
    hiddenMobile(row, column) {
      return {
        class: 'is-hidden-mobile',
      }
    },
    showShareVehicleModal() {
      this.$buefy.modal.open({
        parent: this,
        component: ShareVehicle,
        hasModalCard: false,
        props: { vehicle: this.vehicle },
        onCancel: (x) => {
          this.fetchVehicleUsers()
    axios.get('/api/drivers').then(response => { this.drivers = response.data }).catch(() => { this.drivers = [] })
        },
      })
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
  <Layout>
    <div class="columns box hm-vehicle-header">
      <div class="column is-two-thirds" :class="isMobile ? 'has-text-centered' : ''">
        <p class="title">{{ vehicle.nickname }} - {{ vehicle.registration }}</p>
        <p class="subtitle">
          {{ [vehicle.make, vehicle.model, vehicle.fuelTypeDetail.long].join(' | ') }}

          <template v-if="users.length > 1">
            | Shared with :
            {{
              users
                .map((x) => {
                  if (x.userId === me.id) {
                    return 'You'
                  } else {
                    return x.name
                  }
                })
                .join(', ')
            }}
          </template>
        </p>
      </div>
      <div class="column is-one-third buttons has-text-centered hm-vehicle-actions">
        <b-button type="is-primary" tag="router-link" :to="`/vehicles/${vehicle.id}/fillup`">Tankung hinzufügen</b-button>
        <b-button type="is-primary" tag="router-link" :to="`/vehicles/${vehicle.id}/expense`">Ausgabe hinzufügen</b-button>
        <b-button tag="router-link" :to="{name:'tires',query:{vehicleId:vehicle.id}}">Reifen verwalten</b-button>
        <b-button
          v-if="vehicle.isOwner"
          tag="router-link"
          title="Edit Vehicle"
          :to="{
            name: 'vehicle-edit',
            props: { vehicle: vehicle },
            params: { vehicleId: vehicle.id },
          }"
        >
          <b-icon pack="fas" icon="edit" type="is-info"> </b-icon
        ></b-button>
        <b-button v-if="vehicle.isOwner" title="Share vehicle" @click="showShareVehicleModal">
          <b-icon pack="fas" icon="user-friends" type="is-info"> </b-icon
        ></b-button>
        <b-button v-if="vehicle.isOwner" title="Delete Vehicle" @click="deleteVehicle">
          <b-icon pack="fas" icon="trash" type="is-danger"> </b-icon
        ></b-button>
      </div>
    </div>
    <div v-for="(currencyLevel, index) in summaryObject" :key="index" class="box hm-vehicle-stats">
      <div v-for="item in currencyLevel" :key="item.label" class="hm-vehicle-stat has-text-centered">
        <div>
          <p class="heading">{{ item.label }}</p>
          <p class="title is-4">{{ item.value }}</p>
        </div>
      </div>
    </div>
    <div class="box">
      <h1 class="title is-4">Tankvorgänge</h1>

      <b-table class="hm-vehicle-table" :data="fillups" hoverable mobile-cards paginated per-page="10">
        <b-table-column v-slot="props" field="date" label="Datum" :td-attrs="columnTdAttrs" sortable date>
          {{ formatDate(props.row.date) }}
        </b-table-column>
        <b-table-column v-slot="props" field="fuelSubType" label="Kraftstoffsorte" :td-attrs="columnTdAttrs">
          {{ props.row.fuelSubType }}
        </b-table-column>
        <b-table-column v-slot="props" field="fuelQuantity" label="Menge" :td-attrs="hiddenMobile" numeric>
          {{ formatLiter(props.row.fuelQuantity) }} {{ props.row.fuelUnitDetail.short }}
        </b-table-column>
        <b-table-column
          v-slot="props"
          field="perUnitPrice"
          :label="'Preis pro ' + vehicle.fuelUnitDetail.short"
          :td-attrs="hiddenMobile"
          numeric
          sortable
        >
          {{ formatPrice(props.row.perUnitPrice, props.row.currency) }}
        </b-table-column>
        <b-table-column v-slot="props" field="totalAmount" label="Gesamtbetrag" sortable numeric>
          <strong>{{ formatCurrency(props.row.totalAmount, props.row.currency) }}</strong>
        </b-table-column>
        <b-table-column v-slot="props" width="20" field="isTankFull" label="Vollgetankt" :td-attrs="hiddenMobile">
          <b-icon pack="fas" :icon="props.row.isTankFull ? 'check' : 'times'" type="is-info"> </b-icon>
        </b-table-column>
        <b-table-column v-slot="props" field="odoReading" label="Kilometerstand" :td-attrs="hiddenMobile" numeric>
          {{ formatDecimal(props.row.odoReading, 0) }} {{ me.distanceUnitDetail.short }}
        </b-table-column>
        <b-table-column v-slot="props" field="fillingStation" label="Tankstelle" :td-attrs="hiddenMobile">
          {{ `${props.row.fillingStation}` }}
        </b-table-column>
        <b-table-column v-slot="props" field="driverId" label="Getankt von" :td-attrs="hiddenMobile">
          {{ driverName(props.row.driverId) }}
        </b-table-column>
        <b-table-column v-slot="props" field="userId" label="Erfasst von" :td-attrs="hiddenMobile">
          {{ `${props.row.user.name}` }}
        </b-table-column>
        <b-table-column v-slot="props">
          <b-button
            type="is-ghost"
            tag="router-link"
            :to="{
              name: 'vehicle-edit-fillup',
              props: { fillup: props.row, vehicle: vehicle },
              params: { fillupId: props.row.id, id: vehicle.id },
            }"
          >
            <b-icon pack="fas" icon="edit" type="is-info"> </b-icon
          ></b-button>
          <b-button type="is-ghost" title="Delete this fillup" @click="deleteFillup(props.row.id)">
            <b-icon pack="fas" icon="trash" type="is-danger"> </b-icon
          ></b-button>
        </b-table-column>
        <template v-slot:empty> Noch keine Tankvorgänge</template>
      </b-table>
    </div>
    <br />
    <div class="box">
      <h1 class="title is-4">Werkstatt & Ausgaben</h1>

      <b-table class="hm-vehicle-table" :data="expenses" hoverable mobile-cards paginated per-page="10">
        <b-table-column v-slot="props" field="date" label="Datum" :td-attrs="columnTdAttrs" date>
          {{ formatDate(props.row.date) }}
        </b-table-column>
        <b-table-column v-slot="props" field="expenseType" label="Ausgabenart">
          {{ `${props.row.expenseType}` }}
        </b-table-column>

        <b-table-column v-slot="props" field="workshop" label="Werkstatt">
          {{ props.row.workshop || '–' }}
        </b-table-column>
        <b-table-column v-slot="props" field="workPerformed" label="Durchgeführte Arbeiten">
          {{ props.row.workPerformed || '–' }}
        </b-table-column>
        <b-table-column v-slot="props" field="invoiceAttachmentId" label="Rechnung">
          <a v-if="props.row.invoiceAttachmentId" href="#" @click.prevent="downloadInvoice(props.row.invoiceAttachmentId)">Herunterladen</a>
          <span v-else>–</span>
        </b-table-column>
        <b-table-column v-slot="props" field="amount" label="Gesamtbetrag" sortable numeric>
          {{ `${formatCurrency(props.row.amount, props.row.currency)}` }}
        </b-table-column>

        <b-table-column v-slot="props" field="odoReading" label="Kilometerstand" :td-attrs="columnTdAttrs" numeric>
          {{ formatDecimal(props.row.odoReading, 0) }} {{ me.distanceUnitDetail.short }}
        </b-table-column>

        <b-table-column v-slot="props" field="userId" label="Erfasst von" :td-attrs="columnTdAttrs">
          {{ `${props.row.user.name}` }}
        </b-table-column>
        <b-table-column v-slot="props">
          <b-button
            type="is-ghost"
            tag="router-link"
            :to="{
              name: 'vehicle-edit-expense',
              props: { expense: props.row, vehicle: vehicle },
              params: { expenseId: props.row.id, id: vehicle.id },
            }"
          >
            <b-icon pack="fas" icon="edit" type="is-info"> </b-icon
          ></b-button>
          <b-button type="is-ghost" title="Delete this expense" @click="deleteExpense(props.row.id)">
            <b-icon pack="fas" icon="trash" type="is-danger"> </b-icon
          ></b-button>
        </b-table-column>
        <template v-slot:empty> No Expenses so far</template>
      </b-table>
    </div>
    <br />
    <div class="box">
      <div class="columns">
        <div class="column is-three-quarters"> <h1 class="title is-4">Anhänge</h1></div>
        <div class="column buttons">
          <b-button type="is-primary" @click="showAttachmentForm = true">
            Add Attachment
          </b-button>
        </div>
      </div>
      <div v-if="showAttachmentForm" class="box">
        <div class="columns">
          <div class="column"></div>
          <div class="column is-two-thirds">
            <form @submit.prevent="addAttachment">
              <b-field :grouped="!isMobile">
                <b-field class="file is-primary" :class="{ 'has-name': !!file }">
                  <b-upload v-model="file" class="file-label" required>
                    <span class="file-cta">
                      <b-icon class="file-icon" icon="upload"></b-icon>
                      <span class="file-label">Choose File</span>
                    </span>
                    <span v-if="file" class="file-name" :class="isMobile ? 'file-name-mobile' : 'file-name-desktop'">
                      {{ file.name }}
                    </span>
                  </b-upload>
                </b-field>
                <b-field>
                  <b-input v-model="title" required placeholder="Label for this file"></b-input>
                </b-field>

                <b-field class="buttons">
                  <b-button tag="input" native-type="submit" :disabled="tryingToUpload" type="is-primary" label="Upload File" value="Upload File">
                  </b-button>
                  <b-button
                    tag="input"
                    native-type="submit"
                    :disabled="tryingToUpload"
                    type="is-danger"
                    label="Upload File"
                    value="Cancel"
                    @click="showAttachmentForm = false"
                  >
                  </b-button>
                </b-field>
              </b-field>
            </form>
          </div>
        </div>
      </div>

      <b-table class="hm-vehicle-table" :data="attachments" hoverable mobile-cards>
        <b-table-column v-slot="props" field="title" label="Title" :td-attrs="columnTdAttrs">
          {{ `${props.row.title}` }}
        </b-table-column>

        <b-table-column v-slot="props" field="originalName" label="Name" :td-attrs="columnTdAttrs">
          {{ `${props.row.originalName}` }}
        </b-table-column>
        <b-table-column v-slot="props" field="id" label="Download" :td-attrs="columnTdAttrs">
          <b-button tag="a" :href="`/api/attachments/${props.row.id}/file?access_token=${currentUser.token}`" :download="props.row.originalName">
            <b-icon type="is-primary" icon="download"></b-icon>
          </b-button>
        </b-table-column>
        <template v-slot:empty> No Attachments so far</template>
      </b-table>
    </div>
    <div class="box">
      <div class="hm-stats-heading">
        <h1 class="title is-4">Statistiken</h1>
        <div>
          <b-select v-model="dateRangeOption" class="hm-stats-period">
            <option v-for="option in dateRangeOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </b-select></div
        >
      </div>
      <MileageChart :vehicle="vehicle" :since="getStartDate()" :user="me" :height="300" />
    </div>
  </Layout>
</template>
