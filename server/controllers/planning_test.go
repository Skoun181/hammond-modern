package controllers
import (
 "testing"
 "time"
 "path/filepath"
 "errors"
 "github.com/akhilrex/hammond/db"
 "gorm.io/driver/sqlite"
 "gorm.io/gorm"
)
func TestMaintenanceDueAndRepeat(t *testing.T){
 g,err:=gorm.Open(sqlite.Open(filepath.Join(t.TempDir(),"maintenance.sqlite")),&gorm.Config{DisableForeignKeyConstraintWhenMigrating:true})
 if err!=nil{t.Fatal(err)}
 raw,_:=g.DB();defer raw.Close()
 if err=g.AutoMigrate(&db.Vehicle{},&db.UserVehicle{},&db.MaintenanceTask{},&db.MaintenanceRule{});err!=nil{t.Fatal(err)}
 v:=db.Vehicle{Nickname:"Test"};if err=g.Create(&v).Error;err!=nil{t.Fatal(err)}
 if err=g.Create(&db.UserVehicle{UserID:"owner",VehicleID:v.ID,IsOwner:true}).Error;err!=nil{t.Fatal(err)}
 now:=time.Date(2024,1,31,12,0,0,0,time.Local);due:=now.AddDate(0,0,45);odo:=15000
 task:=db.MaintenanceTask{OwnerUserID:"owner",VehicleID:v.ID,Title:"Service",DueDate:&due,DueOdo:&odo}
 if err=g.Create(&task).Error;err!=nil{t.Fatal(err)}
 rule:=db.MaintenanceRule{TaskID:task.ID,RepeatMonths:1,RepeatKm:15000,ReminderDays:30,ReminderKm:1000}
 if err=g.Create(&rule).Error;err!=nil{t.Fatal(err)}
 current:=14500
 view:=maintenanceStatus(task,rule,&current,now)
 if view.Severity!="soon"||*view.KmLeft!=500{t.Fatalf("status %#v",view)}
 current=15001
 if maintenanceStatus(task,rule,&current,now).Severity!="overdue"{t.Fatal("km overdue not detected")}
 in:=maintenanceCompletion{Date:"2024-01-31",Odo:&current}
 run:=func(user string)error{return g.Transaction(func(tx *gorm.DB)error{return completeMaintenance(tx,user,task.ID,in,now)})}
 if !errors.Is(run("other"),gorm.ErrRecordNotFound){t.Fatal("other owner accepted")}
 if err=run("owner");err!=nil{t.Fatal(err)}
 if !errors.Is(run("owner"),errMaintenanceCompleted){t.Fatal("completed twice")}
 var next []db.MaintenanceTask
 if err=g.Where("completed_at IS NULL").Find(&next).Error;err!=nil||len(next)!=1{t.Fatalf("next %#v %v",next,err)}
 if next[0].DueDate.Format("2006-01-02")!="2024-02-29"||*next[0].DueOdo!=30001{t.Fatalf("next %#v",next)}
}
