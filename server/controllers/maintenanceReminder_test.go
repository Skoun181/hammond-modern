package controllers
import("net/http";"net/http/httptest";"testing";"io";"strings")
func TestMaintenancePush(t *testing.T){
 if _,err:=reminderEndpoint("https://user:pass@example.com/topic");err==nil{t.Fatal("embedded credentials accepted")}
 if _,err:=reminderEndpoint("file:///tmp/topic");err==nil{t.Fatal("unsupported scheme accepted")}
 if reminderStage("soon")!=1||reminderStage("due")!=2||reminderStage("overdue")!=2||reminderStage("planned")!=0{t.Fatal("invalid stages")}
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Method!="POST"||r.Header.Get("Authorization")!="Bearer example"{t.Error("missing auth or method")}
  b,_:=io.ReadAll(r.Body);if !strings.Contains(string(b),"fällig"){t.Error("missing reminder")}
  w.WriteHeader(200)
 }))
 defer server.Close()
 endpoint,err:=reminderEndpoint(server.URL+"/topic");if err!=nil{t.Fatal(err)}
 if err=sendMaintenanceReminder(endpoint,"example",2);err!=nil{t.Fatal(err)}
}
