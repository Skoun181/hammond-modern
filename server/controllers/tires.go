package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/akhilrex/hammond/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func tireOneOf(value string, allowed ...string) bool {
	for _, item := range allowed { if value == item { return true } }
	return false
}

func validateTire(t *db.TireDetails) error {
	for _, field := range []*string{&t.VehicleID,&t.Manufacturer,&t.Model,&t.Season,&t.Size,&t.DOT,&t.SetName,&t.LoadIndex,&t.SpeedIndex,&t.Status,&t.Position,&t.PurchaseDate,&t.Currency,&t.Storage,&t.Notes} {
		*field = strings.TrimSpace(*field)
	}
	if t.Manufacturer == "" || t.Size == "" { return errors.New("Hersteller und Reifengröße sind erforderlich") }
	for _, field := range []string{t.VehicleID,t.Manufacturer,t.Model,t.Size,t.SetName,t.Storage} {
		if utf8.RuneCountInString(field) > 150 { return errors.New("Eingabe zu lang") }
	}
	if utf8.RuneCountInString(t.Notes) > 2000 { return errors.New("Notizen sind zu lang") }
	if !tireOneOf(t.Season,"summer","winter","all-season","other") { return errors.New("Ungültige Reifenart") }
	if !tireOneOf(t.Status,"stored","mounted","retired") { return errors.New("Ungültiger Reifenstatus") }
	if !tireOneOf(t.Position,"","front-left","front-right","rear-left","rear-right","spare","other") { return errors.New("Ungültige Einbauposition") }
	if t.Status == "mounted" && (t.VehicleID == "" || t.Position == "") { return errors.New("Montierte Reifen benötigen Fahrzeug und Position") }
	if t.Status != "mounted" { t.Position = "" }
	if t.DOT != "" {
		if len(t.DOT) != 4 { return errors.New("DOT bitte als vierstelligen Produktionscode WWJJ eingeben") }
		week, e1 := strconv.Atoi(t.DOT[:2])
		year, e2 := strconv.Atoi(t.DOT[2:])
		if e1 != nil || e2 != nil || week < 1 || week > 53 || year < 0 || year > 99 || !tireDigits(t.DOT) { return errors.New("Ungültiger DOT-Code (WWJJ)") }
	}
	if len(t.LoadIndex)>12 || len(t.SpeedIndex)>12 { return errors.New("Traglast- oder Geschwindigkeitsindex zu lang") }
	if t.PurchaseDate != "" { if _,err:=time.Parse("2006-01-02",t.PurchaseDate);err!=nil { return errors.New("Ungültiges Kaufdatum") } }
	for _,v:=range []struct{number *float64; maximum float64}{{t.TreadDepth,30},{t.Pressure,20},{t.PurchasePrice,1000000}} {
		if v.number != nil && (math.IsNaN(*v.number)||math.IsInf(*v.number,0)||*v.number<0||*v.number>v.maximum) { return errors.New("Messwert oder Kaufpreis außerhalb des zulässigen Bereichs") }
	}
	if t.Mileage!=nil && (*t.Mileage<0||*t.Mileage>10000000) { return errors.New("Ungültige Reifenlaufleistung") }
	t.Currency=strings.ToUpper(t.Currency)
	if t.PurchasePrice!=nil {
		if len(t.Currency)!=3 {return errors.New("Währung zum Kaufpreis angeben (z. B. EUR)")}
		for _,r:=range t.Currency { if r<'A'||r>'Z' {return errors.New("Ungültige Währung")} }
	} else {t.Currency=""}
	return nil
}

func tireDigits(s string) bool { for _,r:=range s {if r<'0'||r>'9' {return false}};return true }

func tireAccess(tx *gorm.DB,userID,vehicleID string) (bool,error) {
	if vehicleID=="" {return true,nil}
	var count int64
	if err:=tx.Model(&db.Vehicle{}).Where("id = ?",vehicleID).Count(&count).Error;err!=nil {return false,err}
	if count==0 {return false,nil}
	err:=tx.Model(&db.UserVehicle{}).Where("user_id = ? AND vehicle_id = ?",userID,vehicleID).Count(&count).Error
	return count>0,err
}

