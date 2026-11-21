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

func tireFailure(c *gin.Context,err error) {
	switch {
	case errors.Is(err,errTireVehicle): c.JSON(403,gin.H{"error":errTireVehicle.Error()})
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
		c.JSON(200,items)
	})
	group.POST("",func(c *gin.Context){saveTire(c,false)})
	group.PUT("/:id",func(c *gin.Context){saveTire(c,true)})
	group.DELETE("/:id",func(c *gin.Context){
		res:=db.DB.Where("id = ? AND owner_user_id = ?",c.Param("id"),c.GetString("userId")).Delete(&db.Tire{})
		if res.Error!=nil {tireFailure(c,res.Error);return}
		if res.RowsAffected!=1 {tireFailure(c,gorm.ErrRecordNotFound);return}
		c.Status(http.StatusNoContent)
	})
}

func saveTire(c *gin.Context, editing bool) {
	c.Request.Body=http.MaxBytesReader(c.Writer,c.Request.Body,32768)
	var input db.TireDetails
	if err:=c.ShouldBindJSON(&input);err!=nil {c.JSON(422,gin.H{"error":"Ungültige Reifendaten"});return}
	if err:=validateTire(&input);err!=nil {c.JSON(422,gin.H{"error":err.Error()});return}
	userID:=c.GetString("userId")
	var item db.Tire
	err:=db.DB.Transaction(func(tx *gorm.DB) error{
		if editing {
			if err:=tx.Where("id = ? AND owner_user_id = ?",c.Param("id"),userID).First(&item).Error;err!=nil{return err}
		}
		access,err:=tireAccess(tx,userID,input.VehicleID)
		if err!=nil{return err}
		if !access{return errTireVehicle}
		if editing {
			item.TireDetails=input
			// Select explicitly so clearing optional fields and false flags persists.
			res:=tx.Model(&db.Tire{}).Where("id = ? AND owner_user_id = ?",item.ID,userID).
				Select("vehicle_id","manufacturer","model","season","size","dot","set_name","load_index","speed_index","run_flat","status","position","tread_depth","pressure","purchase_date","purchase_price","currency","mileage","storage","notes","updated_at").Updates(&item)
			if res.Error!=nil{return res.Error}
			if res.RowsAffected!=1{return gorm.ErrRecordNotFound}
			return tx.Where("id = ? AND owner_user_id = ?",item.ID,userID).First(&item).Error
		}
		var id [16]byte
		if _,err=rand.Read(id[:]);err!=nil{return err}
		item=db.Tire{ID:hex.EncodeToString(id[:]),OwnerUserID:userID,TireDetails:input}
		return tx.Create(&item).Error
	})
	if err!=nil {tireFailure(c,err);return}
	status:=http.StatusCreated
	if editing {status=http.StatusOK}
	c.JSON(status,item)
}
