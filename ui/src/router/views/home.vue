<script>
import { formatMoney, formatDecimal } from '@utils/format-modern'
import { vehicleFuelStatistics } from '@utils/fuel-statistics'

import appConfig from '@src/app.config'
import Layout from '@layouts/main.vue'
import CreateQuickEntry from '@components/createQuickEntry.vue'
import StatsWidget from '@components/statsWidget.vue'
import { chunk } from 'lodash'

import { parseAndFormatDate } from '@utils/format-date'
// import store from '@state/store'
import { mapGetters, mapState } from 'vuex'

export default {
  page: {
    title: 'Home',
    meta: [{ name: 'description', content: appConfig.description }],
  },
  components: { Layout, CreateQuickEntry, StatsWidget },
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
    ...mapGetters('vehicles', ['unprocessedQuickEntries']),
    chunkedVehicles() {
      return chunk(this.myVehicles, 3)
    },
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
      <b-button type="is-primary" tag="router-link" :to="'/vehicles/create'">＋ Fahrzeug hinzufügen</b-button>
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
    <b-notification
      v-if="unprocessedQuickEntries.length"
      type="is-warning is-light"
      :closable="false"
    >
      <div class="columns">
        <div class="column">
          {{
            `Du hast ${unprocessedQuickEntries.length} unbearbeitete Schnelleinträge.`
          }}
        </div>
        <div class="column" :class="!isMobile ? 'has-text-right' : ''">
          <b-button type="is-warning" class="is-small" tag="router-link" :to="`/quickEntries`"
            >Jetzt bearbeiten</b-button
          ></div
        >
      </div>
    </b-notification>
    <p v-if="vehiclesLoading" class="help">Fahrzeuge werden geladen…</p>
    <CreateQuickEntry />
    <StatsWidget :user="me" />
    <br />
    <section>
      <div class="columns" :class="isMobile ? 'has-text-centered' : ''"
        ><div class="column is-three-quarters"> <h2 class="title">Meine Fahrzeuge</h2></div>
        <div class="column is-one-quarter buttons" :class="!isMobile ? 'has-text-right' : ''">
          <b-button type="is-primary" tag="router-link" :to="`/vehicles/create`"
            >Fahrzeug hinzufügen</b-button
          >
        </div></div
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
                          me.distanceUnitDetail.short
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