var errTireVehicle = errors.New("Fahrzeug nicht verfügbar oder nicht freigegeben")
var errTireOccupied = errors.New("Eine gewählte Radposition ist bereits belegt. Lagere den montierten Reifen zuerst ein oder ändere seine Position.")

func tireFailure(c *gin.Context,err error) {
	switch {
	case errors.Is(err,errTireVehicle): c.JSON(403,gin.H{"error":errTireVehicle.Error()})
	case errors.Is(err,errTireMileage): c.JSON(422,gin.H{"error":err.Error()})
	case errors.Is(err,errTireOccupied): c.JSON(409,gin.H{"error":errTireOccupied.Error()})
	case errors.Is(err,gorm.ErrRecordNotFound): c.JSON(404,gin.H{"error":"Reifen nicht gefunden"})
	default: log.Printf("tire operation failed: %T",err);c.JSON(500,gin.H{"error":"Reifen konnten nicht verarbeitet werden"})
	}
}

func RegisterTireController(r *gin.RouterGroup) {
	group:=r.Group("/tires")
	group.Use(func(c *gin.Context){
		user,ok:=c.Get("userModel")
		model,valid:=user.(db.User)
		if !ok||!valid||model.ID==""||model.IsDisabled||model.ID!=c.GetString("userId") {
			c.AbortWithStatusJSON(403,gin.H{"error":"Anmeldung erforderlich"});return
		}
		c.Header("Cache-Control","no-store")
		c.Next()
	})
	group.GET("",func(c *gin.Context){
		items:=[]db.Tire{}
		if err:=db.DB.Where("owner_user_id = ?",c.GetString("userId")).Order("created_at DESC").Find(&items).Error;err!=nil {tireFailure(c,err);return}
		views,err:=tireMileageViews(db.DB,c.GetString("userId"),items)
		if err!=nil {tireFailure(c,err);return}
		c.JSON(200,views)
	})
	group.POST("/change",changeTireSet)
	group.POST("",func(c *gin.Context){saveTire(c,false)})
	group.PUT("/:id",func(c *gin.Context){saveTire(c,true)})
	group.DELETE("/:id",func(c *gin.Context){
		err:=db.DB.Transaction(func(tx *gorm.DB) error {
			res:=tx.Where("id = ? AND owner_user_id = ?",c.Param("id"),c.GetString("userId")).Delete(&db.Tire{})
			if res.Error!=nil {return res.Error}
			if res.RowsAffected!=1 {return gorm.ErrRecordNotFound}
			return tx.Where("tire_id = ?",c.Param("id")).Delete(&db.TireMileageState{}).Error
		})
		if err!=nil {tireFailure(c,err);return}
		c.Status(http.StatusNoContent)
	})
}

// tireSaveRequest keeps the original single-item API compatible. Quantity and
// positions are create-only instructions, never persisted on a tire record.
type tireSaveRequest struct {
	db.TireDetails
	Quantity *int     `json:"quantity"`
	Positions []string `json:"positions"`
	ChangeOdo *int64 `json:"changeOdo"`
	ChangeDate string `json:"changeDate"`
}

func tireCreateDetails(request tireSaveRequest, editing bool) ([]db.TireDetails, error) {
	quantity := 1
	if request.Quantity != nil { quantity = *request.Quantity }
	if quantity < 1 || quantity > 4 {
		return nil, errors.New("Bitte eine Anzahl von 1 bis 4 Reifen auswählen")
	}
	if editing && (quantity != 1 || len(request.Positions) != 0) {
		return nil, errors.New("Beim Bearbeiten kann nur ein Reifen geändert werden")
	}
	base := request.TireDetails
	if quantity == 1 {
		if len(request.Positions) != 0 { return nil, errors.New("Für einen Reifen bitte eine einzelne Einbauposition angeben") }
		if err := validateTire(&base); err != nil { return nil, err }
		return []db.TireDetails{base}, nil
	}

	// A mounted batch uses distinct wheel positions; stored/retired tires have
	// no position. Per-tire measurements and price are copied, not divided.
	base.Position = ""
	if strings.TrimSpace(base.Status) == "mounted" {
		if len(request.Positions) != quantity {
			return nil, errors.New("Bitte genau so viele Radpositionen auswählen wie Reifen angelegt werden")
		}
	} else if len(request.Positions) != 0 {
		return nil, errors.New("Einbaupositionen sind nur für montierte Reifen möglich")
	}
	seen := map[string]bool{}
	items := make([]db.TireDetails, 0, quantity)
	for i := 0; i < quantity; i++ {
		item := base
		if strings.TrimSpace(base.Status) == "mounted" {
			item.Position = strings.TrimSpace(request.Positions[i])
			if !tireOneOf(item.Position, "front-left", "front-right", "rear-left", "rear-right") || seen[item.Position] {
				return nil, errors.New("Bitte unterschiedliche Radpositionen auswählen")
			}
			seen[item.Position] = true
		}
		if err := validateTire(&item); err != nil { return nil, err }
		items = append(items, item)
	}
	return items, nil
}

