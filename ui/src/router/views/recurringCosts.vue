<script>
import Layout from '@layouts/main.vue'
import axios from 'axios'
import {formatMoney} from '@utils/format-modern'
const categories=['Finanzierung','Leasing','Versicherung','Kfz-Steuer','Stellplatz','Abonnement','Sonstiges']
const today=()=>{const d=new Date();return d.getFullYear()+'-'+String(d.getMonth()+1).padStart(2,'0')+'-'+String(d.getDate()).padStart(2,'0')}
const emptyTerms = () => ({ down: '', downDate: today(), already: false, principal: '', nominal: '', effective: '' })
const blank=()=>({vehicleId:'',title:'',category:'Finanzierung',amount:'',startDate:today(),intervalMonths:1,installments:48,final:'',notes:'',confirmed:false,historyMode:'',alreadyRecorded:0,terms:emptyTerms()})
export default {
  page:{title:'Fixkosten & Finanzierung'},components:{Layout},
  data(){return {plans:[],vehicles:[],form:blank(),categories,showForm:false,loading:true,busy:false,error:'',success:'',filter:'',editingFinance:null,editTerms:emptyTerms()}},
  computed:{
    visible(){return this.plans.filter(p=>!this.filter||p.vehicleId===this.filter)},
    active(){return this.visible.filter(p=>!p.stopped&&!!p.nextDate)},
    totals(){
      const result=Object.create(null)
      for(const p of this.active){
        if(!result[p.currency])result[p.currency]={currency:p.currency,monthly:0,remaining:0,hasFinite:false}
        const x=result[p.currency];x.monthly+=p.amountCents/p.intervalMonths
        if(p.remainingCents!=null){x.remaining+=p.remainingCents;x.hasFinite=true}
      }
      return Object.values(result)
    },
    dueCount() {
      if (!/^\d{4}-\d{2}-\d{2}$/.test(this.form.startDate) || this.form.startDate > today()) return 0
      const [year, month, day] = this.form.startDate.split('-').map(Number)
      if (year < 1900 || month < 1 || month > 12 || day < 1 || day > 31) return 0
      const cap = Number(this.form.installments) > 0 ? Math.min(600, Number(this.form.installments)) : 12000
      const interval = Number(this.form.intervalMonths)
      if (![1, 3, 6, 12].includes(interval)) return 0
      let count = 0
      for (let i = 0; i < cap; i++) {
        const first = new Date(Date.UTC(year, month - 1 + i * interval, 1))
        const last = new Date(Date.UTC(first.getUTCFullYear(), first.getUTCMonth() + 1, 0)).getUTCDate()
        first.setUTCDate(Math.min(day, last))
        if (first.toISOString().slice(0, 10) > today()) break
        count++
      }
      return count
    },
    pastStart() { return !!this.form.startDate && this.form.startDate < today() },
    acknowledgedCount() { return this.pastStart && this.form.historyMode === 'recorded' ? Number(this.form.alreadyRecorded) : 0 },
    historyValid() {
      if (!this.pastStart) return true
      return this.form.historyMode === 'backfill' || (this.form.historyMode === 'recorded' && Number.isInteger(this.acknowledgedCount) && this.acknowledgedCount >= 0 && this.acknowledgedCount <= this.dueCount)
    },
    previewTotal() {
      const n = Number(this.form.installments)
      const a = this.cents(this.form.amount)
      return n > 0 && a > 0 ? (n - 1) * a + (this.form.final === '' ? a : this.cents(this.form.final)) + (this.form.category === 'Finanzierung' ? this.cents(this.form.terms.down || 0) : 0) : null
    },
  },
  watch:{
    dueCount(n) { this.form.alreadyRecorded = n; this.form.confirmed = false },
    'form.historyMode'(v) { if (v === 'recorded') this.form.alreadyRecorded = this.dueCount; this.form.confirmed = false },
    'form.alreadyRecorded'() { this.form.confirmed = false },
    'form.category'(v){if(v==='Finanzierung'){this.form.intervalMonths=1;if(!Number(this.form.installments))this.form.installments=48}else{this.form.installments=0;this.form.final=''}}},
  mounted(){this.load()},
  methods:{
    financePayload(t) {
      return { downPaymentCents: this.cents(t.down || 0), downPaymentDate: Number(t.down) > 0 ? t.downDate : '', alreadyRecorded: Number(t.down) > 0 && t.already, principalCents: t.principal === '' ? null : this.cents(t.principal), nominalBps: t.nominal === '' ? null : this.cents(t.nominal), effectiveBps: t.effective === '' ? null : this.cents(t.effective) }
    },
    startFinance(p) {
      const t = p.financing || {}
      this.showForm = false
      this.editingFinance = p
      this.editTerms = { down: t.downPaymentCents ? t.downPaymentCents / 100 : '', downDate: t.downPaymentDate || today(), already: !!t.alreadyRecorded, principal: t.principalCents == null ? '' : t.principalCents / 100, nominal: t.nominalBps == null ? '' : t.nominalBps / 100, effective: t.effectiveBps == null ? '' : t.effectiveBps / 100 }
      this.$nextTick(() => { const el = this.$el.querySelector('#finance-edit'); if (el) el.scrollIntoView({ block: 'start' }) })
    },
    async saveFinance() {
      if (this.busy || !this.editingFinance) return
      this.busy = true; this.error = ''; this.success = ''
      try {
        await axios.put('/api/recurring-costs/' + this.editingFinance.id + '/financing', this.financePayload(this.editTerms))
        this.editingFinance = null; await this.load(); this.success = 'Finanzierungsdetails gespeichert.'
      } catch (e) { this.error = (e.response && e.response.data && e.response.data.error) || 'Speichern fehlgeschlagen.' }
      finally { this.busy = false }
    },
    today,money(cents,currency){return formatMoney(cents/100,currency)},cents(v){return Math.round(Number(v)*100)},
    vehicleName(id){const v=this.vehicles.find(x=>x.id===id);return v?[v.nickname,v.registration].filter(Boolean).join(' · '):'Fahrzeug nicht verfügbar'},
    date(v){return v?v.split('-').reverse().join('.'):'–'},
    async load(){this.loading=true;try{const [r,v]=await Promise.all([axios.get('/api/recurring-costs'),this.$store.dispatch('vehicles/fetchVehicles')]);this.plans=r.data||[];this.vehicles=v||[]}catch(e){this.error='Kostenpläne konnten nicht geladen werden.'}finally{this.loading=false}},
    add(){this.editingFinance=null;this.form=blank();this.form.vehicleId=this.filter||(this.vehicles[0]||{}).id||'';this.showForm=true;this.error='';this.success=''},
    async save(){
      if(this.busy||!this.form.confirmed||!this.historyValid)return
      this.busy=true;this.error='';this.success=''
      try{const p=this.form;await axios.post('/api/recurring-costs',{vehicleId:p.vehicleId,title:p.title,category:p.category,amountCents:this.cents(p.amount),startDate:p.startDate,intervalMonths:Number(p.intervalMonths),installments:Number(p.installments||0),finalCents:!Number(p.installments)||p.final===''?null:this.cents(p.final),notes:p.notes,alreadyRecorded:this.acknowledgedCount,financing:p.category==='Finanzierung'?this.financePayload(p.terms):null});this.showForm=false;await this.load();this.success='Kostenplan angelegt. Fehlende fällige Raten werden automatisch verbucht.'}
      catch(e){this.error=(e.response&&e.response.data&&e.response.data.error)||'Speichern fehlgeschlagen.'}finally{this.busy=false}
    },
    async run(){if(this.busy)return;this.busy=true;this.error='';this.success='';try{const r=await axios.post('/api/recurring-costs/run');await this.load();this.success=r.data.booked+' fällige Ausgaben erfasst.'}catch(e){this.error=(e.response&&e.response.data&&e.response.data.error)||'Verbuchung fehlgeschlagen.';await this.load()}finally{this.busy=false}},
    async stop(p){if(this.busy||!window.confirm('Kostenplan endgültig stoppen? Bereits erfasste Ausgaben bleiben erhalten. Für weitere Raten einen neuen Plan anlegen.'))return;this.busy=true;this.error='';try{await axios.post('/api/recurring-costs/'+p.id+'/stop');await this.load();this.success='Kostenplan gestoppt.'}catch(e){this.error='Stoppen fehlgeschlagen.'}finally{this.busy=false}},
    status(p){return p.stopped?'Gestoppt':p.nextDate?'Aktiv':'Abgeschlossen'},
  }
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading">
      <div><p class="hm-eyebrow">KOSTENPLANUNG</p><h1 class="title hm-page-title">Fixkosten &amp; Finanzierung</h1><p class="hm-page-description">Wiederkehrende Fahrzeugkosten und Finanzierungsraten automatisch als Ausgaben erfassen.</p></div>
      <div class="buttons"><b-button :disabled="busy || loading" @click="run">Fällige verbuchen</b-button><b-button type="is-primary" :disabled="busy || loading || !vehicles.length" @click="add">＋ Kostenplan</b-button></div>
    </div>
    <b-notification v-if="error" type="is-danger" @close="error = ''">{{ error }}</b-notification>
    <b-notification v-if="success" type="is-success" @close="success = ''">{{ success }}</b-notification>
    <p class="box">Automatische Sollbuchungen sind keine Zahlungsnachweise. Es werden keine Überweisungen ausgelöst. Bei einer bestehenden Finanzierung kannst du den ursprünglichen Start und die gesamte Ratenanzahl angeben. Bereits erfasste Raten nicht erneut verbuchen.</p>
    <div v-if="editingFinance" id="finance-edit" class="box plan-form">
      <h2 class="title is-5">Finanzierungsdetails · {{ editingFinance.title }}</h2>
      <form @submit.prevent="saveFinance">
        <div class="columns is-multiline">
          <div class="column is-half"><b-field label="Anzahlung (optional)" message="Einmalig, zusätzlich zu den Raten."><b-input v-model="editTerms.down" type="number" min="0" max="1000000" step="0.01" :disabled="!!(editingFinance.financing && editingFinance.financing.downPaymentCents)" /></b-field></div>
          <div class="column is-half"><b-field label="Datum der Anzahlung"><input v-model="editTerms.downDate" class="input" type="date" min="1900-01-01" :max="today()" :required="Number(editTerms.down) > 0" :disabled="!!(editingFinance.financing && editingFinance.financing.downPaymentCents)" /></b-field></div>
          <div v-if="Number(editTerms.down) > 0" class="column is-full"><b-checkbox v-model="editTerms.already" :disabled="!!(editingFinance.financing && editingFinance.financing.downPaymentCents)">Anzahlung bereits als Ausgabe erfasst (nicht erneut buchen)</b-checkbox></div>
          <div class="column is-half"><b-field label="Nettodarlehensbetrag" message="Kreditbetrag laut Vertrag, ohne Anzahlung."><b-input v-model="editTerms.principal" type="number" min="0.01" max="1000000" step="0.01" /></b-field></div>
          <div class="column is-half"><b-field label="Sollzins p. a. (%)"><b-input v-model="editTerms.nominal" type="number" min="0" max="100" step="0.01" /></b-field></div>
          <div class="column is-half"><b-field label="Effektiver Jahreszins (%)"><b-input v-model="editTerms.effective" type="number" min="0" max="100" step="0.01" /></b-field></div>
        </div><p class="help">Vertragsraten bleiben unverändert. Ohne Häkchen wird die Anzahlung einmalig zum angegebenen Datum als Ausgabe gebucht. Nach dem Speichern bleiben Betrag und Datum der Anzahlung festgeschrieben.</p>
        <div class="buttons"><b-button type="is-primary" native-type="submit" :loading="busy" :disabled="busy">Details speichern</b-button><b-button :disabled="busy" @click="editingFinance = null">Abbrechen</b-button></div>
      </form>
    </div>
    <div v-if="showForm" class="box plan-form">
      <h2 class="title is-5">Kostenplan anlegen</h2>
      <form @submit.prevent="save">
        <div class="columns is-multiline">
          <div class="column is-half"><b-field label="Fahrzeug *"><b-select v-model="form.vehicleId" required expanded><option v-for="v in vehicles" :key="v.id" :value="v.id">{{ vehicleName(v.id) }}</option></b-select></b-field></div>
          <div class="column is-half"><b-field label="Kategorie"><b-select v-model="form.category" expanded><option v-for="category in categories" :key="category">{{ category }}</option></b-select></b-field></div>
          <div class="column is-half"><b-field label="Bezeichnung *"><b-input v-model="form.title" required maxlength="120" placeholder="z. B. Autofinanzierung" /></b-field></div>
          <div class="column is-half"><b-field label="Reguläre Rate / Betrag *" message="Währung aus deinen Einstellungen."><b-input v-model="form.amount" type="number" min="0.01" max="1000000" step="0.01" required /></b-field></div>
          <div class="column is-half"><b-field label="Erste Rate / Vertragsbeginn *" message="Auch ein bereits vergangenes Datum ist möglich. Datum der ersten tatsächlichen Fälligkeit."><input v-model="form.startDate" class="input" type="date" min="1900-01-01" required /></b-field></div>
          <div class="column is-half"><b-field label="Intervall"><b-select v-model.number="form.intervalMonths" :disabled="form.category === 'Finanzierung'" expanded><option :value="1">Monatlich</option><option :value="3">Vierteljährlich</option><option :value="6">Halbjährlich</option><option :value="12">Jährlich</option></b-select></b-field></div>
          <div class="column is-half"><b-field label="Gesamte Anzahl der Raten" :message="form.category === 'Finanzierung' ? 'Anzahl inklusive letzter Rate.' : '0 = unbegrenzt'"><b-input v-model="form.installments" type="number" :min="form.category === 'Finanzierung' ? 1 : 0" max="600" step="1" required /></b-field></div>
          <div v-if="Number(form.installments) > 0" class="column is-half"><b-field label="Abweichende letzte Rate / Schlussrate" message="Gesamtbetrag der letzten Rate, nicht zusätzlich zur normalen Rate. Leer = normale Rate."><b-input v-model="form.final" type="number" min="0.01" max="1000000" step="0.01" /></b-field></div>
          <div class="column is-full"><b-field label="Notizen"><b-input v-model="form.notes" maxlength="2000" type="textarea" /></b-field></div>
        </div>
        <div v-if="form.category === 'Finanzierung'" class="plan-preview"><h3 class="title is-6">Anzahlung und Vertragszinsen</h3><div class="columns is-multiline">
          <div class="column is-half"><b-field label="Anzahlung (optional)" message="Einmalig, zusätzlich zu den Raten."><b-input v-model="form.terms.down" type="number" min="0" max="1000000" step="0.01" /></b-field></div>
          <div class="column is-half"><b-field label="Datum der Anzahlung"><input v-model="form.terms.downDate" class="input" type="date" min="1900-01-01" :max="today()" :required="Number(form.terms.down) > 0" /></b-field></div>
          <div v-if="Number(form.terms.down) > 0" class="column is-full"><b-checkbox v-model="form.terms.already">Anzahlung bereits als Ausgabe erfasst (nicht erneut buchen)</b-checkbox></div>
          <div class="column is-half"><b-field label="Nettodarlehensbetrag" message="Kreditbetrag laut Vertrag, ohne Anzahlung."><b-input v-model="form.terms.principal" type="number" min="0.01" max="1000000" step="0.01" /></b-field></div>
          <div class="column is-half"><b-field label="Sollzins p. a. (%)"><b-input v-model="form.terms.nominal" type="number" min="0" max="100" step="0.01" /></b-field></div>
          <div class="column is-half"><b-field label="Effektiver Jahreszins (%)"><b-input v-model="form.terms.effective" type="number" min="0" max="100" step="0.01" /></b-field></div>
        </div><p class="help">Vertragsraten bleiben unverändert. Ohne Häkchen wird die Anzahlung einmalig zum angegebenen Datum als Ausgabe gebucht. Nach dem Speichern bleiben Betrag und Datum der Anzahlung festgeschrieben.</p></div>
        <p v-if="previewTotal !== null" class="plan-preview">Geplante Gesamtsumme: <strong>{{ (previewTotal / 100).toLocaleString('de-DE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</strong> in deiner Währung</p>
        <div v-if="pastStart" class="plan-preview">
          <h3 class="title is-6">Vergangene Raten</h3>
          <p>Seit der ersten Fälligkeit sind bis einschließlich heute {{ dueCount }} Raten laut Plan fällig geworden.</p>
          <b-field label="Wie sollen bisherige Raten behandelt werden?">
            <b-select v-model="form.historyMode" required expanded>
              <option value="" disabled>Bitte auswählen</option>
              <option value="backfill">Alle bisherigen Raten als Ausgaben nachtragen</option>
              <option value="recorded">Bereits erfasste Raten berücksichtigen (keine Doppelbuchung)</option>
            </b-select>
          </b-field>
          <b-field v-if="form.historyMode === 'recorded'" label="Anzahl bereits erfasster Raten" message="Die ersten N aufeinanderfolgenden Raten sind bereits in Hammond als Ausgaben erfasst. Nur fehlende spätere Raten werden nachgetragen.">
            <b-input v-model.number="form.alreadyRecorded" type="number" min="0" :max="dueCount" step="1" required />
          </b-field>
          <p v-if="historyValid">Neu zu erfassende bisherige Raten: <strong>{{ dueCount - acknowledgedCount }}</strong></p>
          <p class="help">Die Berechnung folgt dem Ratenplan, nicht dem Bankkonto. Bitte vorhandene Ausgaben vorher abgleichen.</p>
        </div>
        <b-field><b-checkbox v-model="form.confirmed">Ich möchte die fälligen Raten automatisch als Ausgaben erfassen lassen und habe bereits erfasste Raten berücksichtigt.</b-checkbox></b-field>
        <div class="buttons"><b-button type="is-primary" native-type="submit" :loading="busy" :disabled="busy || !form.confirmed || !historyValid">Kostenplan speichern</b-button><b-button :disabled="busy" @click="showForm = false">Abbrechen</b-button></div>
      </form>
    </div>
    <b-field label="Fahrzeug filtern"><b-select v-model="filter" expanded><option value="">Alle Fahrzeuge</option><option v-for="v in vehicles" :key="v.id" :value="v.id">{{ vehicleName(v.id) }}</option></b-select></b-field>
    <div v-for="t in totals" :key="t.currency" class="box plan-summary">
      <div><p class="heading">Reguläre monatliche Fixkosten (Durchschnitt)</p><strong>{{ money(t.monthly,t.currency) }}</strong></div>
      <div v-if="t.hasFinite"><p class="heading">Noch geplante Raten (begrenzte Pläne)</p><strong>{{ money(t.remaining,t.currency) }}</strong></div>
    </div>
    <p v-if="loading">Kostenpläne werden geladen …</p>
    <p v-else-if="!visible.length" class="box">Noch keine Kostenpläne vorhanden.</p>
    <div class="columns is-multiline">
      <div v-for="p in visible" :key="p.id" class="column is-half">
        <article class="box plan-card">
          <p class="hm-eyebrow">{{ p.category }} · {{ status(p) }}</p><h2 class="title is-5">{{ p.title }}</h2><p class="plan-vehicle">{{ vehicleName(p.vehicleId) }}</p>
          <div class="plan-highlight"><p class="heading">Reguläre Rate</p><strong>{{ money(p.amountCents,p.currency) }}</strong><p>Alle {{ p.intervalMonths }} Monat(e)</p></div>
          <div class="plan-grid"><div><p class="heading">Nächste Fälligkeit</p><strong>{{ date(p.nextDate) }}</strong><p v-if="p.nextDate">{{ money(p.nextCents,p.currency) }}</p></div><div><p class="heading">Berücksichtigt (Soll)</p><strong>{{ money(p.consideredCents,p.currency) }}</strong><p>{{ p.generated }} berücksichtigte Raten</p></div></div>
          <p>Erste Fälligkeit: {{ date(p.startDate) }}<span v-if="p.lastDate"> · Letzte Fälligkeit: {{ date(p.lastDate) }}</span></p>
          <template v-if="p.installments">
            <p class="heading">Ratenfortschritt: {{ p.generated }} / {{ p.installments }} erfasst</p><progress class="progress is-warning" :value="p.generated" :max="p.installments">{{ p.generated }}</progress>
            <p>{{ p.remainingInstallments }} noch nicht erfasste Raten · <strong>{{ money(p.remainingCents,p.currency) }}</strong></p><p v-if="p.finalCents !== null && p.finalCents !== undefined">Letzte Rate: {{ money(p.finalCents,p.currency) }}</p><p class="help">Geplante Ratensumme, kein Kreditrestsaldo. Gestoppte Pläne werden nicht weiter gebucht.</p>
          </template>
          <section v-if="p.category === 'Finanzierung'" class="plan-preview">
            <p class="heading">Gesamte geplante Zahlungen inkl. Anzahlung</p><strong>{{ money(p.totalCents,p.currency) }}</strong>
            <template v-if="p.financing">
              <p v-if="p.financing.downPaymentCents">Anzahlung: {{ money(p.financing.downPaymentCents,p.currency) }} · {{ date(p.financing.downPaymentDate) }}</p>
              <p v-if="p.financing.principalCents">Nettodarlehen: {{ money(p.financing.principalCents,p.currency) }}</p>
              <p v-if="p.financing.nominalBps !== null">Sollzins: {{ (p.financing.nominalBps / 100).toLocaleString('de-DE') }} % p. a.</p>
              <p v-if="p.financing.effectiveBps !== null">Effektiver Jahreszins: {{ (p.financing.effectiveBps / 100).toLocaleString('de-DE') }} %</p>
              <p v-if="p.financingCostCents !== null">Raten minus Nettodarlehen: {{ money(p.financingCostCents,p.currency) }}</p>
              <p v-if="p.estimatedBalanceCents !== null">Geschätzte Restschuld: {{ money(p.estimatedBalanceCents,p.currency) }}</p>
              <p class="help">Restschuld unter Annahme monatlicher Sollzinsen und planmäßiger Zahlungen, ohne Gebühren, Sondertilgungen oder abweichende Zinstage. Kein Bankrestsaldo.</p>
            </template>
            <b-button size="is-small" :disabled="busy" @click="startFinance(p)">Finanzierungsdetails ergänzen / bearbeiten</b-button>
          </section>
          <p v-if="p.notes" class="plan-notes">{{ p.notes }}</p>
          <div class="buttons plan-actions"><b-button tag="router-link" :to="{ name: 'vehicle-detail', params: { vehicleId: p.vehicleId } }">Fahrzeugdetails</b-button><b-button v-if="!p.stopped && p.nextDate" :disabled="busy" type="is-danger" outlined @click="stop(p)">Plan stoppen</b-button></div>
        </article>
      </div>
    </div>
  </Layout>
</template>

<style scoped>
.plan-card{height:100%;display:flex;flex-direction:column;gap:10px;overflow-wrap:anywhere}
.plan-card strong,.plan-summary strong,.plan-preview strong{color:var(--hm-text)!important}
.plan-card .title{margin:0!important}
.plan-vehicle{font-size:1.15rem;font-weight:750;color:var(--hm-text)}
.plan-highlight,.plan-grid>div,.plan-preview{background:var(--hm-input);border:1px solid var(--hm-border);padding:12px;border-radius:9px}
.plan-highlight strong,.plan-summary strong{font-size:1.65rem}
.plan-grid,.plan-summary{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px}
.plan-actions{margin-top:auto;padding-top:12px}
.plan-notes{white-space:pre-wrap}
.plan-form .checkbox{color:var(--hm-text)!important}
.plan-card .progress{margin:0}
@media(max-width:450px){.plan-grid,.plan-summary{grid-template-columns:1fr}}
</style>
