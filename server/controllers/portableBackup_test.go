package controllers

import (
  "net/http"
  "net/http/httptest"
  "testing"

  "github.com/akhilrex/hammond/db"
  "github.com/gin-gonic/gin"
)

func TestShouldBeAdminBlocksMissingAndDisabledUsers(t *testing.T) {
  gin.SetMode(gin.TestMode)
  checks:=[]struct {
    label string
    user db.User
    want int
  }{
    {"missing",db.User{},http.StatusForbidden},
    {"deleted-admin",db.User{Role:db.ADMIN},http.StatusForbidden},
    {"normal-user",db.User{Base:db.Base{ID:"user"},Role:db.USER},http.StatusForbidden},
    {"disabled-admin",db.User{Base:db.Base{ID:"admin"},Role:db.ADMIN,IsDisabled:true},http.StatusForbidden},
    {"valid-admin",db.User{Base:db.Base{ID:"admin"},Role:db.ADMIN},http.StatusOK},
  }
  for _,tc:=range checks {
    t.Run(tc.label,func(t *testing.T){
      router:=gin.New()
      called:=false
      router.GET("/secret",func(c *gin.Context){c.Set("userModel",tc.user)},ShouldBeAdmin(),func(c *gin.Context){
        called=true
        c.String(http.StatusOK,"OK")
      })
      rr:=httptest.NewRecorder()
      router.ServeHTTP(rr,httptest.NewRequest(http.MethodGet,"/secret",nil))
      if rr.Code!=tc.want{t.Fatalf("status=%d, expected=%d",rr.Code,tc.want)}
      if called!=(tc.want==http.StatusOK){t.Fatal("unauthorised handler execution")}
    })
  }
}
