package controllers

import (
 "fmt"
 "io"
 "net/http"
 "os"
 "strings"

 "github.com/akhilrex/hammond/db"
 "github.com/gin-gonic/gin"
 "gorm.io/driver/sqlite"
 "gorm.io/gorm"
)

// The input is treated as untrusted data: it is opened read-only, never executed,
// never copied over the live database, and removed after each request.
const legacyDBMaxSize int64 = 50 << 20

type legacyVehiclePreview struct {
 ID string `json:"id"`
 SourceIDs []string `json:"sourceIds"`
 DuplicateCount int `json:"duplicateCount"`
 Name string `json:"name"`
 Registration string `json:"registration"`
 Fillups int `json:"fillups"`
 Expenses int `json:"expenses"`
 Matched bool `json:"matched"`
}
type legacyPreview struct {
 Vehicles []legacyVehiclePreview `json:"vehicles"`
 TotalFillups int `json:"totalFillups"`
 TotalExpenses int `json:"totalExpenses"`
 ExistingVehicles int `json:"existingVehicles"`
 SourceVehicleCount int `json:"sourceVehicleCount"`
 GroupedDuplicates int `json:"groupedDuplicates"`
 Note string `json:"note"`
}
type legacyData struct {
 Vehicles []db.Vehicle
 Fillups []db.Fillup
 Expenses []db.Expense
}

