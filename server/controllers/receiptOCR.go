package controllers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var receiptDate = regexp.MustCompile(`\b(\d{1,2})[./-](\d{1,2})[./-](\d{2,4})\b`)
var receiptNumber = regexp.MustCompile(`\d{1,5}(?:[.,]\d{1,3})?`)
var receiptTotal = regexp.MustCompile(`(?i)(gesamt(?:betrag)?|summe|total|zu zahlen|betrag|ec-karte)\b`)
var receiptLiters = regexp.MustCompile(`(?i)(liter|litre|litres|menge|\bl\b)`)
var receiptPrice = regexp.MustCompile(`(?i)(preis\s*/\s*l|eur\s*/\s*l|€\s*/\s*l|einzelpreis|literpreis|\b€/l\b)`)
var receiptStation = regexp.MustCompile(`(?i)(aral|shell|esso|jet|totalenergies|avia|hem|star|oil!|tankstelle|bft|raiffeisen|orlen)`)

func receiptFloat(s string) float64 {
	s = strings.ReplaceAll(s, ",", ".")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil { return 0 }
	return v
}

func receiptValue(lines []string, marker *regexp.Regexp, min, max float64) float64 {
	for _, line := range lines {
		if !marker.MatchString(line) { continue }
		matches := receiptNumber.FindAllString(line, -1)
		for i := len(matches)-1; i >= 0; i-- {
			v := receiptFloat(matches[i])
			if v >= min && v <= max { return v }
		}
	}
	return 0
}

func parseReceipt(text string) gin.H {
	lines := strings.Split(text, "\n")
	fields := gin.H{"date": "", "fillingStation": "", "fuelQuantity": nil, "perUnitPrice": nil, "totalAmount": nil}
	for _, line := range lines {
		if fields["date"] == "" {
			if match := receiptDate.FindStringSubmatch(line); match != nil {
				day, _ := strconv.Atoi(match[1]); month, _ := strconv.Atoi(match[2]); year, _ := strconv.Atoi(match[3])
				if year < 100 { year += 2000 }
				date := time.Date(year,time.Month(month),day,12,0,0,0,time.UTC)
				if date.Year() == year && int(date.Month()) == month && date.Day() == day && year >= 2000 && year <= time.Now().Year()+1 {
					fields["date"] = date.Format("2006-01-02")
				}
			}
		}
		if fields["fillingStation"] == "" && receiptStation.MatchString(line) && len(strings.TrimSpace(line)) < 90 {
			fields["fillingStation"] = strings.TrimSpace(line)
		}
	}
	if v := receiptValue(lines, receiptLiters, 1, 500); v > 0 { fields["fuelQuantity"] = v }
	if v := receiptValue(lines, receiptPrice, 0.5, 5); v > 0 { fields["perUnitPrice"] = v }
	if v := receiptValue(lines, receiptTotal, 1, 1000); v > 0 { fields["totalAmount"] = v }
	// Deliberately do not infer missing amounts; the user must review all values.
	return fields
}

func RegisterReceiptController(router *gin.RouterGroup) {
	router.POST("/receipts/recognize", func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 6<<20)
		file, err := c.FormFile("file")
		if err != nil || file.Size <= 0 || file.Size > 5<<20 { c.JSON(400,gin.H{"error":"Bitte JPG oder PNG bis 5 MB auswählen"}); return }
		f, err := file.Open()
		if err != nil { c.JSON(400,gin.H{"error":"Bild konnte nicht gelesen werden"}); return }
		defer f.Close()
		head := make([]byte,512)
		n, _ := io.ReadFull(f, head)
		kind := http.DetectContentType(head[:n])
		if kind != "image/jpeg" && kind != "image/png" { c.JSON(415, gin.H{"error":"Nur JPEG und PNG unterstützt"}); return }
		temp, err := os.CreateTemp("", "hammond-receipt-*.img")
		if err != nil { c.Status(500); return }
		defer os.Remove(temp.Name())
		if _, err = temp.Write(head[:n]); err == nil { _, err = io.Copy(temp, io.LimitReader(f, 5<<20)) }
		temp.Close()
		if err != nil { c.JSON(500,gin.H{"error":"Bildverarbeitung fehlgeschlagen"}); return }
		ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "tesseract", temp.Name(), "stdout", "-l", "deu+eng")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		if err != nil { c.JSON(422,gin.H{"error":"Texterkennung fehlgeschlagen. Bitte Werte manuell eintragen."}); return }
		c.JSON(200, gin.H{"suggestions":parseReceipt(string(output)), "rawText":strings.TrimSpace(string(output)), "notice":"Automatisch erkannte Werte immer kontrollieren. Belegbild wird nicht gespeichert."})
	})
}
