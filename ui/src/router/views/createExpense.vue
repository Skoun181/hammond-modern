<script>
import Layout from '@layouts/main.vue'
import { mapState } from 'vuex'
import axios from 'axios'
import store from '@state/store'
import { format } from 'date-fns'

export default {
  page: {
    title: 'Create Expense',
  },
  components: { Layout },
  props: {
    vehicle: {
      type: Object,
      required: true,
    },
    expense: {
      type: Object,
      required: false,
      default: function() {
        return {}
      },
    },
  },
  data() {
    return {
      tryingToCreate: false,
      showMore: false,
      myVehicles: [],
      users: [],
      selectedVehicle: this.vehicle,
      expenseModel: this.expense,
      invoiceUploading: false,
      invoiceName: '',
    }
  },

  computed: {
    user() {
      return store.state.auth.currentUser
    },
    ...mapState('utils', ['isMobile']),
    ...mapState('users', ['me']),
    ...mapState('vehicles', ['fuelUnitMasters', 'fuelTypeMasters', 'vehicles']),
  },
  mounted() {
    this.myVehicles = this.vehicles
    this.selectedVehicle = this.vehicle
    this.fetchVehicleUsers()
    if (!this.expense.id) {
      this.expenseModel = this.getEmptyExpense()
      this.expenseModel.userId = this.me.id
    }
  },
  methods: {
    formatDate(date) {
      const finalFormat = this.me.dateFormat ? this.me.dateFormat : 'MM/dd/yyyy'
      return format(date, finalFormat)
    },
    getEmptyExpense() {
      return {
        vehicleId: this.selectedVehicle.id,
        amount: null,
        expenseType: 'Werkstatt / Reparatur',
        workshop: '',
        workPerformed: '',
        invoiceAttachmentId: '',
        odoReading: '',
        date: new Date(),
        comments: '',
        userId: '',
      }
    },
    fetchVehicleUsers() {
      store
        .dispatch('vehicles/fetchUsersByVehicleId', { vehicleId: this.selectedVehicle.id })
        .then((data) => {
          this.users = data
        })
        .catch((err) => console.log(err))
    },
    async downloadInvoice() {
      if (!this.expenseModel.invoiceAttachmentId) return
      try {
        const response = await axios.get('/api/invoices/' + this.expenseModel.invoiceAttachmentId + '/file', { responseType: 'blob' })
        const url = window.URL.createObjectURL(response.data)
        const a = document.createElement('a')
        a.href = url; a.download = this.invoiceName || 'werkstattrechnung'; document.body.appendChild(a); a.click(); a.remove()
        window.URL.revokeObjectURL(url)
      } catch (error) { this.$buefy.toast.open({ message: 'Rechnung konnte nicht geladen werden.', type: 'is-danger' }) }
    },
    async uploadInvoice(event) {
      const file = event.target.files && event.target.files[0]
      if (!file) return
      if (!['application/pdf', 'image/jpeg', 'image/png'].includes(file.type) || file.size > 10 * 1024 * 1024) {
        this.$buefy.toast.open({ message: 'Bitte PDF, JPG oder PNG bis 10 MB wählen.', type: 'is-danger' }); return
      }
      this.invoiceUploading = true
      try {
        const form = new FormData()
        form.append('file', file)
        const response = await axios.post('/api/invoices', form, { headers: { 'Content-Type': 'multipart/form-data' } })
        const attachment = Array.isArray(response.data) ? response.data[0] : response.data
        if (!attachment || !attachment.id) throw new Error('Keine Anhang-ID erhalten')
        this.expenseModel.invoiceAttachmentId = attachment.id
        this.invoiceName = file.name
      } catch (err) { this.$buefy.toast.open({ message: 'Rechnung konnte nicht hochgeladen werden.', type: 'is-danger' }) }
      finally { this.invoiceUploading = false }
    },
    createExpense() {
      this.tryingToCreate = true
      this.expenseModel.vehicleId = this.selectedVehicle.id
      //   this.expenseModel.userId = this.me.id

      if (this.expense.id) {
        axios
          .put(`/api/vehicles/${this.selectedVehicle.id}/expenses/${this.expense.id}`, this.expenseModel)
          .then((data) => {
            this.$buefy.toast.open({
              message: 'Expense Updated Successfully',
              type: 'is-success',
              duration: 3000,
            })
            this.expenseModel = this.getEmptyExpense()
          })
          .catch((ex) => {
            this.$buefy.toast.open({
              duration: 5000,
              message: ex,
              position: 'is-bottom',
              type: 'is-danger',
            })
          })
          .finally(() => {
            this.tryingToCreate = false
          })
      } else {
        axios
          .post(`/api/vehicles/${this.selectedVehicle.id}/expenses`, this.expenseModel)
          .then((data) => {
            this.$buefy.toast.open({
              message: 'Expense Created Successfully',
              type: 'is-success',
              duration: 3000,
            })
            this.expenseModel = this.getEmptyExpense()
          })
          .catch((ex) => {
            this.$buefy.toast.open({
              duration: 5000,
              message: ex,
              position: 'is-bottom',
              type: 'is-danger',
            })
          })
          .finally(() => {
            this.tryingToCreate = false
          })
      }
    },
  },
}
</script>

