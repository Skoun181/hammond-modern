package controllers

import (
 "math"
 "testing"
)

func near(t *testing.T, field string, actual interface{}, want float64) {
 t.Helper()
 value,ok:=actual.(float64)
 if !ok || math.Abs(value-want)>0.001 {t.Errorf("%s = %v, want %.3f",field,actual,want)}
}

func TestParseReceiptGermanSample(t *testing.T) {
 fields:=parseReceipt("Shell Tankstelle\n17.11.2026 18:22\nSuper E10\n42,00 L x 1,830 EUR/L\nGesamtbetrag 76,86 EUR")
 if fields["date"]!="2026-11-17" {t.Errorf("date: %v",fields["date"])}
 if fields["fillingStation"]!="Shell Tankstelle" {t.Errorf("station: %v",fields["fillingStation"])}
 near(t,"fuelQuantity",fields["fuelQuantity"],42)
 near(t,"perUnitPrice",fields["perUnitPrice"],1.83)
 near(t,"totalAmount",fields["totalAmount"],76.86)
}

func TestParseReceiptHEMMixedShopping(t *testing.T) {
 fields:=parseReceipt("HEM Tankstelle\n41,00 Liter SAULENNUMMER 6\nSuper 76,63 EUR\n1,869 EUR/Liter\nDPG Red Bull 3,00 EUR\nDatum 15.11.2026\nBetrag 78,24 EUR\nTOTAL 78,24 EUR")
 if fields["date"]!="2026-11-15" {t.Errorf("date: %v",fields["date"])}
 near(t,"fuelQuantity",fields["fuelQuantity"],41)
 near(t,"perUnitPrice",fields["perUnitPrice"],1.869)
 near(t,"totalAmount",fields["totalAmount"],76.63)
 if fields["amountEstimated"]!=true {t.Errorf("expected calculated fuel amount: %v",fields)}
}

func TestParseReceiptWithoutValues(t *testing.T) {
 fields:=parseReceipt("Danke fuer Ihren Einkauf")
 if fields["fuelQuantity"]!=nil || fields["totalAmount"]!=nil || fields["perUnitPrice"]!=nil {
  t.Fatalf("unexpected guessed values: %v",fields)
 }
}
