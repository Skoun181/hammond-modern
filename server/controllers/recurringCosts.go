package controllers

import (
 "net/http"
 "strings"
 "time"
 "github.com/akhilrex/hammond/db"
 "github.com/akhilrex/hammond/service"
 "github.com/gin-gonic/gin"
)
type costInput struct{
 VehicleID string `json:"vehicleId"`
 Title string `json:"title"`
 Category string `json:"category"`
 AmountCents int64 `json:"amountCents"`
 StartDate string `json:"startDate"`
 IntervalMonths int `json:"intervalMonths"`
 Installments int `json:"installments"`
 FinalCents *int64 `json:"finalCents"`
 Notes string `json:"notes"`
}
type costView struct{
 db.RecurringCost
 NextDate string `json:"nextDate"`
 NextCents int64 `json:"nextCents"`
 RemainingInstallments int `json:"remainingInstallments"`
 RemainingCents *int64 `json:"remainingCents"`
}
func costOutput(plan db.RecurringCost)costView{
 n,total:=service.CostRemaining(plan)
 v:=costView{RecurringCost:plan,RemainingInstallments:n,RemainingCents:total}
 if !plan.Stopped&&(plan.Installments==0||n>0){if d,e:=service.CostDue(plan,plan.Generated);e==nil{v.NextDate=d.Format("2006-01-02");v.NextCents=service.CostAmount(plan,plan.Generated)}}
 return v
}
func validCost(in *costInput,now time.Time)bool{
 in.Title=strings.TrimSpace(in.Title);in.Category=strings.TrimSpace(in.Category);in.Notes=strings.TrimSpace(in.Notes)
 start,err:=time.Parse("2006-01-02",in.StartDate)
 if err!=nil||in.StartDate<now.Format("2006-01-02")||start.Year()>now.Year()+20{return false}
 if in.VehicleID==""||in.Title==""||len(in.Title)>120||len(in.Notes)>2000||!tireOneOf(in.Category,"Finanzierung","Leasing","Versicherung","Kfz-Steuer","Stellplatz","Abonnement","Sonstiges"){return false}
 if in.AmountCents<=0||in.AmountCents>100000000||in.Installments<0||in.Installments>600||!(in.IntervalMonths==1||in.IntervalMonths==3||in.IntervalMonths==6||in.IntervalMonths==12){return false}
 if in.Category=="Finanzierung"&&(in.Installments<1||in.IntervalMonths!=1){return false}
 if in.FinalCents!=nil&&(in.Installments<1||*in.FinalCents<1||*in.FinalCents>100000000){return false}
 return true
}
func planningAuth(c *gin.Context){
 u,ok:=c.Get("userModel");user,valid:=u.(db.User)
 if !ok||!valid||user.ID==""||user.IsDisabled||user.ID!=c.GetString("userId"){c.AbortWithStatusJSON(403,gin.H{"error":"Anmeldung erforderlich"});return}
 c.Header("Cache-Control","no-store");c.Next()
}
func RegisterRecurringCostController(router *gin.RouterGroup){
 group:=router.Group("/recurring-costs");group.Use(planningAuth)
 group.GET("",func(c *gin.Context){
  var plans []db.RecurringCost
  if err:=db.DB.Where("owner_user_id = ? AND deleted_at IS NULL",c.GetString("userId")).Order("created_at DESC").Find(&plans).Error;err!=nil{c.JSON(500,gin.H{"error":"Kostenpläne konnten nicht geladen werden"});return}
  views:=make([]costView,0,len(plans));for _,p:=range plans{views=append(views,costOutput(p))};c.JSON(200,views)
 })
 group.POST("",func(c *gin.Context){
  c.Request.Body=http.MaxBytesReader(c.Writer,c.Request.Body,16384)
  var in costInput
  if c.ShouldBindJSON(&in)!=nil||!validCost(&in,time.Now()){c.JSON(422,gin.H{"error":"Gültige Beträge, Kategorie, Intervall, nächste Fälligkeit ab heute und bei Finanzierung die Anzahl monatlicher Raten angeben"});return}
  user:=c.MustGet("userModel").(db.User)
  access,err:=service.PlanningAccess(db.DB,user.ID,in.VehicleID)
  if err!=nil{c.JSON(500,gin.H{"error":"Fahrzeugprüfung fehlgeschlagen"});return}
  if !access{c.JSON(403,gin.H{"error":"Fahrzeug nicht freigegeben"});return}
  if strings.TrimSpace(user.Currency)==""{c.JSON(422,gin.H{"error":"Bitte zuerst eine Währung in den Einstellungen festlegen"});return}
  plan:=db.RecurringCost{OwnerUserID:user.ID,VehicleID:in.VehicleID,Title:in.Title,Category:in.Category,AmountCents:in.AmountCents,Currency:user.Currency,StartDate:in.StartDate,IntervalMonths:in.IntervalMonths,Installments:in.Installments,FinalCents:in.FinalCents,Notes:in.Notes}
  if err=db.DB.Create(&plan).Error;err!=nil{c.JSON(500,gin.H{"error":"Kostenplan konnte nicht gespeichert werden"});return}
  c.JSON(201,costOutput(plan))
 })
 group.POST("/run",func(c *gin.Context){
  n,err:=service.BookRecurringCosts(c.GetString("userId"),time.Now())
  if err!=nil{c.JSON(500,gin.H{"error":"Einige Buchungen fehlgeschlagen. Bereits gebuchte Raten werden nicht erneut erzeugt","booked":n});return}
  c.JSON(200,gin.H{"booked":n})
 })
 group.POST("/:id/stop",func(c *gin.Context){
  result:=db.DB.Model(&db.RecurringCost{}).Where("id = ? AND owner_user_id = ? AND deleted_at IS NULL",c.Param("id"),c.GetString("userId")).Update("stopped",true)
  if result.Error!=nil{c.JSON(500,gin.H{"error":"Kostenplan konnte nicht gestoppt werden"});return}
  if result.RowsAffected!=1{c.Status(404);return};c.Status(204)
 })
}
