package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/akhilrex/hammond/db"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTireValidation(t *testing.T) {
	base:=db.TireDetails{Manufacturer:"  Michelin ", Size:"225/45 R17", Season:"winter", Status:"stored", DOT:"2325"}
	if err:=validateTire(&base);err!=nil {t.Fatal(err)}
	if base.Manufacturer!="Michelin" {t.Fatal("not trimmed")}
	for _,dot:=range []string{"0025","5425","2A25","123","23250","+125"} {
		item:=base;item.DOT=dot
		if validateTire(&item)==nil{t.Errorf("accepted invalid DOT %q",dot)}
	}
	item:=base;item.Status="mounted"
	if validateTire(&item)==nil {t.Fatal("mounted without vehicle or position")}
	item.VehicleID="vehicle-a";item.Position="front-left"
	if err:=validateTire(&item);err!=nil{t.Fatal(err)}
	negative:=-1.0;item.TreadDepth=&negative
	if validateTire(&item)==nil{t.Fatal("accepted negative tread depth")}
}

func TestTireCRUDAndAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	original:=db.DB
	g,err:=gorm.Open(sqlite.Open(filepath.Join(t.TempDir(),"test.sqlite")),&gorm.Config{DisableForeignKeyConstraintWhenMigrating:true})
	if err!=nil{t.Fatal(err)}
	db.DB=g
	t.Cleanup(func(){db.DB=original; raw,_:=g.DB(); if raw!=nil{raw.Close()}})
	// Minimal vehicle and share tables, including soft-delete columns, exercise
	// exactly the permission checks used by the API.
	for _,sql:=range []string{
		"CREATE TABLE vehicles (id TEXT PRIMARY KEY,deleted_at DATETIME)",
		"CREATE TABLE user_vehicles (id TEXT PRIMARY KEY,user_id TEXT,vehicle_id TEXT,deleted_at DATETIME)",
		"INSERT INTO vehicles (id) VALUES ('v-a'),('v-b')",
		"INSERT INTO user_vehicles (id,user_id,vehicle_id) VALUES ('share-a','a','v-a'),('share-b','b','v-b')",
	}{
		if err:=g.Exec(sql).Error;err!=nil{t.Fatal(err)}
	}
	if err:=g.AutoMigrate(&db.Tire{});err!=nil{t.Fatal(err)}
	r:=gin.New()
	api:=r.Group("/api")
	api.Use(func(c *gin.Context){var u db.User;u.ID=c.GetHeader("X-Test-User");u.IsDisabled=u.ID=="disabled";c.Set("userId",u.ID);c.Set("userModel",u);c.Next()})
	RegisterTireController(api)
	call:=func(method,path,user string,data interface{},want int)*httptest.ResponseRecorder{
		t.Helper()
		body,_:=json.Marshal(data)
		req:=httptest.NewRequest(method,path,bytes.NewReader(body));req.Header.Set("Content-Type","application/json");req.Header.Set("X-Test-User",user)
		w:=httptest.NewRecorder();r.ServeHTTP(w,req)
		if w.Code!=want{t.Fatalf("%s %s as %s: got %d want %d: %s",method,path,user,w.Code,want,w.Body.String())}
		return w
	}
	pressure:=2.5
	item:=db.TireDetails{Manufacturer:"Continental",Model:"Test",Size:"225/45 R17",Season:"winter",DOT:"2325",Status:"mounted",Position:"front-left",VehicleID:"v-a",Pressure:&pressure,RunFlat:true}
	call(http.MethodGet,"/api/tires","",nil,403)
	call(http.MethodGet,"/api/tires","disabled",nil,403)
	call(http.MethodPost,"/api/tires","b",item,403)
	w:=call(http.MethodPost,"/api/tires","a",item,201)
	var created db.Tire
	if err:=json.Unmarshal(w.Body.Bytes(),&created);err!=nil||created.ID==""{t.Fatalf("bad created tire: %s",w.Body.String())}
	if bytes.Contains(w.Body.Bytes(),[]byte("ownerUserId")) {t.Fatal("leaked owner id")}
	path:="/api/tires/"+created.ID
	call(http.MethodPut,path,"b",item,404)
	call(http.MethodDelete,path,"b",nil,404)
	w=call(http.MethodGet,"/api/tires","b",nil,200);if w.Body.String()!="[]"{t.Fatal("other user's tire disclosed")}
	item.VehicleID="v-b";call(http.MethodPut,path,"a",item,403)
	item.VehicleID="";item.Status="stored";item.Position="";item.DOT="";item.Pressure=nil;item.RunFlat=false
	w=call(http.MethodPut,path,"a",item,200)
	var updated db.Tire
	if err:=json.Unmarshal(w.Body.Bytes(),&updated);err!=nil{t.Fatal(err)}
	if updated.VehicleID!=""||updated.Pressure!=nil||updated.RunFlat||updated.DOT!=""||updated.Status!="stored"{t.Fatalf("fields were not cleared: %#v",updated)}
	call(http.MethodDelete,path,"a",nil,204)
	call(http.MethodDelete,path,"a",nil,404)
	// Batch creation: unique IDs, per-tire values and distinct positions.
	item.VehicleID="v-a";item.Status="mounted";item.Position="";item.Pressure=&pressure
	four:=4
	batch:=tireSaveRequest{TireDetails:item,Quantity:&four,Positions:[]string{"front-left","front-right","rear-left","rear-right"}}
	call(http.MethodPost,"/api/tires","b",batch,403)
	w=call(http.MethodPost,"/api/tires","a",batch,201)
	var response struct{Count int `json:"count"`; Tires []db.Tire `json:"tires"`}
	if err:=json.Unmarshal(w.Body.Bytes(),&response);err!=nil||response.Count!=4||len(response.Tires)!=4{t.Fatalf("invalid batch: %s",w.Body.String())}
	ids:=map[string]bool{}; positions:=map[string]bool{}
	for _,tire:=range response.Tires{
		if tire.ID==""||ids[tire.ID]||positions[tire.Position]||tire.Pressure==nil||*tire.Pressure!=pressure{t.Fatalf("invalid tire in batch: %#v",tire)}
		ids[tire.ID]=true;positions[tire.Position]=true
	}
	call(http.MethodPost,"/api/tires","a",batch,409)
	for _,tire:=range response.Tires{if tire.Position!="rear-right"{call(http.MethodDelete,"/api/tires/"+tire.ID,"a",nil,204)}}
	// The last position is occupied: preceding inserts must be rolled back.
	call(http.MethodPost,"/api/tires","a",batch,409)
	var count int64
	if err:=g.Model(&db.Tire{}).Where("owner_user_id = ?","a").Count(&count).Error;err!=nil||count!=1{t.Fatalf("batch rollback failed, count=%d, err=%v",count,err)}
	two:=2
	batch.Quantity=&two;batch.Positions=[]string{"front-left","front-left"}
	call(http.MethodPost,"/api/tires","a",batch,422)
	batch.Positions=[]string{"front-left"}
	call(http.MethodPost,"/api/tires","a",batch,422)
	batch.Positions=nil;batch.Status="stored"
	w=call(http.MethodPost,"/api/tires","a",batch,201)
	if err:=json.Unmarshal(w.Body.Bytes(),&response);err!=nil||response.Count!=2||len(response.Tires)!=2{t.Fatalf("invalid stored batch: %s",w.Body.String())}
	for _,tire:=range response.Tires{if tire.Position!=""{t.Fatal("stored tire has position")}}
	zero:=0;batch.Quantity=&zero
	call(http.MethodPost,"/api/tires","a",batch,422)
	five:=5;batch.Quantity=&five
	call(http.MethodPost,"/api/tires","a",batch,422)

}
