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
 if err != nil { return nil,nil,fmt.Errorf("Bitte die Originaldatei hammond.db auswählen") }
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
 preview:=legacyPreview{Vehicles:make([]legacyVehiclePreview,0,len(data.Vehicles)),TotalFillups:len(data.Fillups),TotalExpenses:len(data.Expenses),Note:"Benutzer/Passwörter, Dateianhänge und alte Erinnerungen werden nicht importiert."}
 for _,v:=range data.Vehicles{
  item:=legacyVehiclePreview{ID:v.ID,Name:v.Nickname,Registration:v.Registration}
  for _,current:=range mine {if sameVehicle(v,current){item.Matched=true;preview.ExistingVehicles++;break}}
  for _,f:=range data.Fillups {if f.VehicleID==v.ID {item.Fillups++}}
  for _,e:=range data.Expenses {if e.VehicleID==v.ID {item.Expenses++}}
  preview.Vehicles=append(preview.Vehicles,item)
 }
 if !execute {c.JSON(200,preview);return}
 if c.PostForm("confirm")!="IMPORTIEREN" {c.JSON(422,gin.H{"error":"Bitte den Import ausdrücklich bestätigen"});return}
 if len(data.Vehicles)==0 {c.JSON(422,gin.H{"error":"Keine Fahrzeuge zum Importieren gefunden"});return}
 var currentUser db.User
 if err=db.DB.First(&currentUser,"id=?",userID).Error;err!=nil{c.JSON(500,gin.H{"error":"Benutzer nicht gefunden"});return}
 tx:=db.DB.Begin()
 if tx.Error!=nil{c.JSON(500,gin.H{"error":"Transaktion konnte nicht gestartet werden"});return}
 defer tx.Rollback()
 mapped:=map[string]string{}
 newVehicles,newFillups,newExpenses,skippedFillups,skippedExpenses:=0,0,0,0,0
 for _,v:=range data.Vehicles{
  for _,existing:=range mine {if sameVehicle(v,existing){mapped[v.ID]=existing.ID;break}}
  if mapped[v.ID]!=""{continue}
  cloned:=db.Vehicle{Nickname:v.Nickname,Registration:v.Registration,Make:v.Make,Model:v.Model,YearOfManufacture:v.YearOfManufacture,EngineSize:v.EngineSize,FuelUnit:v.FuelUnit,FuelType:v.FuelType}
  if err=tx.Create(&cloned).Error;err!=nil{c.JSON(500,gin.H{"error":"Fahrzeugimport fehlgeschlagen"});return}
  if err=tx.Create(&db.UserVehicle{UserID:userID,VehicleID:cloned.ID,IsOwner:true}).Error;err!=nil{c.JSON(500,gin.H{"error":"Fahrzeugzuordnung fehlgeschlagen"});return}
  mapped[v.ID]=cloned.ID;newVehicles++
 }
 for _,f:=range data.Fillups{
  var count int64
  key:="Hammond:"+f.ID
  if err=tx.Model(&db.Fillup{}).Where("source=?",key).Count(&count).Error;err!=nil{c.JSON(500,gin.H{"error":"Duplikatprüfung fehlgeschlagen"});return}
  if count>0 {skippedFillups++;continue}
  clone:=db.Fillup{VehicleID:mapped[f.VehicleID],FuelUnit:f.FuelUnit,FuelQuantity:f.FuelQuantity,PerUnitPrice:f.PerUnitPrice,TotalAmount:f.TotalAmount,OdoReading:f.OdoReading,IsTankFull:f.IsTankFull,HasMissedFillup:f.HasMissedFillup,Comments:f.Comments,FillingStation:f.FillingStation,UserID:userID,Date:f.Date,Currency:f.Currency,DistanceUnit:f.DistanceUnit,Source:key,FuelSubType:f.FuelSubType}
  if clone.Currency==""{clone.Currency=currentUser.Currency}
  if err=tx.Create(&clone).Error;err!=nil{c.JSON(500,gin.H{"error":"Tankungsimport fehlgeschlagen"});return}
  newFillups++
 }
 for _,e:=range data.Expenses{
  var count int64
  key:="Hammond:"+e.ID
  if err=tx.Model(&db.Expense{}).Where("source=?",key).Count(&count).Error;err!=nil{c.JSON(500,gin.H{"error":"Duplikatprüfung fehlgeschlagen"});return}
  if count>0 {skippedExpenses++;continue}
  clone:=db.Expense{VehicleID:mapped[e.VehicleID],Amount:e.Amount,OdoReading:e.OdoReading,Comments:e.Comments,ExpenseType:e.ExpenseType,Workshop:e.Workshop,WorkPerformed:e.WorkPerformed,UserID:userID,Date:e.Date,Currency:e.Currency,DistanceUnit:e.DistanceUnit,Source:key}
  if clone.Currency==""{clone.Currency=currentUser.Currency}
  if err=tx.Create(&clone).Error;err!=nil{c.JSON(500,gin.H{"error":"Ausgabenimport fehlgeschlagen"});return}
  newExpenses++
 }
 if err=tx.Commit().Error;err!=nil{c.JSON(500,gin.H{"error":"Import konnte nicht abgeschlossen werden"});return}
 c.JSON(200,gin.H{"vehicles":newVehicles,"fillups":newFillups,"expenses":newExpenses,"skippedFillups":skippedFillups,"skippedExpenses":skippedExpenses})
}
