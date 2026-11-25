<script>
import Layout from '@layouts/main.vue'
import axios from 'axios'
import { mapState } from 'vuex'
import store from '@state/store'

export default {
  components: { Layout },
  page: {
    title: 'First Setup',
  },
  data: function() {
    return {
      isWorking: false,
      registerModel: {
        name: '',
        email: '',
        password: '',
        distanceUnit: 1,
        currency: 'EUR',
      },
    }
  },
  computed: {
    ...mapState('vehicles', ['currencyMasters']),
  },
  mounted() {
    store.dispatch('vehicles/fetchMasters').catch(() => {})
  },
  methods: {
    showSuccessModal() {
      const message = 'Einrichtung abgeschlossen. Du kannst dich jetzt anmelden.'
      this.$buefy.toast.open({
        duration: 10000,
        message: message,
        position: 'is-bottom',
        type: 'is-success',
      })
      setTimeout(() => {
        this.$router
          .push({ name: 'login' })
          .then((succ) => {})
          .catch((err) => console.log('error:', err))
      }, 10000)
    },
    register() {
      this.isWorking = true
      axios
        .post('/api/auth/initialize', this.registerModel)
        .then((response) => {
          const success = response.data.success
          if (success) {
            store.dispatch('auth/systemInitialized').then((data) => {
              this.showSuccessModal()
            })
          }
        })
        .catch((ex) => {
          this.testSuccess = false
          this.$buefy.toast.open({
            duration: 5000,
            message: ex.message,
            position: 'is-bottom',
            type: 'is-danger',
          })
        })
        .finally(() => (this.isWorking = false))
    },

  },
}
</script>

<template>
  <Layout>
    <section class="section">
      <h1 class="title">Hammond Modern einrichten</h1>
      <form @submit.prevent="register">
        <b-field label="Name"><b-input v-model="registerModel.name" required /></b-field>
        <b-field label="E-Mail"><b-input v-model="registerModel.email" type="email" required /></b-field>
        <b-field label="Passwort"><b-input v-model="registerModel.password" type="password" minlength="8" required /></b-field>
        <b-field label="Währung">
          <b-select v-model="registerModel.currency" required>
            <option v-for="currency in currencyMasters" :key="currency.code" :value="currency.code">{{ currency.code }}</option>
          </b-select>
        </b-field>
        <b-field label="Entfernung">
          <b-select v-model.number="registerModel.distanceUnit">
            <option :value="0">Meilen</option>
            <option :value="1">Kilometer</option>
          </b-select>
        </b-field>
        <b-button native-type="submit" type="is-primary" :loading="isWorking">Admin-Konto erstellen</b-button>
      </form>
      <p>Eine Clarkson-Datenbank lässt sich lokal mit dem Serverbefehl <code>--import-clarkson</code> übernehmen.</p>
    </section>
  </Layout>
</template>