func RegisteHammondImportController(router *gin.RouterGroup) {
 router.POST("/import/hammond/preview", func(c *gin.Context) { processHammondImport(c, false) })
 router.POST("/import/hammond/execute", func(c *gin.Context) { processHammondImport(c, true) })
}
func readLegacyDatabase(c *gin.Context) (*gorm.DB, func(), error) {
 c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, legacyDBMaxSize+(1<<20))
 f, err := c.FormFile("file")
 if err != nil {
  if strings.Contains(err.Error(), "request body too large") {
   return nil,nil,fmt.Errorf("Die hochgeladene Datei überschreitet das Limit von 50 MB")
  }
  if strings.Contains(err.Error(), "multipart") || strings.Contains(err.Error(), "no such file") {
   return nil,nil,fmt.Errorf("Die Datenbank wurde nicht als Datei übertragen (Upload-Format ungültig). Bitte erneut auswählen. Ursache: %v",err)
  }
  return nil,nil,fmt.Errorf("Datenbankdatei konnte nicht gelesen werden: %v",err)
 }
 if f.Size <= 0 || f.Size > legacyDBMaxSize { return nil,nil,fmt.Errorf("Die SQLite-Datei darf höchstens 50 MB groß sein") }
 input, err := f.Open()
 if err != nil { return nil,nil,fmt.Errorf("Datei konnte nicht gelesen werden") }
 defer input.Close()
 target, err := os.CreateTemp("", "hammond-import-*.db")
 if err != nil { return nil,nil,fmt.Errorf("Temporäre Datei konnte nicht erstellt werden") }
 path := target.Name()
 cleanup := func(){ os.Remove(path) }
 // SQLite files must have a valid header. Refuse arbitrary files early.
 header := make([]byte,16)
 if _,err = io.ReadFull(input,header); err != nil || string(header)!="SQLite format 3\x00" {
  target.Close(); cleanup(); return nil,nil,fmt.Errorf("Das ist keine gültige SQLite-Datenbank")
 }
 if _,err = target.Write(header); err == nil { _,err = io.Copy(target,io.LimitReader(input,legacyDBMaxSize-int64(len(header)))) }
 if closeErr := target.Close(); err == nil { err = closeErr }
 if err != nil { cleanup(); return nil,nil,fmt.Errorf("Datei konnte nicht verarbeitet werden") }
 source, err := gorm.Open(sqlite.Open("file:"+path+"?mode=ro"), &gorm.Config{DisableAutomaticPing:false})
 if err != nil { cleanup(); return nil,nil,fmt.Errorf("Datenbank konnte nicht geöffnet werden") }
 sqlDB,err:=source.DB()
 if err != nil {cleanup();return nil,nil,fmt.Errorf("Datenbankverbindung fehlgeschlagen")}
 sqlDB.SetMaxOpenConns(1)
 wrappedCleanup:=func(){sqlDB.Close();cleanup()}
 for _,table:=range []string{"vehicles","fillups","expenses"} {
  if !source.Migrator().HasTable(table) {wrappedCleanup();return nil,nil,fmt.Errorf("Die Tabelle %s fehlt – bitte eine Original-Hammond-Datenbank auswählen",table)}
 }
 return source,wrappedCleanup,nil
}
func loadLegacy(source *gorm.DB) (*legacyData,error) {
 data:=&legacyData{}
 // Raw SELECT * is intentional: newer Hammond Modern fields can be absent in
 // older Hammond databases. GORM maps the columns that are actually present.
 if err:=source.Raw("SELECT * FROM vehicles WHERE deleted_at IS NULL").Scan(&data.Vehicles).Error;err!=nil{return nil,fmt.Errorf("Fahrzeuge nicht lesbar: %w",err)}
 if err:=source.Raw("SELECT * FROM fillups WHERE deleted_at IS NULL").Scan(&data.Fillups).Error;err!=nil{return nil,fmt.Errorf("Tankungen nicht lesbar: %w",err)}
 if err:=source.Raw("SELECT * FROM expenses WHERE deleted_at IS NULL").Scan(&data.Expenses).Error;err!=nil{return nil,fmt.Errorf("Ausgaben nicht lesbar: %w",err)}
 if len(data.Vehicles)>1000||len(data.Fillups)>200000||len(data.Expenses)>200000 {return nil,fmt.Errorf("Die Sicherung enthält zu viele Einträge")}
 ids:=map[string]bool{}
 for _,v:=range data.Vehicles{ if v.ID=="" {return nil,fmt.Errorf("Ungültige Fahrzeug-ID in der Quelldatenbank")}; ids[v.ID]=true }
 for _,f:=range data.Fillups{if !ids[f.VehicleID] || f.ID=="" {return nil,fmt.Errorf("Tankung ohne gültiges Fahrzeug erkannt")}}
 for _,e:=range data.Expenses{if !ids[e.VehicleID] || e.ID=="" {return nil,fmt.Errorf("Ausgabe ohne gültiges Fahrzeug erkannt")}}
 return data,nil
}
func sameVehicle(a,b db.Vehicle)bool {
 // Require full identity to prevent accidental merging of unrelated vehicles.
 return strings.EqualFold(strings.TrimSpace(a.Nickname),strings.TrimSpace(b.Nickname)) &&
 strings.EqualFold(strings.TrimSpace(a.Registration),strings.TrimSpace(b.Registration)) &&
 strings.EqualFold(strings.TrimSpace(a.Make),strings.TrimSpace(b.Make)) &&
 strings.EqualFold(strings.TrimSpace(a.Model),strings.TrimSpace(b.Model))
}
func processHammondImport(c *gin.Context,execute bool) {
 userID:=c.MustGet("userId").(string)
 source,cleanup,err:=readLegacyDatabase(c)
 if err!=nil{c.JSON(422,gin.H{"error":err.Error()});return}
 defer cleanup()
 data,err:=loadLegacy(source)
 if err!=nil{c.JSON(422,gin.H{"error":err.Error()});return}
 var mine []db.Vehicle
 err=db.DB.Model(&db.Vehicle{}).Joins("JOIN user_vehicles ON user_vehicles.vehicle_id=vehicles.id").Where("user_vehicles.user_id=?",userID).Find(&mine).Error
 if err!=nil{c.JSON(500,gin.H{"error":"Ziel-Fahrzeuge konnten nicht geladen werden"});return}
 // Several original Hammond rows can represent the same car. Group only when
 // all identity fields match, retaining every source ID so history is preserved.
 groups:=make([]legacyVehiclePreview,0,len(data.Vehicles))
 representatives:=make([]db.Vehicle,0,len(data.Vehicles))
 sourceGroups:=make(map[string]int,len(data.Vehicles))
 for _,v:=range data.Vehicles{
  groupIndex:=-1
  for i,representative:=range representatives{
   if sameVehicle(v,representative) && strings.TrimSpace(v.Registration)!="" {
    groupIndex=i;break
   }
  }
  if groupIndex<0{
   groupIndex=len(groups)
   groups=append(groups,legacyVehiclePreview{ID:v.ID,SourceIDs:[]string{},Name:v.Nickname,Registration:v.Registration})
   representatives=append(representatives,v)
  }
  groups[groupIndex].SourceIDs=append(groups[groupIndex].SourceIDs,v.ID)
  sourceGroups[v.ID]=groupIndex
 }
 for _,f:=range data.Fillups{groups[sourceGroups[f.VehicleID]].Fillups++}
 for _,e:=range data.Expenses{groups[sourceGroups[e.VehicleID]].Expenses++}
 preview:=legacyPreview{
  Vehicles:groups,TotalFillups:len(data.Fillups),TotalExpenses:len(data.Expenses),
  SourceVehicleCount:len(data.Vehicles),GroupedDuplicates:len(data.Vehicles)-len(groups),
  Note:"Benutzer/Passwörter, Dateianhänge und alte Erinnerungen werden nicht importiert.",
 }
 for i:=range groups{
  groups[i].DuplicateCount=len(groups[i].SourceIDs)-1
  for _,existing:=range mine{
   if sameVehicle(representatives[i],existing){groups[i].Matched=true;preview.ExistingVehicles++;break}
  }
 }
 if !execute {c.JSON(200,preview);return}
 if c.PostForm("confirm")!="IMPORTIEREN" {c.JSON(422,gin.H{"error":"Bitte den Import ausdrücklich bestätigen"});return}
 // The preview selects group IDs, not individual historical database rows.
 selectedIDs:=c.PostFormArray("vehicleIds")
 if len(selectedIDs)==0{c.JSON(422,gin.H{"error":"Bitte mindestens ein Fahrzeug auswählen"});return}
 selectedGroups:=map[int]bool{}
 for _,id:=range selectedIDs{
  found:=false
  for i,g:=range groups{if g.ID==id{selectedGroups[i]=true;found=true;break}}
  if !found{c.JSON(422,gin.H{"error":"Ungültige Fahrzeugauswahl; Vorschau erneut laden"});return}
 }
 if len(selectedGroups)==0{c.JSON(422,gin.H{"error":"Kein Fahrzeug ausgewählt"});return}
 var currentUser db.User
 if err=db.DB.First(&currentUser,"id=?",userID).Error;err!=nil{c.JSON(500,gin.H{"error":"Benutzer nicht gefunden"});return}
 tx:=db.DB.Begin()
 if tx.Error!=nil{c.JSON(500,gin.H{"error":"Transaktion konnte nicht gestartet werden"});return}
 defer tx.Rollback()
 mapped:=map[string]string{}
 newVehicles,newFillups,newExpenses,skippedFillups,skippedExpenses:=0,0,0,0,0
 for i,group:=range groups{
  if !selectedGroups[i]{continue}
  v:=representatives[i]
  targetID:=""
  for _,existing:=range mine{
   if sameVehicle(v,existing){targetID=existing.ID;break}
  }
  if targetID==""{
   cloned:=db.Vehicle{Nickname:v.Nickname,Registration:v.Registration,Make:v.Make,Model:v.Model,YearOfManufacture:v.YearOfManufacture,EngineSize:v.EngineSize,FuelUnit:v.FuelUnit,FuelType:v.FuelType}
   if err=tx.Create(&cloned).Error;err!=nil{c.JSON(500,gin.H{"error":"Fahrzeugimport fehlgeschlagen"});return}
   if err=tx.Create(&db.UserVehicle{UserID:userID,VehicleID:cloned.ID,IsOwner:true}).Error;err!=nil{c.JSON(500,gin.H{"error":"Fahrzeugzuordnung fehlgeschlagen"});return}
   targetID=cloned.ID;newVehicles++
  }
  for _,originalID:=range group.SourceIDs{mapped[originalID]=targetID}
 }
 for _,f:=range data.Fillups{
  if mapped[f.VehicleID]==""{continue}
  var count int64
  key:="Hammond:"+f.ID
  if err=tx.Model(&db.Fillup{}).Where("source = ? OR id = ?",key,f.ID).Count(&count).Error;err!=nil{c.JSON(500,gin.H{"error":"Duplikatprüfung fehlgeschlagen"});return}
  if count>0 {skippedFillups++;continue}
  clone:=db.Fillup{VehicleID:mapped[f.VehicleID],FuelUnit:f.FuelUnit,FuelQuantity:f.FuelQuantity,PerUnitPrice:f.PerUnitPrice,TotalAmount:f.TotalAmount,OdoReading:f.OdoReading,IsTankFull:f.IsTankFull,HasMissedFillup:f.HasMissedFillup,Comments:f.Comments,FillingStation:f.FillingStation,UserID:userID,Date:f.Date,Currency:f.Currency,DistanceUnit:f.DistanceUnit,Source:key,FuelSubType:f.FuelSubType}
  if clone.Currency==""{clone.Currency=currentUser.Currency}
  if err=tx.Create(&clone).Error;err!=nil{c.JSON(500,gin.H{"error":"Tankungsimport fehlgeschlagen"});return}
  newFillups++
 }
 for _,e:=range data.Expenses{
  if mapped[e.VehicleID]==""{continue}
  var count int64
  key:="Hammond:"+e.ID
  if err=tx.Model(&db.Expense{}).Where("source = ? OR id = ?",key,e.ID).Count(&count).Error;err!=nil{c.JSON(500,gin.H{"error":"Duplikatprüfung fehlgeschlagen"});return}
  if count>0 {skippedExpenses++;continue}
  clone:=db.Expense{VehicleID:mapped[e.VehicleID],Amount:e.Amount,OdoReading:e.OdoReading,Comments:e.Comments,ExpenseType:e.ExpenseType,Workshop:e.Workshop,WorkPerformed:e.WorkPerformed,UserID:userID,Date:e.Date,Currency:e.Currency,DistanceUnit:e.DistanceUnit,Source:key}
  if clone.Currency==""{clone.Currency=currentUser.Currency}
  if err=tx.Create(&clone).Error;err!=nil{c.JSON(500,gin.H{"error":"Ausgabenimport fehlgeschlagen"});return}
  newExpenses++
 }
 if err=tx.Commit().Error;err!=nil{c.JSON(500,gin.H{"error":"Import konnte nicht abgeschlossen werden"});return}
 c.JSON(200,gin.H{"vehicles":newVehicles,"fillups":newFillups,"expenses":newExpenses,"skippedFillups":skippedFillups,"skippedExpenses":skippedExpenses})
}
