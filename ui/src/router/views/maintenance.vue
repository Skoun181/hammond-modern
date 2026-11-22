<script>
import Layout from '@layouts/main.vue'
import MaintenanceOverview from '@components/maintenanceOverview.vue'
import axios from 'axios'
import {mapState} from 'vuex'
const today=()=>{const d=new Date();return d.getFullYear()+'-'+String(d.getMonth()+1).padStart(2,'0')+'-'+String(d.getDate()).padStart(2,'0')}
const blank=(id='')=>({vehicleId:id,title:'',notes:'',dueDate:'',dueOdo:'',repeatMonths:0,repeatKm:0,reminderDays:30,reminderKm:1000,notify:false})
export default {
  page:{title:'Wartungscenter'},components:{Layout,MaintenanceOverview},
  data(){return {tasks:[],loading:false,saving:false,error:'',success:'',editingId:'',showForm:false,overviewKey:0,form:blank(),completion:null,completionDate:'',completionOdo:''}},
  computed:{...mapState('vehicles',['vehicles']),completedTasks(){return this.tasks.filter(t=>t.completedAt)}},
  watch:{'$route.query.neu'(v){if(v==='1')this.add()}},
  mounted(){
    this.$store.dispatch('vehicles/fetchVehicles').then(()=>{if(this.vehicles.length&&!this.form.vehicleId)this.form.vehicleId=this.vehicles[0].id}).catch(()=>{this.error='Fahrzeuge konnten nicht geladen werden.'})
    if(this.$route.query.neu==='1')this.add()
    this.loadTasks()
  },
  methods:{
    today,vehicleName(id){const v=this.vehicles.find(v=>v.id===id);return v?v.nickname:'Fahrzeug'},
    async loadTasks(){this.loading=true;try{const r=await axios.get('/api/maintenance');this.tasks=r.data||[]}catch(e){this.error='Wartungsplanung konnte nicht geladen werden.'}finally{this.loading=false}},
    scroll(){this.$nextTick(()=>{const el=this.$el.querySelector('#maintenance-form');if(el)el.scrollIntoView({behavior:'smooth',block:'start'})})},
    add(){this.editingId='';this.form=blank((this.vehicles[0]||{}).id||'');this.completion=null;this.showForm=true;this.scroll()},
    reset(){this.showForm=false;this.editingId='';this.form=blank((this.vehicles[0]||{}).id||'');if(this.$route.query.neu)this.$router.replace({name:'maintenance',query:{}}).catch(()=>{})},
    edit(t){this.completion=null;this.showForm=true;this.editingId=t.id;this.form={...blank(),vehicleId:t.vehicleId,title:t.title,notes:t.notes||'',dueDate:t.dueDate?t.dueDate.slice(0,10):'',dueOdo:t.dueOdo==null?'':t.dueOdo,...(t.rule||{})};this.scroll()},
    optional(v){return v===''||v==null?null:Number(v)},
    async save(){
      if(this.saving)return
      this.saving=true;this.error='';this.success=''
      const p=this.form
      try{const data={vehicleId:p.vehicleId,title:p.title,notes:p.notes,dueDate:p.dueDate?p.dueDate+'T12:00:00Z':null,dueOdo:this.optional(p.dueOdo),repeatMonths:Number(p.repeatMonths||0),repeatKm:Number(p.repeatKm||0),reminderDays:Number(p.reminderDays),reminderKm:Number(p.reminderKm),notify:!!p.notify}
        if(this.editingId)await axios.put('/api/maintenance/'+this.editingId,data);else await axios.post('/api/maintenance',data)
        this.reset();await this.loadTasks();this.overviewKey++;this.success='Wartung gespeichert.'
      }catch(e){this.error=(e.response&&e.response.data&&e.response.data.error)||'Speichern fehlgeschlagen.'}finally{this.saving=false}
    },
    complete(t){this.reset();this.completion=t;this.completionDate=today();this.completionOdo=t.currentOdo==null?'':t.currentOdo;this.scroll()},
    async saveComplete(){
      if(this.saving||!this.completion)return
      this.saving=true;this.error='';this.success=''
      try{const repeat=this.completion.rule&&(this.completion.rule.repeatMonths||this.completion.rule.repeatKm)
        await axios.post('/api/maintenance/'+this.completion.id+'/complete',{date:this.completionDate,odo:this.optional(this.completionOdo)})
        this.completion=null;await this.loadTasks();this.overviewKey++;this.success=repeat?'Wartung erledigt. Der nächste Termin wurde automatisch angelegt.':'Wartung erledigt.'
      }catch(e){this.error=(e.response&&e.response.data&&e.response.data.error)||'Abschließen fehlgeschlagen.'}finally{this.saving=false}
    },
    async remove(t){if(!window.confirm('Wartung wirklich löschen?'))return;try{await axios.delete('/api/maintenance/'+t.id);await this.loadTasks();this.overviewKey++}catch(e){this.error='Löschen fehlgeschlagen.'}},
  }
}
</script>

