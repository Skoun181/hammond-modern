<script>
import Layout from '@layouts/main.vue'
import store from '@state/store'

import { mapState } from 'vuex'
import axios from 'axios'
import { round } from 'lodash'
import { format } from 'date-fns'

export default {
  page: {
    title: 'Tankung hinzufügen',
  },
  components: { Layout },
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
      myVehicles: [],
      users: [],
      fuelSubTypes: [],
      drivers: [],
      stationPostal: '',
      stations: [],
      stationLoading: false,
      stationError: '',
      receiptLoading: false,
      receiptError: '',
      receiptNotice: '',
      receiptFields: [],
      selectedVehicle: this.vehicle,
      fillupModel: this.fillup,
    }
  },

  computed: {
    user() {
      return store.state.auth.currentUser
    },
    ...mapState('users', ['me']),
    ...mapState('vehicles', ['fuelUnitMasters', 'fuelTypeMasters', 'vehicles']),
    filteredFuelSubtypes() {
      const fuelType = this.selectedVehicle && this.selectedVehicle.fuelType
      const defaults = fuelType === 0 || fuelType === 6 ? ['Super E5', 'Super E10'] : []
      const options = Array.from(new Set([...defaults, ...(this.fuelSubTypes || [])])).filter(option => typeof option === 'string')
      if (!this.fillupModel.fuelSubType) {
        return options
      }
      return options.filter((option) => {
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
    // iPhone images often arrive as HEIC or large JPEGs.
    // Decode them on-device and send a compressed JPEG to our local OCR API.
    prepareReceiptImage(file) {
      return new Promise((resolve, reject) => {
        const url = URL.createObjectURL(file)
        const image = new Image()
        image.onload = () => {
          URL.revokeObjectURL(url)
          const longestSide = Math.max(image.naturalWidth, image.naturalHeight)
          if (!longestSide) { reject(new Error('Bild konnte nicht gelesen werden.')); return }
          const scale = Math.min(1, 2400 / longestSide)
          const canvas = document.createElement('canvas')
          canvas.width = Math.round(image.naturalWidth * scale)
          canvas.height = Math.round(image.naturalHeight * scale)
          const context = canvas.getContext('2d')
          if (!context) { reject(new Error('Bildverarbeitung nicht verfügbar.')); return }
          context.fillStyle = '#ffffff'
          context.fillRect(0, 0, canvas.width, canvas.height)
          context.drawImage(image, 0, 0, canvas.width, canvas.height)
          canvas.toBlob(blob => {
            if (!blob) { reject(new Error('Bild konnte nicht konvertiert werden.')); return }
            resolve(new File([blob], 'tankbeleg.jpg', { type: 'image/jpeg' }))
          }, 'image/jpeg', 0.88)
        }
        image.onerror = () => {
          URL.revokeObjectURL(url)
          reject(new Error('Dieses Fotoformat konnte nicht geöffnet werden. Bitte ein JPEG auswählen.'))
        }
        image.src = url
      })
    },
    async recognizeReceipt(event) {
      const file = event.target.files && event.target.files[0]
      if (!file) return
      this.receiptError = ''
      this.receiptNotice = ''
      this.receiptFields = []
      if (!file.type.startsWith('image/') && !/\.(heic|heif|jpg|jpeg|png)$/i.test(file.name)) {
        this.receiptError = 'Bitte eine Bilddatei auswählen (JPEG, PNG oder HEIC).'
        event.target.value = ''
        return
      }
      if (file.size > 25 * 1024 * 1024) {
        this.receiptError = 'Das Originalfoto ist größer als 25 MB.'
        event.target.value = ''
        return
      }
      this.receiptLoading = true
      try {
        const converted = await this.prepareReceiptImage(file)
        if (converted.size > 12 * 1024 * 1024) throw new Error('Das Foto ist auch nach der Komprimierung zu groß.')
        const form = new FormData()
        form.append('file', converted, converted.name)
        const response = await axios.post('/api/receipts/recognize', form, { timeout: 65000 })
        const fields = response.data.suggestions || {}
        if (fields.fillingStation) {
          this.fillupModel.fillingStation = fields.fillingStation
          this.receiptFields.push('Tankstelle')
          this.showMore = true
        }
        if (fields.date) {
          const date = new Date(fields.date + 'T12:00:00')
          if (!isNaN(date.getTime()) && date <= new Date()) {
            this.fillupModel.date = date
            this.receiptFields.push('Datum')
          }
        }
        if (fields.fuelQuantity != null) {
          this.fillupModel.fuelQuantity = fields.fuelQuantity
          this.receiptFields.push('Liter')
        }
        if (fields.perUnitPrice != null) {
          this.fillupModel.perUnitPrice = fields.perUnitPrice
          this.receiptFields.push('Preis/Liter')
        }
        await this.$nextTick()
        if (fields.totalAmount != null) {
          this.fillupModel.totalAmount = fields.totalAmount
          this.receiptFields.push('Gesamtbetrag')
        }
        this.receiptNotice = this.receiptFields.length
          ? 'Erkannt: ' + this.receiptFields.join(', ') + '. ' + (fields.amountEstimated ? 'Der Kraftstoffbetrag wurde aus Liter × Literpreis berechnet; andere Einkäufe auf dem Bon sind nicht enthalten. ' : '') + 'Bitte alle Angaben inklusive Datum kontrollieren.'
          : 'Keine Felder zuverlässig erkannt. Bitte Werte selbst eingeben.'
      } catch (error) {
        this.receiptError = (error.response && error.response.data && error.response.data.error) ||
          error.message || 'Beleg konnte nicht erkannt werden. Bitte Daten manuell eingeben.'
      } finally {
        this.receiptLoading = false
        event.target.value = ''
      }
    },
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
        <div class="column">
          <h1 class="title">Tankung hinzufügen</h1>
          <h1 class="subtitle">
            {{ [selectedVehicle.nickname, selectedVehicle.registration, selectedVehicle.make, selectedVehicle.model].join(' | ') }}
          </h1>
        </div>
      </div>
    </div>
    <div v-if="!fillup.id" class="box">
      <h2 class="title is-5">Tankbeleg scannen</h2>
      <p class="help">Auf dem Handy die Kamera öffnen oder ein vorhandenes Foto auswählen. Die Texterkennung läuft lokal auf deinem Hammond-Server. Das Foto wird nur für die Erkennung verarbeitet und nicht gespeichert.</p>
      <label class="button is-primary hm-camera-button" for="receipt-camera">📷 Beleg fotografieren / hochladen</label>
      <input id="receipt-camera" type="file" accept="image/*,.heic,.heif" capture="environment" style="position:absolute;width:1px;height:1px;opacity:0" :disabled="receiptLoading" @change="recognizeReceipt">
      <p v-if="receiptLoading" class="help">Beleg wird erkannt…</p>
      <b-notification v-if="receiptError" type="is-danger" :closable="false">{{ receiptError }}</b-notification>
      <b-notification v-if="receiptNotice" type="is-info" :closable="false">{{ receiptNotice }}</b-notification>
      <p class="help">Auch iPhone-HEIC-Fotos werden auf dem Gerät in ein komprimiertes JPEG umgewandelt. Datum, Tankstelle, Menge und Preise vor dem Speichern prüfen.</p>
    </div>
    <form class="hm-fillup-form" @submit.prevent="createFillup">
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
          placeholder="z. B. Super E5 oder Super E10"
          open-on-focus
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
      <b-field :label="'Preis pro ' + vehicle.fuelUnitDetail.short + '*'"
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