<template>
  <Layout>
    <div class="columns">
      <div class="column is-two-thirds">
        <h1 class="title">Werkstatt & Ausgaben</h1>
        <h1 class="subtitle">
          {{ [selectedVehicle.nickname, selectedVehicle.registration, selectedVehicle.make, selectedVehicle.model].join(' | ') }}
        </h1>
      </div>
    </div>
    <form @submit.prevent="createExpense">
      <b-field label="Fahrzeug">
        <b-select v-model="selectedVehicle" placeholder="Vehicle" required expanded :disabled="expense.id">
          <option v-for="option in myVehicles" :key="option.id" :value="option">
            {{ option.nickname }}
          </option>
        </b-select>
      </b-field>
      <b-field label="Erfasst von">
        <b-select v-model="expenseModel.userId" placeholder="User" required expanded :disabled="expense.id">
          <option v-for="option in users" :key="option.userId" :value="option.userId">
            {{ option.name }}
          </option>
        </b-select>
      </b-field>
      <b-field label="Datum der Ausgabe">
        <b-datepicker
          v-model="expenseModel.date"
          :date-formatter="formatDate"
          placeholder="Click to select..."
          icon="calendar"
          :max-date="new Date()"
        >
        </b-datepicker>
      </b-field>
      <b-field label="Werkstatt / Dienstleister">
        <b-input v-model="expenseModel.workshop" placeholder="z. B. Autohaus Müller" expanded></b-input>
      </b-field>
      <b-field label="Was wurde gemacht oder ersetzt?">
        <b-input v-model="expenseModel.workPerformed" type="textarea" placeholder="z. B. Bremsbeläge vorne erneuert" expanded></b-input>
      </b-field>
      <b-field label="Rechnung hochladen (PDF, JPG, PNG · maximal 10 MB)">
        <input type="file" accept=".pdf,.jpg,.jpeg,.png" :disabled="invoiceUploading" @change="uploadInvoice" />
        <p class="help">{{ invoiceUploading ? 'Wird hochgeladen…' : (invoiceName || (expenseModel.invoiceAttachmentId ? 'Rechnung hinterlegt' : 'Noch keine Datei ausgewählt')) }}</p>
        <a v-if="expenseModel.invoiceAttachmentId" href="#" @click.prevent="downloadInvoice">Rechnung herunterladen</a>
      </b-field>
      <b-field label="Art der Ausgabe*">
        <b-input v-model="expenseModel.expenseType" expanded required></b-input>
      </b-field>

      <b-field label="Rechnungsbetrag">
        <p class="control">
          <span class="button is-static">{{ me.currency }}</span>
        </p>
        <b-input v-model.number="expenseModel.amount" type="number" min="0" expanded step=".001" required></b-input>
      </b-field>
      <b-field label="Kilometerstand">
        <p class="control">
          <span class="button is-static">{{ me.distanceUnitDetail.short }}</span>
        </p>
        <b-input v-model.number="expenseModel.odoReading" type="number" min="0" expanded required></b-input>
      </b-field>

      <b-field>
        <b-switch v-model="showMore">Fill more details</b-switch>
      </b-field>
      <fieldset v-if="showMore">
        <b-field label="Zusätzliche Notizen">
          <b-input v-model="expenseModel.comments" type="textarea" expanded></b-input>
        </b-field>
      </fieldset>
      <br />
      <b-field>
        <b-button tag="input" native-type="submit" :disabled="tryingToCreate || invoiceUploading" type="is-primary" label="Ausgabe speichern" expanded> </b-button>
      </b-field>
    </form>
  </Layout>
</template>
