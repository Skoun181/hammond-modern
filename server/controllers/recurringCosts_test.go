package controllers

import (
 "testing"
 "time"
)
func TestHistoricalFinanceValidation(t *testing.T){
 now:=time.Date(2026,11,22,12,0,0,0,time.Local)
 in:=costInput{VehicleID:"v",Title:"Auto",Category:"Finanzierung",AmountCents:30000,StartDate:"2025-06-01",IntervalMonths:1,Installments:60}
 if validCost(&in,now){t.Fatal("historical plan must explicitly choose handling of previous payments")}
 for _,n:=range []int{0,12,18}{
  in.AlreadyRecorded=&n
  if !validCost(&in,now){t.Fatalf("valid historical count rejected: %d",n)}
 }
 for _,n:=range []int{-1,19,61}{
  in.AlreadyRecorded=&n
  if validCost(&in,now){t.Fatalf("invalid historical count accepted: %d",n)}
 }
 in.StartDate="2026-12-01";in.AlreadyRecorded=nil
 if !validCost(&in,now){t.Fatal("future plan rejected")}
}
