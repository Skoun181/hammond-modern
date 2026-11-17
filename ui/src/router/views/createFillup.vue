<script>
import Layout from '@layouts/main.vue'
import QuickEntryDisplay from '@components/quickEntryDisplay.vue'
import store from '@state/store'

import { mapState } from 'vuex'
import axios from 'axios'
import { round } from 'lodash'
import { format } from 'date-fns'

export default {
  page: {
    title: 'Tankung hinzufügen',
  },
  components: { Layout, QuickEntryDisplay },
  props: {
    vehicle: {
      type: Object,
      required: true,
    },
    fillup: {
      type: Object,
      required: false,
      default: function() {
        return {}
      },
    },
  },
  data() {
    return {
      authError: null,
      tryingToCreate: false,
      showMore: false,
      quickEntry: null,
      myVehicles: [],
      users: [],
      fuelSubTypes: [],
      drivers: [],
      stationPostal: '',
      stations: [],
      stationLoading: false,
      stationError: '',
      selectedVehicle: this.vehicle,
      fillupModel: this.fillup,
      processQuickEntry: false,
    }
  },

  computed: {
    user() {
      return store.state.auth.currentUser
    },
    ...mapState('users', ['me']),
    ...mapState('vehicles', ['fuelUnitMasters', 'fuelTypeMasters', 'vehicles']),
    filteredFuelSubtypes() {
      if (!this.fillupModel.fuelSubType) {
        return this.fuelSubTypes
      }
      return this.fuelSubTypes.filter((option) => {
        return option.toLowerCase().indexOf(this.fillupModel.fuelSubType.toLowerCase()) >= 0
      })
    },
  },
  watch: {
    'fillupModel.fuelQuantity': function(old, newOne) {
      this.fillupModel.totalAmount = round(this.fillupModel.fuelQuantity * this.fillupModel.perUnitPrice, 2)
    },
    'fillupModel.perUnitPrice': function(old, newOne) {
      this.fillupModel.totalAmount = round(this.fillupModel.fuelQuantity * this.fillupModel.perUnitPrice, 2)
    },
    quickEntry: function(newOne, old) {
      if (old == null && newOne !== null) {
        this.processQuickEntry = true
      }
    },
  },
  mounted() {
    this.myVehicles = this.vehicles
    this.selectedVehicle = this.vehicle
    this.fetchVehicleUsers()
    axios.get('/api/drivers').then((r) => { this.drivers = r.data }).catch(() => { this.drivers = [] })
    this.fetchVehicleFuelSubTypes()
    if (!this.fillup.id) {
      this.fillupModel = this.getEmptyFillup()
      this.fillupModel.userId = this.me.id
    }
  },
  methods: {
    async searchStations() {
      if (!/^\d{5}$/.test(this.stationPostal)) { this.stationError = 'Bitte eine fünfstellige PLZ eingeben.'; return }
      this.stationLoading = true
      this.stationError = ''
      try { const response = await axios.get('/api/stations', { params: { plz: this.stationPostal } }); this.stations = response.data }
      catch (error) { this.stationError = (error.response && error.response.data && error.response.data.error) || 'Tankstellensuche nicht verfügbar.' }
      finally { this.stationLoading = false }
    },
    chooseStation(station) {
      this.fillupModel.fillingStation = [station.brand || station.name, station.street, station.place].filter(Boolean).join(', ')
    },
    formatDate(date) {
      const finalFormat = this.me.dateFormat ? this.me.dateFormat : 'MM/dd/yyyy'
      return format(date, finalFormat)
    },
    fetchVehicleUsers() {
      store
        .dispatch('vehicles/fetchUsersByVehicleId', { vehicleId: this.selectedVehicle.id })
        .then((data) => {
          this.users = data
        })
        .catch((err) => console.log(err))
    },
    fetchVehicleFuelSubTypes() {
      store
        .dispatch('vehicles/fetchFuelSubtypesByVehicleId', { vehicleId: this.selectedVehicle.id })
        .then((data) => {
          this.fuelSubTypes = data
        })
        .catch((err) => console.log(err))
    },
    getEmptyFillup() {
      return {
        vehicleId: this.selectedVehicle.id,
        fuelUnit: this.selectedVehicle.fuelUnit,
        perUnitPrice: null,
        fuelQuantity: null,
        totalAmount: null,
        odoReading: '',
        isTankFull: true,
        hasMissedFillup: false,
        date: new Date(),
        fillingStation: '',
        driverId: '',
        comments: '',
        userId: '',
        fuelSubType: '',
      }
    },
    async createFillup() {
      this.tryingToCreate = true
      this.fillupModel.vehicleId = this.selectedVehicle.id
      //  this.fillupModel.userId = this.me.id
      if (this.fillup.id) {
        axios
          .put(`/api/vehicles/${this.selectedVehicle.id}/fillups/${this.fillup.id}`, this.fillupModel)
          .then((data) => {
            this.$buefy.toast.open({
              message: 'Fillup Updated Successfully',
              type: 'is-success',
              duration: 3000,
            })
            this.fillupModel = this.getEmptyFillup()
            if (this.processQuickEntry) {
              store.dispatch('vehicles/setQuickEntryAsProcessed', { id: this.quickEntry.id }).then((data) => {
                this.quickEntry = null
              })
            }
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
          .post(`/api/vehicles/${this.selectedVehicle.id}/fillups`, this.fillupModel)
          .then((data) => {
            this.$buefy.toast.open({
              message: 'Fillup Created Successfully',
              type: 'is-success',
              duration: 3000,
            })
            this.fillupModel = this.getEmptyFillup()
            if (this.processQuickEntry) {
              store.dispatch('vehicles/setQuickEntryAsProcessed', { id: this.quickEntry.id }).then((data) => {})
            }
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
    <div class="has-text-centered">
      <div class="columns">
        <div class="column is-two-thirds">
          <h1 class="title">Tankung hinzufügen</h1>
          <h1 class="subtitle">
            {{ [selectedVehicle.nickname, selectedVehicle.registration, selectedVehicle.make, selectedVehicle.model].join(' | ') }}
          </h1>
        </div>
        <div class="column is-one-thirds">
          <QuickEntryDisplay v-model="quickEntry" :user="user" />
        </div>
      </div>
    </div>
    <form class="" @submit.prevent="createFillup">
      <b-field label="Fahrzeug">
        <b-select v-model="selectedVehicle" placeholder="Vehicle" required expanded :disabled="fillup.id">
          <option v-for="option in myVehicles" :key="option.id" :value="option">
            {{ option.nickname }}
          </option>
        </b-select>
      </b-field>
      <b-field label="Erfasst von">
        <b-select v-model="fillupModel.userId" placeholder="User" required expanded :disabled="fillup.id">
          <option v-for="option in users" :key="option.userId" :value="option.userId">
            {{ option.name }}
          </option>
        </b-select>
      </b-field>
      <b-field label="Wer hat getankt? (optional)">
        <b-select v-model="fillupModel.driverId" expanded placeholder="Fahrer auswählen">
          <option value="">Kein Fahrer ausgewählt</option>
          <option v-for="driver in drivers" :key="driver.id" :value="driver.id">{{ driver.name }}</option>
        </b-select>
        <p class="help">Fahrer kannst du links unter „Fahrer“ anlegen.</p>
      </b-field>
      <b-field label="Datum der Tankung">
        <b-datepicker
          v-model="fillupModel.date"
          :date-formatter="formatDate"
          placeholder="Click to select..."
          icon="calendar"
          trap-focus
          :max-date="new Date()"
        >
        </b-datepicker>
      </b-field>
      <b-field label="Kraftstoffsorte">
        <b-autocomplete
          v-model="fillupModel.fuelSubType"
          :data="filteredFuelSubtypes"
          placeholder="Octane etc."
          clearable
          autofocus
          @select="(option) => (fillupModel.fuelSubType = option)"
        >
        </b-autocomplete>
      </b-field>
      <b-field label="Getankte Menge*" addons>
        <b-input v-model.number="fillupModel.fuelQuantity" type="number" step=".001" min="0" expanded required></b-input>
        <b-select v-model="fillupModel.fuelUnit" placeholder="Einheit" required>
          <option v-for="(option, key) in fuelUnitMasters" :key="key" :value="key">
            {{ option.long }}
          </option>
        </b-select>
      </b-field>
      <b-field :label="'Price per ' + vehicle.fuelUnitDetail.short + '*'"
        ><p class="control">
          <span class="button is-static">{{ me.currency }}</span>
        </p>
        <b-input v-model.number="fillupModel.perUnitPrice" type="number" min="0" step=".001" expanded required></b-input>
      </b-field>
      <b-field label="Gesamtbetrag">
        <p class="control">
          <span class="button is-static">{{ me.currency }}</span>
        </p>
        <b-input v-model.number="fillupModel.totalAmount" type="number" min="0" step=".001" expanded required></b-input>
      </b-field>
      <b-field label="Aktueller Kilometerstand">
        <p class="control">
          <span class="button is-static">{{ me.distanceUnitDetail.short }}</span>
        </p>
        <b-input v-model.number="fillupModel.odoReading" type="number" min="0" expanded required></b-input>
      </b-field>
      <b-field>
        <b-checkbox v-model="fillupModel.isTankFull">Vollgetankt?</b-checkbox>
      </b-field>
      <b-field>
        <b-checkbox v-model="fillupModel.hasMissedFillup">Wurde seit der letzten Volltankung eine Tankung ausgelassen?</b-checkbox>
      </b-field>
      <b-field>
        <b-switch v-model="showMore">Weitere Angaben</b-switch>
      </b-field>
      <fieldset v-if="showMore">
        <b-field label="Tankstelle">
          <b-input v-model="fillupModel.fillingStation" type="text" placeholder="Tankstelle selbst eintragen oder unten suchen" expanded></b-input>
        </b-field>
        <b-field label="Tankstellen in Deutschland suchen (PLZ)">
          <b-input v-model="stationPostal" maxlength="5" placeholder="z. B. 69115" expanded></b-input>
          <p class="control"><b-button :loading="stationLoading" @click="searchStations">Suchen</b-button></p>
        </b-field>
        <p v-if="stationError" class="help is-danger">{{ stationError }}</p>
        <b-field v-if="stations.length" label="Tankstelle übernehmen">
          <b-select expanded placeholder="Tankstelle wählen" @input="chooseStation">
            <option v-for="station in stations" :key="station.id" :value="station">{{ station.brand || station.name }} – {{ station.street }}, {{ station.place }} ({{ station.dist }} km)</option>
          </b-select>
        </b-field>
        <b-field label="Notizen">
          <b-input v-model="fillupModel.comments" type="textarea" expanded></b-input>
        </b-field>
      </fieldset>
      <b-field>
        <b-switch v-if="quickEntry" v-model="processQuickEntry">Mark selected Quick Entry as processed</b-switch>
      </b-field>
      <br />
      <b-field>
        <b-button tag="input" native-type="submit" :disabled="tryingToCreate" type="is-primary" label="Tankung hinzufügen" expanded> </b-button>
        <p v-if="authError">
          There was an error logging in to your account.
        </p>
      </b-field>
    </form>
  </Layout>
</template>
