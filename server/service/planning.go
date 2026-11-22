package service

import (
 "errors"
 "fmt"
 "log"
 "os"
 "sync"
 "time"
 "github.com/akhilrex/hammond/db"
 "gorm.io/gorm"
)

var planningMu sync.Mutex

// Month-end dates stay anchored (Jan 31 -> Feb 28 -> Mar 31).
func PlanningMonth(date time.Time, months int) time.Time {
 first:=time.Date(date.Year(),date.Month()+time.Month(months),1,12,0,0,0,date.Location())
 day:=date.Day()
 if last:=first.AddDate(0,1,-1).Day();day>last{day=last}
 return first.AddDate(0,0,day-1)
}
func PlanningAccess(tx *gorm.DB,userID,vehicleID string)(bool,error){
 var count int64
 if userID==""||vehicleID==""{return false,nil}
 err:=tx.Model(&db.UserVehicle{}).Where("user_id = ? AND vehicle_id = ? AND deleted_at IS NULL",userID,vehicleID).Count(&count).Error
 if err!=nil||count==0{return false,err}
 err=tx.Model(&db.Vehicle{}).Where("id = ? AND deleted_at IS NULL",vehicleID).Count(&count).Error
 return count>0,err
}
func PlanningOdo(tx *gorm.DB,vehicleID string,now time.Time)(*int,error){
 var best *int
 for _,model:=range []interface{}{&db.Fillup{},&db.Expense{}}{
  var result struct{Odo *int}
  err:=tx.Model(model).Select("MAX(odo_reading) AS odo").Where("vehicle_id = ? AND deleted_at IS NULL AND distance_unit = ? AND odo_reading > 0 AND odo_reading <= 10000000 AND date(date) <= ?",vehicleID,db.KILOMETERS,now.Format("2006-01-02")).Scan(&result).Error
  if err!=nil{return nil,err}
  if result.Odo!=nil&&(best==nil||*result.Odo>*best){best=result.Odo}
 }
 return best,nil
}
func CostDue(plan db.RecurringCost,index int)(time.Time,error){
 start,err:=time.ParseInLocation("2006-01-02",plan.StartDate,time.Local)
 if err!=nil{return time.Time{},err}
 if index<0||index>12000||plan.IntervalMonths<1||plan.IntervalMonths>12{return time.Time{},errors.New("invalid schedule")}
 return PlanningMonth(start,index*plan.IntervalMonths),nil
}
func CostAmount(plan db.RecurringCost,index int)int64{
 if plan.Installments>0&&index==plan.Installments-1&&plan.FinalCents!=nil{return *plan.FinalCents}
 return plan.AmountCents
}
func CostRemaining(plan db.RecurringCost)(int,*int64){
 if plan.Installments<=0{return 0,nil}
 n:=plan.Installments-plan.Generated
 if n<0{n=0}
 total:=int64(n)*plan.AmountCents
 if n>0&&plan.FinalCents!=nil{total+=*plan.FinalCents-plan.AmountCents}
 return n,&total
}
// One payment and its counter commit together. Even concurrent workers cannot
// create duplicates; a failed expense insert rolls back the counter.
func bookCost(tx *gorm.DB,id string,now time.Time)(bool,error){
 var plan db.RecurringCost
 if err:=tx.Where("id = ? AND deleted_at IS NULL",id).First(&plan).Error;err!=nil{return false,err}
 if plan.Stopped||(plan.Installments>0&&plan.Generated>=plan.Installments){return false,nil}
 if plan.AmountCents<=0||plan.AmountCents>100000000||plan.Generated<0||plan.Generated>=12000{return false,errors.New("invalid cost plan")}
 var user db.User
 if err:=tx.Where("id = ? AND deleted_at IS NULL AND is_disabled = ?",plan.OwnerUserID,false).First(&user).Error;err!=nil{
  if errors.Is(err,gorm.ErrRecordNotFound){return false,nil};return false,err
 }
 allowed,err:=PlanningAccess(tx,plan.OwnerUserID,plan.VehicleID)
 if err!=nil||!allowed{return false,err}
 due,err:=CostDue(plan,plan.Generated)
 if err!=nil{return false,err}
 if due.Format("2006-01-02")>now.Format("2006-01-02"){return false,nil}
 amount:=CostAmount(plan,plan.Generated)
 if amount<=0||amount>100000000{return false,errors.New("invalid payment amount")}
 update:=tx.Model(&db.RecurringCost{}).Where("id = ? AND generated = ? AND stopped = ? AND deleted_at IS NULL",plan.ID,plan.Generated,false).Updates(map[string]interface{}{"generated":plan.Generated+1,"booked_cents":plan.BookedCents+amount})
 if update.Error!=nil{return false,update.Error}
 if update.RowsAffected!=1{return false,nil}
 exp:=db.Expense{VehicleID:plan.VehicleID,UserID:plan.OwnerUserID,Amount:float32(amount)/100,Currency:plan.Currency,DistanceUnit:user.DistanceUnit,Date:due,ExpenseType:plan.Category,Comments:fmt.Sprintf("%s – automatische Sollbuchung, Rate %d (kein Zahlungsnachweis)",plan.Title,plan.Generated+1),Source:"RecurringCost"}
 if err=tx.Create(&exp).Error;err!=nil{return false,err}
 return true,nil
}
func BookRecurringCosts(owner string,now time.Time)(int,error){
 planningMu.Lock();defer planningMu.Unlock()
 var plans []db.RecurringCost
 query:=db.DB.Where("stopped = ? AND deleted_at IS NULL AND (installments = 0 OR generated < installments)",false)
 if owner!=""{query=query.Where("owner_user_id = ?",owner)}
 if err:=query.Find(&plans).Error;err!=nil{return 0,err}
 booked:=0
 var firstErr error
 for _,plan:=range plans{
  for i:=0;i<120;i++{
   created:=false
   err:=db.DB.Transaction(func(tx *gorm.DB)error{var e error;created,e=bookCost(tx,plan.ID,now);return e})
   if err!=nil{if firstErr==nil{firstErr=err};log.Printf("recurring cost booking failed: %T",err);break}
   if !created{break}
   booked++
  }
 }
 return booked,firstErr
}
func RunRecurringCosts(){
 PortableDataLock.RLock();defer PortableDataLock.RUnlock()
 if _,err:=os.Stat(portableJournalPath(os.Getenv("CONFIG")));!os.IsNotExist(err){return}
 if _,err:=BookRecurringCosts("",time.Now());err!=nil{log.Printf("recurring costs job failed: %T",err)}
}
