package controllers

import (
	"math"
	"testing"
)

func TestParseReceiptGermanSample(t *testing.T) {
	text := "Shell Tankstelle\n17.11.2026 18:22\nSuper E10\n42,00 L x 1,830 EUR/L\nGesamtbetrag 76,86 EUR"
	fields := parseReceipt(text)
	if fields["date"] != "2026-11-17" { t.Errorf("date: %v", fields["date"]) }
	if fields["fillingStation"] != "Shell Tankstelle" { t.Errorf("station: %v", fields["fillingStation"]) }
	for key, want := range map[string]float64{"fuelQuantity":42, "perUnitPrice":1.83, "totalAmount":76.86} {
		value, ok := fields[key].(float64)
		if !ok || math.Abs(value-want) > 0.0001 { t.Errorf("%s: got %v, want %f", key, fields[key], want) }
	}
}

func TestParseReceiptWithoutValues(t *testing.T) {
	fields := parseReceipt("Danke fuer Ihren Einkauf")
	if fields["fuelQuantity"] != nil || fields["totalAmount"] != nil || fields["perUnitPrice"] != nil {
		t.Fatalf("unexpected guessed values: %v", fields)
	}
}
