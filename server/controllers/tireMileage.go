package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/akhilrex/hammond/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var errTireMileage = errors.New("Ungültiger Kilometerstand oder Reifenwechsel")
func mileageError(message string) error { return fmt.Errorf("%w: %s", errTireMileage, message) }

func mileageEvent(odo *int64, date string) (string,error) {
	if odo == nil || *odo < 0 || *odo > 10000000 {
		return "",mileageError("Bitte aktuellen Fahrzeug-Kilometerstand (0–10.000.000 km) eingeben")
	}
	today:=time.Now().Format("2006-01-02")
	date=strings.TrimSpace(date)
	if date=="" {date=today}
	if _,err:=time.Parse("2006-01-02",date);err!=nil || date>today || date<"1900-01-01" {
		return "",mileageError("Bitte ein gültiges Wechseldatum bis heute angeben")
	}
	return date,nil
}
func rollingTire(position string) bool { return position!="spare" }
func tireState(tx *gorm.DB,id string)(*db.TireMileageState,error){
	var state db.TireMileageState
	err:=tx.Where("tire_id = ?",id).First(&state).Error
	if errors.Is(err,gorm.ErrRecordNotFound){return nil,nil}
	return &state,err
}
func beginTireMileage(tx *gorm.DB,id,position string,odo *int64,date string) error{
	date,err:=mileageEvent(odo,date);if err!=nil{return err}
	if !rollingTire(position){return nil}
	return tx.Create(&db.TireMileageState{TireID:id,MountOdo:*odo,MountedOn:date}).Error
}
func mileageTotal(base *int64,delta int64)(*int64,error){
	value:=delta
	if base!=nil{value+=*base}
	if value<0||value>10000000{return nil,mileageError("Reifen-Gesamtlaufleistung außerhalb des zulässigen Bereichs")}
	return &value,nil
}

// Only a confirmed unmount odometer finalises the interval. No tanking is
// counted twice, and a later remount starts a new interval on the saved total.
func closeTireMileage(tx *gorm.DB,old db.Tire,odo *int64,date string)(*int64,error){
	day,err:=mileageEvent(odo,date);if err!=nil{return nil,err}
	state,err:=tireState(tx,old.ID);if err!=nil{return nil,err}
	if state==nil{return old.Mileage,nil} // legacy or spare; never invent distance
	if day<state.MountedOn||*odo<state.MountOdo{
		return nil,mileageError("Abmontage darf nicht vor der Montage oder unter deren Kilometerstand liegen")
	}
	total,err:=mileageTotal(old.Mileage,*odo-state.MountOdo);if err!=nil{return nil,err}
	if err=tx.Where("tire_id = ?",old.ID).Delete(&db.TireMileageState{}).Error;err!=nil{return nil,err}
	return total,nil
}
func updateTireMileage(tx *gorm.DB,old *db.Tire,input *db.TireDetails,odo *int64,date string) error {
	if old.Status=="mounted"&&input.Status=="mounted"&&old.VehicleID!=input.VehicleID{
		return mileageError("Reifen zuerst am bisherigen Fahrzeug abmontieren")
	}
	if old.Status=="mounted"&&input.Status!="mounted"{
		total,err:=closeTireMileage(tx,*old,odo,date);if err!=nil{return err}
		input.Mileage=total
		return nil
	}
	if old.Status!="mounted"&&input.Status=="mounted"{
		return beginTireMileage(tx,old.ID,input.Position,odo,date)
	}
	if old.Status=="mounted"&&input.Status=="mounted"{
		state,err:=tireState(tx,old.ID);if err!=nil{return err}
		if rollingTire(old.Position)!=rollingTire(input.Position){
			total,err:=closeTireMileage(tx,*old,odo,date);if err!=nil{return err}
			input.Mileage=total
			return beginTireMileage(tx,old.ID,input.Position,odo,date)
		}
		if state==nil&&rollingTire(input.Position)&&odo!=nil{
			return beginTireMileage(tx,old.ID,input.Position,odo,date)
		}
		if state!=nil{
			// Editing dimensions, notes or wheel position must not reset the
			// accumulated baseline or the current mounting interval.
			input.Mileage=old.Mileage
			if odo!=nil{return mileageError("Montage-Kilometerstand ist bereits gesetzt; für Änderungen zuerst abmontieren")}
		}
	}
	return nil
}

type tireMileageView struct {
	db.Tire
	CurrentMileage *int64 `json:"currentMileage"`
	MountOdo *int64 `json:"mountOdo"`
	MountedOn string `json:"mountedOn"`
	CurrentVehicleOdo *int64 `json:"currentVehicleOdo"`
	Tracking bool `json:"tracking"`
}