func checkTirePosition(tx *gorm.DB, userID, ignoreID string, input db.TireDetails) error {
	if input.Status != "mounted" || !tireOneOf(input.Position, "front-left", "front-right", "rear-left", "rear-right") {
		return nil
	}
	var count int64
	if err := tx.Model(&db.Tire{}).Where("owner_user_id = ? AND vehicle_id = ? AND status = ? AND position = ? AND id <> ?", userID, input.VehicleID, "mounted", input.Position, ignoreID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 { return errTireOccupied }
	return nil
}

func saveTire(c *gin.Context, editing bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32768)
	var request tireSaveRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(422, gin.H{"error":"Ungültige Reifendaten"}); return
	}
	inputs, err := tireCreateDetails(request, editing)
	if err != nil { c.JSON(422, gin.H{"error":err.Error()}); return }
	userID := c.GetString("userId")
	input := inputs[0]
	var item db.Tire
	created := make([]db.Tire, 0, len(inputs))
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if editing {
			if err := tx.Where("id = ? AND owner_user_id = ?", c.Param("id"), userID).First(&item).Error; err != nil { return err }
		}
		access, err := tireAccess(tx, userID, input.VehicleID)
		if err != nil { return err }
		if !access { return errTireVehicle }
		if editing {
			// Existing imports may contain historical duplicate positions; allow
			// editing their other fields, but do not introduce a new conflict.
			if input.VehicleID != item.VehicleID || input.Status != item.Status || input.Position != item.Position {
				if err := checkTirePosition(tx, userID, item.ID, input); err != nil { return err }
			}
			if err := updateTireMileage(tx, &item, &input, request.ChangeOdo, request.ChangeDate); err != nil { return err }
			item.TireDetails = input
			res := tx.Model(&db.Tire{}).Where("id = ? AND owner_user_id = ?", item.ID, userID).
				Select("vehicle_id","manufacturer","model","season","size","dot","set_name","load_index","speed_index","run_flat","status","position","tread_depth","pressure","purchase_date","purchase_price","currency","mileage","storage","notes","updated_at").Updates(&item)
			if res.Error != nil { return res.Error }
			if res.RowsAffected != 1 { return gorm.ErrRecordNotFound }
			return tx.Where("id = ? AND owner_user_id = ?", item.ID, userID).First(&item).Error
		}
		// All inserts use the same transaction: a conflict on the last wheel
		// rolls back the earlier inserts too.
		for _, details := range inputs {
			if err := checkTirePosition(tx, userID, "", details); err != nil { return err }
			var id [16]byte
			if _, err := rand.Read(id[:]); err != nil { return err }
			newItem := db.Tire{ID:hex.EncodeToString(id[:]), OwnerUserID:userID, TireDetails:details}
			if err := tx.Create(&newItem).Error; err != nil { return err }
			if details.Status == "mounted" {
				if err := beginTireMileage(tx, newItem.ID, details.Position, request.ChangeOdo, request.ChangeDate); err != nil { return err }
			}
			created = append(created, newItem)
		}
		return nil
	})
	if err != nil { tireFailure(c,err); return }
	if editing { c.JSON(http.StatusOK,item); return }
	if len(created) == 1 { c.JSON(http.StatusCreated,created[0]); return }
	c.JSON(http.StatusCreated,gin.H{"count":len(created),"tires":created})
}
