package controllers
import (
 "fmt"
 "net/http"
 "net/url"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "time"
 "github.com/akhilrex/hammond/db"
 "github.com/akhilrex/hammond/service"
)
var reminderMu sync.Mutex
func reminderEndpoint(raw string)(string,error){
 u,err:=url.Parse(strings.TrimSpace(raw))
 if err!=nil||u.Hostname()==""||(u.Scheme!="https"&&u.Scheme!="http")||u.User!=nil||u.RawQuery!=""||u.Fragment!=""||strings.Trim(u.Path,"/")==""{return "",fmt.Errorf("invalid ntfy endpoint")}
 return u.String(),nil
}
func reminderStage(severity string)int{
 switch severity{case "soon":return 1;case "due","overdue":return 2};return 0
}
func sendMaintenanceReminder(endpoint,token string,stage int)error{
 text:="Eine aktivierte Fahrzeugwartung wird demnächst fällig. Bitte Hammond Modern öffnen."
 if stage==2{text="Eine aktivierte Fahrzeugwartung ist fällig oder überfällig. Bitte Hammond Modern öffnen."}
 req,err:=http.NewRequest(http.MethodPost,endpoint,strings.NewReader(text));if err!=nil{return fmt.Errorf("ntfy request failed")}
 req.Header.Set("Title","Hammond Modern")
 req.Header.Set("Content-Type","text/plain; charset=utf-8")
 if token!=""{req.Header.Set("Authorization","Bearer "+token)}
 client:=http.Client{Timeout:5*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}}
 res,err:=client.Do(req);if err!=nil{return fmt.Errorf("ntfy connection failed")}
 defer res.Body.Close()
 if res.StatusCode<200||res.StatusCode>=300{return fmt.Errorf("ntfy HTTP status %d",res.StatusCode)}
 return nil
}
// Operator-configured destination, per-task opt-in, generic payload.
// Persist successful warning/due stages. A crash after send can repeat.
func RunMaintenanceReminders(){
 endpoint,err:=reminderEndpoint(os.Getenv("HAMMOND_NTFY_URL"));if err!=nil{return}
 if !reminderMu.TryLock(){return};defer reminderMu.Unlock()
 service.PortableDataLock.RLock();defer service.PortableDataLock.RUnlock()
 if _,err:=os.Stat(filepath.Join(os.Getenv("CONFIG"),"restore-safety","in-progress.json"));!os.IsNotExist(err){return}
 var tasks []db.MaintenanceTask
 err=db.DB.Model(&db.MaintenanceTask{}).Joins("JOIN maintenance_rules ON maintenance_rules.task_id = maintenance_tasks.id").Where("maintenance_tasks.completed_at IS NULL AND maintenance_tasks.deleted_at IS NULL AND maintenance_rules.notify = ? AND maintenance_rules.notification_stage < 2",true).Find(&tasks).Error
 if err!=nil{return}
 sent:=0
 for _,task:=range tasks{
  if sent>=10{return}
  var user db.User
  if err=db.DB.Where("id = ? AND deleted_at IS NULL AND is_disabled = ?",task.OwnerUserID,false).First(&user).Error;err!=nil{continue}
  access,err:=service.PlanningAccess(db.DB,user.ID,task.VehicleID);if err!=nil||!access{continue}
  rule,err:=getMaintenanceRule(db.DB,task.ID);if err!=nil||!rule.Notify{continue}
  now:=time.Now()
  odo,err:=service.PlanningOdo(db.DB,task.VehicleID,now);if err!=nil{continue}
  stage:=reminderStage(maintenanceStatus(task,rule,odo,now).Severity)
  if stage<=rule.NotificationStage{continue}
  if err=sendMaintenanceReminder(endpoint,os.Getenv("HAMMOND_NTFY_TOKEN"),stage);err!=nil{return}
  sent++
  if err=db.DB.Model(&db.MaintenanceRule{}).Where("task_id = ? AND notify = ? AND notification_stage < ?",task.ID,true,stage).Update("notification_stage",stage).Error;err!=nil{return}
 }
}
