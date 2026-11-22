package service

import (
 "errors"
 "math"
 "time"
 "github.com/akhilrex/hammond/db"
 "gorm.io/gorm"
)
var ErrFinancingTerms = errors.New("Ungültige Finanzierungsdaten oder bereits festgeschriebene Anzahlung")
func ValidFinancingTerms(t db.FinancingTerms,now time.Time)bool{
 if t.DownPaymentCents<0||t.DownPaymentCents>100000000{return false}
 if t.DownPaymentCents>0{
  d,err:=time.Parse("2006-01-02",t.DownPaymentDate)
  if err!=nil||d.Year()<1900||t.DownPaymentDate>now.Format("2006-01-02"){return false}
 }else if t.DownPaymentDate!=""||t.AlreadyRecorded{return false}
 if t.PrincipalCents!=nil&&(*t.PrincipalCents<=0||*t.PrincipalCents>100000000){return false}
 for _,r:=range []*int{t.NominalBps,t.EffectiveBps}{if r!=nil&&(*r<0||*r>10000){return false}}
 return true
}
// Caller owns transaction. Metadata and expense commit together.
func SaveFinancingTerms(tx *gorm.DB,plan db.RecurringCost,user db.User,t db.FinancingTerms,now time.Time)error{
 if plan.ID==""||plan.OwnerUserID!=user.ID||plan.Category!="Finanzierung"||!ValidFinancingTerms(t,now){return ErrFinancingTerms}
 t.PlanID=plan.ID
 var old db.FinancingTerms
 err:=tx.Where("plan_id = ?",plan.ID).First(&old).Error
 exists:=err==nil
 if err!=nil&&!errors.Is(err,gorm.ErrRecordNotFound){return err}
 if exists&&old.DownPaymentCents>0&&(old.DownPaymentCents!=t.DownPaymentCents||old.DownPaymentDate!=t.DownPaymentDate||old.AlreadyRecorded!=t.AlreadyRecorded){return ErrFinancingTerms}
 if exists{
  if err=tx.Model(&db.FinancingTerms{}).Where("plan_id = ?",plan.ID).Select("down_payment_cents","down_payment_date","already_recorded","principal_cents","nominal_bps","effective_bps").Updates(&t).Error;err!=nil{return err}
 }else if err=tx.Create(&t).Error;err!=nil{return err}
 if old.DownPaymentCents>0||t.DownPaymentCents==0||t.AlreadyRecorded{return nil}
 d,err:=time.ParseInLocation("2006-01-02",t.DownPaymentDate,time.Local)
 if err!=nil{return err}
 exp:=db.Expense{VehicleID:plan.VehicleID,UserID:user.ID,Amount:float32(t.DownPaymentCents)/100,Currency:plan.Currency,DistanceUnit:user.DistanceUnit,Date:time.Date(d.Year(),d.Month(),d.Day(),12,0,0,0,time.Local),ExpenseType:"Finanzierung",Comments:plan.Title+" – einmalige Anzahlung (Sollbuchung)",Source:"FinancingDownPayment"}
 return tx.Create(&exp).Error
}
// Constant nominal annual rate / 12, end-of-month payments; no special
// repayments, fees or actual bank settlement. APR is informational.
func FinancingSummary(p db.RecurringCost,t *db.FinancingTerms)(total,creditCost,balance *int64){
 if p.Category!="Finanzierung"||p.Installments<1||p.Installments>600{return}
 sum:=int64(p.Installments)*p.AmountCents
 if p.FinalCents!=nil{sum+=*p.FinalCents-p.AmountCents}
 all:=sum
 if t!=nil{all+=t.DownPaymentCents}
 total=&all
 if t==nil||t.PrincipalCents==nil{return}
 cost:=sum-*t.PrincipalCents
 creditCost=&cost
 if t.NominalBps==nil||p.IntervalMonths!=1{return}
 b:=float64(*t.PrincipalCents)
 rate:=float64(*t.NominalBps)/120000
 for i:=0;i<p.Generated&&i<p.Installments;i++{
  b=b*(1+rate)-float64(CostAmount(p,i))
  if b<0{b=0}
  if math.IsInf(b,0)||math.IsNaN(b)||b>1e15{return}
 }
 v:=int64(math.Round(b));balance=&v
 return
}
