package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/akhilrex/hammond/db"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHomeAssistantKeyAndReadOnlyMaintenance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	original := db.DB
	g, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "ha.sqlite")), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil { t.Fatal(err) }
	db.DB = g
	t.Cleanup(func(){ db.DB = original; raw,_ := g.DB(); raw.Close() })
	if err=g.AutoMigrate(&db.User{}, &db.Vehicle{}, &db.UserVehicle{}, &db.MaintenanceTask{}, &db.MaintenanceRule{}, &db.Fillup{}, &db.Expense{}, &db.HomeAssistantKey{});err!=nil{t.Fatal(err)}
	a:=db.User{Email:"ha-a@example.test", Currency:"EUR", DistanceUnit:db.KILOMETERS}
	b:=db.User{Email:"ha-b@example.test", Currency:"EUR", DistanceUnit:db.KILOMETERS}
	if err=g.Create(&a).Error;err!=nil{t.Fatal(err)}
	if err=g.Create(&b).Error;err!=nil{t.Fatal(err)}
	car:=db.Vehicle{Nickname:"Golf"}
	if err=g.Create(&car).Error;err!=nil{t.Fatal(err)}
	if err=g.Create(&db.UserVehicle{UserID:a.ID,VehicleID:car.ID,IsOwner:true}).Error;err!=nil{t.Fatal(err)}
	today:=time.Now()
	late:=today.AddDate(0,0,-2)
	soon:=today.AddDate(0,0,7)
	for _,task:=range []db.MaintenanceTask{
		{OwnerUserID:a.ID,VehicleID:car.ID,Title:"Ölwechsel",DueDate:&late},
		{OwnerUserID:a.ID,VehicleID:car.ID,Title:"HU",DueDate:&soon},
		{OwnerUserID:b.ID,VehicleID:car.ID,Title:"Andere Person",DueDate:&late},
	}{
		if err=g.Create(&task).Error;err!=nil{t.Fatal(err)}
	}
	r:=gin.New()
	api:=r.Group("/api")
	RegisterHomeAssistantReadController(api)
	api.Use(func(c *gin.Context){
		id:=c.GetHeader("X-Test-Owner")
		user:=a;if id==b.ID{user=b}
		if id=="disabled"{user.IsDisabled=true}
		c.Set("userId",id);c.Set("userModel",user);c.Next()
	})
	RegisterHomeAssistantKeyController(api)
	call:=func(method,path,token,owner string,want int)*httptest.ResponseRecorder{
		t.Helper()
		req:=httptest.NewRequest(method,path,nil)
		if token!=""{req.Header.Set("Authorization",token)}
		if owner!=""{req.Header.Set("X-Test-Owner",owner)}
		w:=httptest.NewRecorder();r.ServeHTTP(w,req)
		if w.Code!=want{t.Fatalf("%s %s: got %d want %d: %s",method,path,w.Code,want,w.Body.String())}
		return w
	}
	apiPath:="/api/integrations/homeassistant/maintenance"
	keyPath:="/api/integrations/homeassistant/key"
	call("GET",apiPath,"", "",401)
	call("GET",apiPath,"Bearer invalid","",401)
	call("GET",keyPath,"", "",403)
	w:=call("POST",keyPath,"",a.ID,201)
	var response struct{Token string `json:"token"`}
	if err=json.Unmarshal(w.Body.Bytes(),&response);err!=nil||!strings.HasPrefix(response.Token,"hmha_"){t.Fatalf("invalid key %s",w.Body.String())}
	var stored db.HomeAssistantKey
	if err=g.Where("owner_user_id = ?",a.ID).First(&stored).Error;err!=nil||stored.TokenHash==response.Token||stored.TokenHash==""{t.Fatal("token stored in plaintext")}
	auth:="Bearer "+response.Token
	w=call("GET",apiPath,auth,"",200)
	var result struct {
		AccountID string `json:"account_id"`
		Summary homeAssistantSummary `json:"summary"`
		Tasks []homeAssistantTask `json:"tasks"`
	}
	if err=json.Unmarshal(w.Body.Bytes(),&result);err!=nil{t.Fatal(err)}
	if result.AccountID!=a.ID||result.Summary.Total!=2||result.Summary.Overdue!=1||result.Summary.Soon!=1||len(result.Tasks)!=2||result.Tasks[0].Title!="Ölwechsel"||result.Tasks[0].VehicleName!="Golf"{t.Fatalf("bad result: %#v",result)}
	if strings.Contains(w.Body.String(),"Andere Person")||strings.Contains(w.Body.String(),response.Token){t.Fatal("secret or another user's task disclosed")}
	call("GET",keyPath,"",b.ID,200)
	call("GET",apiPath,"Bearer "+strings.Repeat("A",48),"",401)
	w=call("POST",keyPath,"",a.ID,201)
	var rotated struct{Token string `json:"token"`}
	if err=json.Unmarshal(w.Body.Bytes(),&rotated);err!=nil||rotated.Token==response.Token{t.Fatal("rotation failed")}
	call("GET",apiPath,auth,"",401)
	call("GET",apiPath,"Bearer "+rotated.Token,"",200)
	call("DELETE",keyPath,"",a.ID,204)
	call("GET",apiPath,"Bearer "+rotated.Token,"",401)
	// Disabled accounts never grant maintenance access.
	w=call("POST",keyPath,"",a.ID,201)
	if err=json.Unmarshal(w.Body.Bytes(),&response);err!=nil{t.Fatal(err)}
	if err=g.Model(&db.User{}).Where("id = ?",a.ID).Update("is_disabled",true).Error;err!=nil{t.Fatal(err)}
	call("GET",apiPath,"Bearer "+response.Token,"",401)
}
