package controllers

import (
 "errors"
 "net/http"
 "strings"
 "time"
 "github.com/akhilrex/hammond/db"
 "github.com/akhilrex/hammond/service"
 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
)

type maintenanceInput struct {
 VehicleID string `json:"vehicleId"`
 Title string `json:"title"`
 Notes string `json:"notes"`
 DueDate *time.Time `json:"dueDate"`
 DueOdo *int `json:"dueOdo"`
 RepeatMonths int `json:"repeatMonths"`
 RepeatKm int `json:"repeatKm"`
 ReminderDays *int `json:"reminderDays"`
 ReminderKm *int `json:"reminderKm"`
 Notify bool `json:"notify"`
}
type maintenanceView struct{
 db.MaintenanceTask
 Rule db.MaintenanceRule `json:"rule"`
 CurrentOdo *int `json:"currentOdo"`
 DaysLeft *int `json:"daysLeft"`
 KmLeft *int `json:"kmLeft"`
 Severity string `json:"severity"`
}
func defaultMaintenanceRule(id string)db.MaintenanceRule{return db.MaintenanceRule{TaskID:id,ReminderDays:30,ReminderKm:1000}}
func validMaintenance(in *maintenanceInput)bool{
 in.Title=strings.TrimSpace(in.Title)
 return in.VehicleID!=""&&in.Title!=""&&len(in.Title)<=120&&len(in.Notes)<=2000&&(in.DueDate!=nil||in.DueOdo!=nil)&&
  (in.DueDate==nil||(in.DueDate.Year()>=1900&&in.DueDate.Year()<=2200))&&
  (in.DueOdo==nil||(*in.DueOdo>=0&&*in.DueOdo<=10000000))&&
  in.RepeatMonths>=0&&in.RepeatMonths<=120&&in.RepeatKm>=0&&in.RepeatKm<=1000000&&
  (in.RepeatMonths==0||in.DueDate!=nil)&&(in.RepeatKm==0||in.DueOdo!=nil)&&
  (in.ReminderDays==nil||(*in.ReminderDays>=0&&*in.ReminderDays<=365))&&
  (in.ReminderKm==nil||(*in.ReminderKm>=0&&*in.ReminderKm<=1000000))
}
func getMaintenanceRule(tx *gorm.DB,id string)(db.MaintenanceRule,error){
 var r db.MaintenanceRule
 err:=tx.Where("task_id = ?",id).First(&r).Error
 if errors.Is(err,gorm.ErrRecordNotFound){return defaultMaintenanceRule(id),nil}
 return r,err
}
func setMaintenanceRule(tx *gorm.DB,id string,in maintenanceInput)error{
 r,err:=getMaintenanceRule(tx,id);if err!=nil{return err}
 r.RepeatMonths=in.RepeatMonths;r.RepeatKm=in.RepeatKm
 if in.ReminderDays!=nil{r.ReminderDays=*in.ReminderDays}
 if in.ReminderKm!=nil{r.ReminderKm=*in.ReminderKm}
 if in.Notify&&!r.Notify{r.NotificationStage=0}
 r.Notify=in.Notify
 return tx.Save(&r).Error
}
func maintenanceStatus(task db.MaintenanceTask,rule db.MaintenanceRule,odo *int,now time.Time)maintenanceView{
 v:=maintenanceView{MaintenanceTask:task,Rule:rule,CurrentOdo:odo,Severity:"planned"}
 if task.CompletedAt!=nil{v.Severity="done";return v}
 overdue,due,soon:=false,false,false
 if task.DueDate!=nil{
  day:=func(t time.Time)time.Time{return time.Date(t.Year(),t.Month(),t.Day(),0,0,0,0,time.UTC)}
  n:=int(day(*task.DueDate).Sub(day(now)).Hours()/24);v.DaysLeft=&n
  overdue=overdue||n<0;due=due||n==0;soon=soon||n<=rule.ReminderDays
 }
 if task.DueOdo!=nil&&odo!=nil{
  n:=*task.DueOdo-*odo;v.KmLeft=&n
  overdue=overdue||n<0;due=due||n==0;soon=soon||n<=rule.ReminderKm
 }
 if overdue{v.Severity="overdue"}else if due{v.Severity="due"}else if soon{v.Severity="soon"}
 return v
}
type maintenanceCompletion struct{
 Date string `json:"date"`
 Odo *int `json:"odo"`
}
var errMaintenanceInput=errors.New("Ungültige Wartungsdaten")
var errMaintenanceCompleted=errors.New("Wartung bereits erledigt")
func completeMaintenance(tx *gorm.DB,userID,id string,in maintenanceCompletion,now time.Time)error{
 var task db.MaintenanceTask
 if err:=tx.Where("id = ? AND owner_user_id = ? AND deleted_at IS NULL",id,userID).First(&task).Error;err!=nil{return err}
 ok,err:=service.PlanningAccess(tx,userID,task.VehicleID);if err!=nil{return err};if !ok{return errTireVehicle}
 if task.CompletedAt!=nil{return errMaintenanceCompleted}
 rule,err:=getMaintenanceRule(tx,id);if err!=nil{return err}
 date,err:=time.ParseInLocation("2006-01-02",in.Date,time.Local)
 if err!=nil||in.Date>now.Format("2006-01-02")||date.Year()<1900||(in.Odo!=nil&&(*in.Odo<0||*in.Odo>10000000))||(rule.RepeatKm>0&&(in.Odo==nil||*in.Odo+rule.RepeatKm>10000000)){return errMaintenanceInput}
 date=time.Date(date.Year(),date.Month(),date.Day(),12,0,0,0,time.Local)
 result:=tx.Model(&db.MaintenanceTask{}).Where("id = ? AND owner_user_id = ? AND completed_at IS NULL",id,userID).Update("completed_at",date)
 if result.Error!=nil{return result.Error};if result.RowsAffected!=1{return errMaintenanceCompleted}
 rule.CompletedOdo=in.Odo
 if err=tx.Save(&rule).Error;err!=nil{return err}
 if rule.RepeatMonths==0&&rule.RepeatKm==0{return nil}
 next:=db.MaintenanceTask{OwnerUserID:userID,VehicleID:task.VehicleID,Title:task.Title,Notes:task.Notes}
 if rule.RepeatMonths>0{d:=service.PlanningMonth(date,rule.RepeatMonths);next.DueDate=&d}
 if rule.RepeatKm>0{n:=*in.Odo+rule.RepeatKm;next.DueOdo=&n}
 if err=tx.Create(&next).Error;err!=nil{return err}
 rule.TaskID=next.ID;rule.CompletedOdo=nil;rule.NotificationStage=0
 return tx.Create(&rule).Error
}
func maintenanceError(c *gin.Context,err error){
 switch{
 case errors.Is(err,errTireVehicle):c.JSON(403,gin.H{"error":"Fahrzeug nicht freigegeben"})
 case errors.Is(err,gorm.ErrRecordNotFound):c.JSON(404,gin.H{"error":"Wartung nicht gefunden"})
 case errors.Is(err,errMaintenanceCompleted):c.JSON(409,gin.H{"error":"Wartung bereits erledigt"})
 case errors.Is(err,errMaintenanceInput):c.JSON(422,gin.H{"error":"Bitte gültige Wartungsdaten und bei km-Intervall den tatsächlichen Kilometerstand angeben"})
 default:c.JSON(500,gin.H{"error":"Wartung konnte nicht verarbeitet werden"})
 }
}
func RegisterMaintenanceController(router *gin.RouterGroup) {
 group:=router.Group("/maintenance");group.Use(planningAuth)
 group.GET("",func(c *gin.Context){
  userID:=c.GetString("userId")
  var tasks []db.MaintenanceTask
  if err:=db.DB.Where("owner_user_id = ? AND deleted_at IS NULL",userID).Order("created_at DESC").Find(&tasks).Error;err!=nil{maintenanceError(c,err);return}
  out:=make([]maintenanceView,0,len(tasks));cache:=map[string]*int{}
  now:=time.Now()
  for _,task:=range tasks{
   rule,err:=getMaintenanceRule(db.DB,task.ID);if err!=nil{maintenanceError(c,err);return}
   odo,known:=cache[task.VehicleID]
   if !known{
    access,err:=service.PlanningAccess(db.DB,userID,task.VehicleID);if err!=nil{maintenanceError(c,err);return}
    if access{odo,err=service.PlanningOdo(db.DB,task.VehicleID,now);if err!=nil{maintenanceError(c,err);return}}
    cache[task.VehicleID]=odo
   }
   out=append(out,maintenanceStatus(task,rule,odo,now))
  }
  c.JSON(200,out)
 })
 save:=func(c *gin.Context,editing bool){
  c.Request.Body=http.MaxBytesReader(c.Writer,c.Request.Body,16384)
  var in maintenanceInput
  if c.ShouldBindJSON(&in)!=nil||!validMaintenance(&in){maintenanceError(c,errMaintenanceInput);return}
  userID:=c.GetString("userId")
  var task db.MaintenanceTask
  err:=db.DB.Transaction(func(tx *gorm.DB)error{
   ok,err:=service.PlanningAccess(tx,userID,in.VehicleID);if err!=nil{return err};if !ok{return errTireVehicle}
   if editing{
    if err=tx.Where("id = ? AND owner_user_id = ? AND deleted_at IS NULL",c.Param("id"),userID).First(&task).Error;err!=nil{return err}
    if task.CompletedAt!=nil{return errMaintenanceCompleted}
   }
   task.OwnerUserID=userID;task.VehicleID=in.VehicleID;task.Title=in.Title;task.Notes=in.Notes;task.DueDate=in.DueDate;task.DueOdo=in.DueOdo
   if editing{err=tx.Save(&task).Error}else{err=tx.Create(&task).Error};if err!=nil{return err}
   return setMaintenanceRule(tx,task.ID,in)
  })
  if err!=nil{maintenanceError(c,err);return}
  if editing{c.JSON(200,task)}else{c.JSON(201,task)}
 }
 group.POST("",func(c *gin.Context){save(c,false)})
 group.PUT("/:id",func(c *gin.Context){save(c,true)})
 group.POST("/:id/complete",func(c *gin.Context){
  c.Request.Body=http.MaxBytesReader(c.Writer,c.Request.Body,16384)
  var in maintenanceCompletion
  if c.ShouldBindJSON(&in)!=nil{maintenanceError(c,errMaintenanceInput);return}
  err:=db.DB.Transaction(func(tx *gorm.DB)error{return completeMaintenance(tx,c.GetString("userId"),c.Param("id"),in,time.Now())})
  if err!=nil{maintenanceError(c,err);return};c.Status(204)
 })
 group.DELETE("/:id",func(c *gin.Context){
  err:=db.DB.Transaction(func(tx *gorm.DB)error{
   result:=tx.Where("id = ? AND owner_user_id = ?",c.Param("id"),c.GetString("userId")).Delete(&db.MaintenanceTask{})
   if result.Error!=nil{return result.Error};if result.RowsAffected!=1{return gorm.ErrRecordNotFound}
   return tx.Where("task_id = ?",c.Param("id")).Delete(&db.MaintenanceRule{}).Error
  })
  if err!=nil{maintenanceError(c,err);return};c.Status(204)
 })
}
