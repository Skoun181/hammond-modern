<script>
import { formatMoney, formatDecimal } from '@utils/format-modern'
import { vehicleFuelStatistics } from '@utils/fuel-statistics'

import appConfig from '@src/app.config'
import Layout from '@layouts/main.vue'
import MaintenanceOverview from '@components/maintenanceOverview.vue'
import StatsWidget from '@components/statsWidget.vue'
import CostHistory from '@components/costHistory.vue'
import { chunk } from 'lodash'

import { parseAndFormatDate } from '@utils/format-date'
// import store from '@state/store'
import { mapState } from 'vuex'

export default {
  page: {
    title: 'Home',
    meta: [{ name: 'description', content: appConfig.description }],
  },
  components: { Layout, MaintenanceOverview, StatsWidget, CostHistory },
  props: {
    user: {
      type: Object,
      required: true,
    },
  },

  data: function() {
    return {
      myVehicles: [],

      tryingToCreate: false,
      vehiclesLoading: false,
      vehiclesError: '',
    }
  },
  computed: {
    ...mapState('users', ['me']),
    ...mapState('vehicles', ['vehicles']),
    ...mapState('utils', ['isMobile']),
    chunkedVehicles() {
      return chunk(this.myVehicles, 3)
    },
    firstVehicleId() { return this.myVehicles.length ? this.myVehicles[0].id : null },
  },
  watch: {
    vehicles(newVehicles) {
      this.myVehicles = newVehicles || []
    },
  },
  mounted() {
    this.myVehicles = this.vehicles || []
    this.refreshVehicles()
  },
  methods: {
    startFillup(id) { this.$router.push('/vehicles/' + id + '/fillup') },
    async refreshVehicles() {
      this.vehiclesLoading = true
      this.vehiclesError = ''
      try {
        await this.$store.dispatch('vehicles/fetchVehicles')
      } catch (error) {
        this.vehiclesError = 'Fahrzeuge konnten nicht geladen werden. Bitte neu anmelden oder erneut versuchen.'
      } finally {
        this.vehiclesLoading = false
      }
    },
    formatDecimal(value, digits = 2) { return formatDecimal(value, digits) },
    fuelStats(fillups) { return vehicleFuelStatistics(fillups) },
    formatDate(date) {
      return parseAndFormatDate(date)
    },
    formatCurrency(number) {
      return formatMoney(number, (this.me && this.me.currency) || 'EUR')
    },
  },
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading">
      <div><p class="hm-eyebrow">ÜBERSICHT</p><h1 class="title hm-page-title">Dein Dashboard</h1><p class="hm-page-description">Fahrzeuge, Kosten und Tankvorgänge auf einen Blick.</p></div>
      <b-button v-if="myVehicles.length === 1" type="is-primary" tag="router-link" :to="'/vehicles/' + firstVehicleId + '/fillup'">＋ Tankung hinzufügen</b-button>
      <b-dropdown v-else-if="myVehicles.length > 1" position="is-bottom-left" aria-role="list">
        <template slot="trigger">
          <b-button type="is-primary">＋ Tankung hinzufügen</b-button>
        </template>
        <b-dropdown-item v-for="v in myVehicles" :key="v.id" aria-role="listitem" @click="startFillup(v.id)">
          {{ v.nickname }} – {{ v.registration }}
        </b-dropdown-item>
      </b-dropdown>
    </div>
    <b-notification v-if="vehiclesError" type="is-danger is-light" :closable="false">{{ vehiclesError }} <b-button size="is-small" @click="refreshVehicles">Erneut versuchen</b-button></b-notification>
    <div class="hm-overview-strip"><div class="hm-overview-icon">▣</div><div><p class="hm-overview-label">VERWALTETE FAHRZEUGE</p><strong>{{ myVehicles.length }}</strong></div><span class="hm-overview-note">Aktive Übersicht</span></div>
    <b-notification v-if="!vehiclesLoading && !vehiclesError && myVehicles.length === 0" type="is-warning is-light" :closable="false">
      <div class="columns">
        <div class="column">
          Noch keine Fahrzeuge vorhanden. Lege dein erstes Fahrzeug an, um Ausgaben und Tankvorgänge zu verfolgen.
        </div>
        <div class="column" :class="!isMobile ? 'has-text-right' : ''">
          <b-button type="is-warning" class="" tag="router-link" :to="`/vehicles/create`"
            >Jetzt anlegen</b-button
          ></div
        >
      </div>
    </b-notification>
    <p v-if="vehiclesLoading" class="help">Fahrzeuge werden geladen…</p>
    <MaintenanceOverview :limit="4" />
    <div class="box"><div class="hm-page-heading"><div><h2 class="title is-5">Fixkosten &amp; Finanzierung</h2><p class="help">Versicherung, Steuer, Leasing und Finanzierungsraten automatisch erfassen.</p></div><b-button tag="router-link" :to="{name:'recurring-costs'}">Kostenpläne</b-button></div></div>
    <StatsWidget :user="me" />
    <CostHistory />
    <br />
    <section>
      <div class="columns" :class="isMobile ? 'has-text-centered' : ''"
        ><div class="column is-three-quarters"> <h2 class="title">Meine Fahrzeuge</h2></div>
</div
      >

      <div v-if="myVehicles.length">
  <div v-for="chunk,index in chunkedVehicles" :key="index"  class="columns">
        <div v-for="vehicle in chunk" :key="vehicle.id" class="column is-4">
          <b-collapse animation="slide" aria-id="contentIdForA11y3" class="card" :open="!isMobile">
            <template v-slot:trigger="props">
              <div class="card-header" role="button" aria-controls="contentIdForA11y3">
                <div class="card-header-title">
                  <div>{{ `${vehicle.nickname} - ${vehicle.registration}` }} </div>
                </div>
                <a class="card-header-icon">
                  <b-icon :icon="props.open ? 'angle-down' : 'angle-up'"> </b-icon>
                </a>
              </div>
            </template>
            <div v-if="(vehicle.fillups || []).length" class="card-content">
              <div class="content">
                <table class="table">
                  <div class="columns">
                    <div class="column is-one-third">Letzte Tankung</div>
                    <div class="column"
                      >{{ formatDate(vehicle.fillups[0].date) }} <br />
                      {{ `${formatCurrency(vehicle.fillups[0].totalAmount)}` }} ({{
                        `${formatDecimal(vehicle.fillups[0].fuelQuantity)} ${vehicle.fillups[0].fuelUnitDetail.short}`
                      }}
                      @ {{ `${formatCurrency(vehicle.fillups[0].perUnitPrice)}` }})</div
                    >
                  </div>

                  <div class="columns">
                    <div class="column is-one-third">Kilometerstand</div>
                    <div class="column">
                      <template v-if="vehicle.fillups.length">
                        {{ formatDecimal(vehicle.fillups[0].odoReading, 0) }}&nbsp;{{
                          (me && me.distanceUnitDetail ? me.distanceUnitDetail.short : "km")
                        }}</template
                      >
                    </div>
                  </div>
                </table>
              </div>
            </div>
            <div class="card-content">
              <div class="columns is-mobile hm-fuel-metrics">
                <div class="column"><p class="heading">Getankt im Monat</p><strong>{{ formatDecimal(fuelStats(vehicle.fillups || []).monthLiters) }} L</strong></div>
                <div class="column"><p class="heading">Getankt im Jahr</p><strong>{{ formatDecimal(fuelStats(vehicle.fillups).yearLiters) }} L</strong></div>
              </div>
              <p class="heading">Ø Verbrauch (Volltankungen)</p>
              <p><strong>{{ fuelStats(vehicle.fillups).consumption === null ? 'Noch nicht berechenbar' : formatDecimal(fuelStats(vehicle.fillups).consumption) + ' L/100 km' }}</strong></p>
              <p class="help">Benötigt mindestens zwei Volltankungen mit steigenden Kilometerständen und keine ausgelassene Tankung.</p>
            </div>
            <footer class="card-footer">
              <router-link class="card-footer-item" :to="'/vehicles/' + vehicle.id">
                Details
              </router-link>
              <router-link class="card-footer-item" :to="`/vehicles/${vehicle.id}/fillup`">
                Tankung hinzufügen </router-link
              ><router-link class="card-footer-item" :to="`/vehicles/${vehicle.id}/expense`">
                Ausgabe hinzufügen
              </router-link>
            </footer>
          </b-collapse>
        </div>
        </div>
      </div>
    </section>
  </Layout>
</template>