<template>
  <Layout>
    <div class="hm-page-heading"><div><p class="hm-eyebrow">FAHRZEUGPFLEGE</p><h1 class="title hm-page-title">Wartungscenter</h1><p class="hm-page-description">Erinnerungen nach Datum oder Kilometerstand. Wiederkehrende Wartungen werden nach dem Abschluss automatisch neu geplant.</p></div><b-button type="is-primary" @click="add">＋ Wartung planen</b-button></div>
    <b-notification v-if="error" type="is-danger" @close="error = ''">{{ error }}</b-notification>
    <b-notification v-if="success" type="is-success" @close="success = ''">{{ success }}</b-notification>
    <MaintenanceOverview :key="overviewKey" heading="Was steht als Nächstes an?" :show-actions="true" @edit="edit" @complete="complete" @remove="remove" />
    <div v-if="completion" id="maintenance-form" class="box maintenance-form">
      <h2 class="title is-5">{{ completion.title }} abschließen</h2>
      <form @submit.prevent="saveComplete">
        <b-field label="Tatsächliches Wartungsdatum *"><input v-model="completionDate" class="input" type="date" :max="today()" required /></b-field>
        <b-field label="Tatsächlicher Kilometerstand (km)" message="Letzten erfassten Stand am Tacho prüfen. Bei km-Intervall erforderlich."><b-input v-model="completionOdo" type="number" min="0" max="10000000" step="1" :required="!!(completion.rule && completion.rule.repeatKm)" /></b-field>
        <p v-if="completion.rule && (completion.rule.repeatMonths || completion.rule.repeatKm)" class="help">Der neue Termin wird ab dem tatsächlichen Wartungsdatum bzw. Kilometerstand berechnet.</p>
        <div class="buttons"><b-button type="is-primary" native-type="submit" :loading="saving">Erledigt speichern</b-button><b-button :disabled="saving" @click="completion = null">Abbrechen</b-button></div>
      </form>
    </div>
    <div v-if="showForm" id="maintenance-form" class="box maintenance-form">
      <h2 class="title is-5">{{ editingId ? 'Wartung bearbeiten' : 'Neue Wartung planen' }}</h2>
      <form @submit.prevent="save">
        <b-field label="Fahrzeug"><b-select v-model="form.vehicleId" required expanded><option v-for="v in vehicles" :key="v.id" :value="v.id">{{ v.nickname }} – {{ v.registration }}</option></b-select></b-field>
        <b-field label="Wartung / Termin"><b-input v-model="form.title" required maxlength="120" placeholder="z. B. Ölwechsel, HU / TÜV, Inspektion" /></b-field>
        <div class="columns"><div class="column"><b-field label="Fällig am"><input v-model="form.dueDate" class="input" type="date" /></b-field></div><div class="column"><b-field label="Fällig bei Kilometerstand (km)"><b-input v-model="form.dueOdo" type="number" min="0" max="10000000" step="1" /></b-field></div></div>
        <p class="help">Mindestens ein Kriterium erforderlich. Bei beiden gilt: das zuerst erreichte Kriterium löst die Erinnerung aus.</p>
        <div class="maintenance-panel"><h3 class="title is-6">Automatisch wiederholen</h3>
          <div class="columns"><div class="column"><b-field label="Alle Monate (0 = aus)"><b-input v-model="form.repeatMonths" type="number" min="0" max="120" step="1" required /></b-field></div><div class="column"><b-field label="Alle Kilometer (0 = aus)"><b-input v-model="form.repeatKm" type="number" min="0" max="1000000" step="1" required /></b-field></div></div>
          <p class="help">Für ein Intervall das entsprechende erste Fälligkeitsdatum bzw. den ersten Fälligkeits-Kilometerstand eintragen. Beispiel: 12 Monate / 15.000 km (nach Herstellervorgabe anpassen).</p>
        </div>
        <div class="maintenance-panel"><h3 class="title is-6">Vorwarnung im Dashboard</h3><div class="columns"><div class="column"><b-field label="Tage vorher"><b-input v-model="form.reminderDays" type="number" min="0" max="365" step="1" required /></b-field></div><div class="column"><b-field label="Kilometer vorher"><b-input v-model="form.reminderKm" type="number" min="0" max="1000000" step="1" required /></b-field></div></div></div>
        <b-field><b-checkbox v-model="form.notify">Zusätzlich per ntfy erinnern</b-checkbox></b-field><p class="help">Optional: Der Betreiber muss HAMMOND_NTFY_URL setzen. Aktivierte Push-Erinnerungen gehen an den gemeinsamen Kanal ohne Fahrzeug- oder Benutzerdetails.</p>
        <b-field label="Notizen"><b-input v-model="form.notes" type="textarea" maxlength="2000" /></b-field>
        <div class="buttons"><b-button type="is-primary" native-type="submit" :loading="saving">Wartung speichern</b-button><b-button :disabled="saving" @click="reset">Abbrechen</b-button></div>
      </form>
    </div>
    <h2 v-if="completedTasks.length" class="title is-5">Wartungshistorie ({{ completedTasks.length }})</h2>
    <div v-for="t in completedTasks" :key="t.id" class="box maintenance-history">
      <strong>{{ t.title }}</strong> · {{ vehicleName(t.vehicleId) }}<p class="help">Erledigt am {{ t.completedAt.slice(0,10).split('-').reverse().join('.') }}<span v-if="t.rule && t.rule.completedOdo !== null && t.rule.completedOdo !== undefined"> · {{ t.rule.completedOdo.toLocaleString('de-DE') }} km</span></p><b-button size="is-small" type="is-danger" outlined @click="remove(t)">Löschen</b-button>
    </div>
  </Layout>
</template>

<style scoped>
.maintenance-panel{padding:14px;border:1px solid var(--hm-border);border-radius:10px;background:var(--hm-input);margin:14px 0}
.maintenance-history strong{color:var(--hm-text)!important}
.maintenance-form .help{margin-bottom:10px}
.maintenance-form .buttons{margin-top:14px}
</style>
