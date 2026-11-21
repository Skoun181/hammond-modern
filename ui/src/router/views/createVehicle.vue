<script>
import Layout from '@layouts/main.vue'
import { mapState } from 'vuex'
import axios from 'axios'
import { vehicleBrands } from '@utils/vehicle-brands'

export default {
  page: {
    title: 'Fahrzeug hinzufügen',
  },
  components: { Layout },
  props: {
    vehicle: {
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
      vehicleModel: {},
      brands: vehicleBrands,
      models: [],
      catalogError: '',
      loadingModels: false,
      modelRequestId: 0,
    }
  },

  computed: {
    ...mapState('users', ['me']),
    ...mapState('vehicles', ['fuelUnitMasters', 'fuelTypeMasters', 'vehicles']),
  },
  watch: {
    'vehicleModel.make': function(newMake, oldMake) {
      if (newMake === oldMake) return
      this.models = []
      if (oldMake) this.vehicleModel.model = ''
      this.loadModels(newMake)
    },
  },
  mounted() {
    if (!this.vehicle) {
      this.vehicleModel = this.getEmptyVehicle()
    } else {
      this.vehicleModel = this.getEmptyVehicle(this.vehicle)
    }
    this.myVehicles = this.vehicles
    if (this.vehicleModel.make) this.loadModels(this.vehicleModel.make)
  },
  methods: {
    async loadModels(make) {
      const requestId = ++this.modelRequestId
      this.models = []
      this.catalogError = ''
      if (!make || make.trim().length < 2) {
        this.loadingModels = false
        return
      }
      this.loadingModels = true
      try {
        const response = await axios.get('/api/vehicle-models', {
          params: { make: make.trim() },
          timeout: 15000,
        })
        if (requestId !== this.modelRequestId || this.vehicleModel.make !== make) return
        this.models = Array.isArray(response.data.models) ? response.data.models.sort((a, b) => a.localeCompare(b, 'de')) : []
        if (!this.models.length) {
          this.catalogError = 'Für diese Marke sind keine Modelle hinterlegt. Bitte Modell selbst eingeben.'
        }
      } catch (error) {
        if (requestId !== this.modelRequestId) return
        this.catalogError = (error.response && error.response.data && error.response.data.error)
          || 'Modellvorschläge derzeit nicht verfügbar. Bitte Modell selbst eingeben.'
      } finally {
        if (requestId === this.modelRequestId) this.loadingModels = false
      }
    },
    getEmptyVehicle(veh) {
      if (!veh || !veh.id) {
        return {
          fuelUnit: null,
          fuelType: null,
          registration: '',
          nickname: '',
          engineSize: null,
          make: '',
          model: '',
          yearOfManufacture: null,
        }
      } else {
        return {
          fuelUnit: veh.fuelUnit,
          fuelType: veh.fuelType,
          registration: veh.registration,
          nickname: veh.nickname,
          engineSize: veh.engineSize,
          make: veh.make,
          model: veh.model,
          yearOfManufacture: veh.yearOfManufacture,
        }
      }
    },
    createVehicle() {
      this.tryingToCreate = true
      this.vehicleModel.userId = this.me.id
      if (this.vehicle.id) {
        axios
          .put(`/api/vehicles/${this.vehicle.id}`, this.vehicleModel)
          .then(async () => {
            await this.$store.dispatch('vehicles/fetchVehicles')
            this.$buefy.toast.open({
              message: 'Fahrzeug erfolgreich aktualisiert',
              type: 'is-success',
              duration: 3000,
            })
            // this.vehicleModel = this.getEmptyVehicle()
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
            this.tryingToCreate = false
          })
      } else {
        axios
          .post(`/api/vehicles`, this.vehicleModel)
          .then(async () => {
            await this.$store.dispatch('vehicles/fetchVehicles')
            this.$buefy.toast.open({
              message: 'Fahrzeug erfolgreich angelegt',
              type: 'is-success',
              duration: 3000,
            })
            this.vehicleModel = this.getEmptyVehicle()
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
      <div class="column is-three-quarters">
        <h1 class="title">{{ vehicle.id ? 'Fahrzeug bearbeiten' : 'Fahrzeug hinzufügen' }}</h1>
      </div>
      <div class="column is-one-quarter">
        <router-link tag="b-button" type="is-primary" :to="{ name: 'vehicles' }">
          Zur Fahrzeugübersicht
        </router-link>
      </div>
    </div>
    <form @submit.prevent="createVehicle">
      <b-field label="Fahrzeugname*">
        <b-input v-model="vehicleModel.nickname" type="text" expanded required></b-input>
      </b-field>
      <b-field label="Kennzeichen*">
        <b-input v-model="vehicleModel.registration" type="text" expanded required></b-input>
      </b-field>
      <b-field label="Kraftstoffart*">
        <b-select v-model.number="vehicleModel.fuelType" placeholder="Fuel Type" required expanded>
          <option v-for="(option, key) in fuelTypeMasters" :key="key" :value="key">
            {{ option.long }}
          </option>
        </b-select>
      </b-field>

      <b-field label="Kraftstoffeinheit*">
        <b-select v-model.number="vehicleModel.fuelUnit" placeholder="Fuel Unit" required expanded>
          <option v-for="(option, key) in fuelUnitMasters" :key="key" :value="key">
            {{ option.long }}
          </option>
        </b-select>
      </b-field>

      <b-field label="Automarke*">
        <b-autocomplete v-model="vehicleModel.make" :data="brands.filter(b => b.toLocaleLowerCase('de').includes((vehicleModel.make || '').toLocaleLowerCase('de'))).slice(0, 30)" placeholder="Marke suchen oder eingeben" keep-first open-on-focus clearable expanded required @select="option => vehicleModel.make = option" />
      </b-field>
      <b-field label="Modell*">
        <b-autocomplete v-model="vehicleModel.model" :data="models.filter(m => m.toLowerCase().includes((vehicleModel.model || '').toLowerCase())).slice(0, 40)" placeholder="Modell auswählen oder eingeben" clearable expanded required @select="option => vehicleModel.model = option" />
        <p class="help">{{ loadingModels ? 'Modelle werden geladen…' : catalogError }}</p>
      </b-field>
      <b-field label="Baujahr">
        <b-input v-model.number="vehicleModel.yearOfManufacture" type="number" expanded number></b-input>
      </b-field>
      <b-field label="Hubraum (cm³)">
        <b-input v-model.number="vehicleModel.engineSize" type="number" expanded number></b-input>
      </b-field>

      <br />
      <b-field>
        <b-button tag="input" native-type="submit" :disabled="tryingToCreate" type="is-primary" :label="vehicle.id ? 'Änderungen speichern' : 'Fahrzeug hinzufügen'" expanded>
          <BaseIcon v-if="tryingToCreate" name="sync" spin />
        </b-button>
        <p v-if="authError">
          There was an error logging in to your account.
        </p>
      </b-field>
    </form>
  </Layout>
</template>