// A fresh read derives the current interval from the greatest recorded km
// odometer since mounting. Ignore earlier and future-dated readings, deleted
// records, and mile-based readings (no mixing). No per-tanking writes needed.
func tireMileageViews(tx *gorm.DB,userID string,items []db.Tire)([]tireMileageView,error){
	out:=make([]tireMileageView,0,len(items))
	type key struct{vehicle,day string}
	cache:=map[key]int64{}
	accessCache:=map[string]bool{}
	for _,item:=range items{
		view:=tireMileageView{Tire:item,CurrentMileage:item.Mileage}
		if item.Status=="mounted"&&rollingTire(item.Position){
			state,err:=tireState(tx,item.ID);if err!=nil{return nil,err}
			if state!=nil{
				view.MountOdo=&state.MountOdo;view.MountedOn=state.MountedOn
				access,known:=accessCache[item.VehicleID]
				if !known{access,err=tireAccess(tx,userID,item.VehicleID);if err!=nil{return nil,err};accessCache[item.VehicleID]=access}
				if access{
					k:=key{item.VehicleID,state.MountedOn}
					latest,ok:=cache[k]
					if !ok{
						var value struct{Odo int64}
						err=tx.Model(&db.Fillup{}).Select("COALESCE(MAX(odo_reading),0) AS odo").
							Where("vehicle_id = ? AND distance_unit = ? AND date(date) >= ? AND date(date) <= ? AND odo_reading BETWEEN 0 AND 10000000",item.VehicleID,db.KILOMETERS,state.MountedOn,time.Now().Format("2006-01-02")).Scan(&value).Error
						if err!=nil{return nil,err};latest=value.Odo;cache[k]=latest
					}
					if latest<state.MountOdo{latest=state.MountOdo}
					view.CurrentVehicleOdo=&latest
					view.CurrentMileage,err=mileageTotal(item.Mileage,latest-state.MountOdo)
					if err!=nil{return nil,err}
					view.Tracking=true
				}
			}
		}
		out=append(out,view)
	}
	return out,nil
}

type tireChangeRequest struct{
	VehicleID string `json:"vehicleId"`
	Odo *int64 `json:"odo"`
	Date string `json:"date"`
	// Each selected position replaces the mounted tire with the supplied stored
	// tire, or stores the mounted tire when the ID is empty.
	Positions map[string]string `json:"positions"`
}
func changeTireSet(c *gin.Context){
	c.Request.Body=http.MaxBytesReader(c.Writer,c.Request.Body,32768)
	var req tireChangeRequest
	if err:=c.ShouldBindJSON(&req);err!=nil{c.JSON(422,gin.H{"error":"Ungültiger Reifenwechsel"});return}
	day,err:=mileageEvent(req.Odo,req.Date)
	if err!=nil{tireFailure(c,err);return}
	if req.VehicleID==""||len(req.Positions)==0||len(req.Positions)>4{tireFailure(c,mileageError("Fahrzeug und mindestens eine Radposition auswählen"));return}
	userID:=c.GetString("userId")
	seen:=map[string]bool{}
	for position,id:=range req.Positions{
		if !tireOneOf(position,"front-left","front-right","rear-left","rear-right")|| (id!=""&&seen[id]){
			tireFailure(c,mileageError("Ungültige oder doppelte Radposition/Reifenauswahl"));return
		}
		seen[id]=true
	}
	err=db.DB.Transaction(func(tx *gorm.DB)error{
		access,err:=tireAccess(tx,userID,req.VehicleID);if err!=nil{return err};if !access{return errTireVehicle}
		// Validate every incoming tire before any old tire is moved to storage.
		// This also prevents an order-dependent rotation of an already mounted tire.
		for _,id:=range req.Positions{
			if id==""{continue}
			var count int64
			if err=tx.Model(&db.Tire{}).Where("id = ? AND owner_user_id = ? AND vehicle_id = ? AND status = ?",id,userID,req.VehicleID,"stored").Count(&count).Error;err!=nil{return err}
			if count!=1{return mileageError("Nur eingelagerte Reifen dieses Fahrzeugs können montiert werden")}
		}
		for position,id:=range req.Positions{
			var mounted []db.Tire
			if err=tx.Where("owner_user_id = ? AND vehicle_id = ? AND status = ? AND position = ?",userID,req.VehicleID,"mounted",position).Find(&mounted).Error;err!=nil{return err}
			if len(mounted)>1{return errTireOccupied}
			if len(mounted)==0&&id==""{return mileageError("An dieser Radposition ist kein Reifen montiert")}
			if len(mounted)==1{
				total,err:=closeTireMileage(tx,mounted[0],req.Odo,day);if err!=nil{return err}
				if err=tx.Model(&db.Tire{}).Where("id = ? AND owner_user_id = ?",mounted[0].ID,userID).Updates(map[string]interface{}{"status":"stored","position":"","mileage":total}).Error;err!=nil{return err}
			}
			if id==""{continue}
			var next db.Tire
			if err=tx.Where("id = ? AND owner_user_id = ? AND vehicle_id = ? AND status = ?",id,userID,req.VehicleID,"stored").First(&next).Error;err!=nil{return err}
			if err=beginTireMileage(tx,next.ID,position,req.Odo,day);err!=nil{return err}
			if err=tx.Model(&db.Tire{}).Where("id = ? AND owner_user_id = ?",next.ID,userID).Updates(map[string]interface{}{"status":"mounted","position":position}).Error;err!=nil{return err}
		}
		return nil
	})
	if err!=nil{tireFailure(c,err);return}
	c.JSON(200,gin.H{"changed":len(req.Positions)})
}
