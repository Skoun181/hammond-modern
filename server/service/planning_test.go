package service

import (
 "testing"
 "time"
 "path/filepath"
 "github.com/akhilrex/hammond/db"
 "gorm.io/driver/sqlite"
 "gorm.io/gorm"
)
func TestRecurringFinancingBookings(t *testing.T){
 old:=db.DB
 g,err:=gorm.Open(sqlite.Open(filepath.Join(t.TempDir(),"test.sqlite")),&gorm.Config{DisableForeignKeyConstraintWhenMigrating:true})
 if err!=nil{t.Fatal(err)}
 db.DB=g
 t.Cleanup(func(){db.DB=old;sql,_:=g.DB();sql.Close()})
 if err=g.AutoMigrate(&db.User{},&db.Vehicle{},&db.UserVehicle{},&db.Expense{},&db.Fillup{},&db.RecurringCost{});err!=nil{t.Fatal(err)}
 user:=db.User{Email:"plan@example.test",Currency:"EUR",DistanceUnit:db.KILOMETERS}
 v:=db.Vehicle{Nickname:"Auto"}
 for _,x:=range []interface{}{&user,&v}{if err=g.Create(x).Error;err!=nil{t.Fatal(err)}}
 // The association must use IDs created by the DB hook.
 if err=g.Create(&db.UserVehicle{UserID:user.ID,VehicleID:v.ID,IsOwner:true}).Error;err!=nil{t.Fatal(err)}
 final:=int64(100000)
 p:=db.RecurringCost{OwnerUserID:user.ID,VehicleID:v.ID,Title:"Finanzierung",Category:"Finanzierung",Currency:"EUR",AmountCents:30000,StartDate:"2024-01-31",IntervalMonths:1,Installments:3,FinalCents:&final}
 if err=g.Create(&p).Error;err!=nil{t.Fatal(err)}
 for index,want:=range []string{"2024-01-31","2024-02-29","2024-03-31"}{d,e:=CostDue(p,index);if e!=nil||d.Format("2006-01-02")!=want{t.Fatalf("date %v, %v",d,e)}}
 now:=time.Date(2024,2,29,13,0,0,0,time.Local)
 n,e:=BookRecurringCosts(user.ID,now);if e!=nil||n!=2{t.Fatalf("first book %d %v",n,e)}
 n,e=BookRecurringCosts(user.ID,now);if e!=nil||n!=0{t.Fatalf("duplicate %d %v",n,e)}
 if err=g.First(&p,"id = ?",p.ID).Error;err!=nil{t.Fatal(err)}
 left,amount:=CostRemaining(p);if left!=1||*amount!=100000||p.BookedCents!=60000{t.Fatalf("balance: %#v",p)}
 n,e=BookRecurringCosts(user.ID,now.AddDate(0,2,0));if e!=nil||n!=1{t.Fatalf("final %d %v",n,e)}
 var expenses []db.Expense
 if err=g.Where("vehicle_id = ?",v.ID).Order("date").Find(&expenses).Error;err!=nil{t.Fatal(err)}
 if len(expenses)!=3||expenses[2].Amount!=1000||expenses[0].Amount!=300{t.Fatalf("expenses %#v",expenses)}
 if err=g.First(&p,"id = ?",p.ID).Error;err!=nil{t.Fatal(err)}
 if p.Generated!=3||p.BookedCents!=160000{t.Fatalf("progress %#v",p)}
 // An existing financing can acknowledge its first 17 already recorded
 // installments. The worker must only generate the missing 18th and future ones.
 historical:=db.RecurringCost{OwnerUserID:user.ID,VehicleID:v.ID,Title:"Altvertrag",Category:"Finanzierung",Currency:"EUR",AmountCents:30000,StartDate:"2025-06-01",IntervalMonths:1,Installments:60,Generated:17,BookedCents:510000}
 historyNow:=time.Date(2026,11,22,12,0,0,0,time.Local)
 count,e:=CostDueCount(historical,historyNow);if e!=nil||count!=18{t.Fatalf("historical count %d %v",count,e)}
 if err=g.Create(&historical).Error;err!=nil{t.Fatal(err)}
 n,e=BookRecurringCosts(user.ID,historyNow);if e!=nil||n!=1{t.Fatalf("history book %d %v",n,e)}
 n,e=BookRecurringCosts(user.ID,historyNow);if e!=nil||n!=0{t.Fatalf("history duplicate %d %v",n,e)}
 if err=g.First(&historical,"id = ?",historical.ID).Error;err!=nil{t.Fatal(err)}
 if historical.Generated!=18||historical.BookedCents!=540000{t.Fatalf("history progress %#v",historical)}
 next,e:=CostDue(historical,historical.Generated);if e!=nil||next.Format("2006-01-02")!="2026-12-01"{t.Fatalf("next %v %v",next,e)}

}
