package controllers

import (
 "bytes"
 "context"
 "io"
 "math"
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
var receiptQuantity = regexp.MustCompile(`(?i)(\d{1,3}(?:[.,]\d{1,3})?)\s*(?:liter|litre|litres|l)\b`)
var receiptUnitPrice = regexp.MustCompile(`(?i)(\d[.,]\d{2,3})\s*(?:EUR|€)?\s*/\s*(?:liter|litre|l)\b`)
var receiptPriceLabel = regexp.MustCompile(`(?i)(einzelpreis|literpreis|preis\s*/\s*l)`)
var receiptTotal = regexp.MustCompile(`(?i)(gesamt(?:betrag)?|summe|total|zu zahlen|betrag|ec-karte)\b`)
var receiptStation = regexp.MustCompile(`(?i)\b(hem|aral|shell|esso|jet|totalenergies|avia|star|bft|orlen|tankstelle)\b`)

func receiptFloat(s string) float64 {
 v, err := strconv.ParseFloat(strings.ReplaceAll(s,",","."),64)
 if err != nil { return 0 }
 return v
}
func receiptLineValue(s string, low, high float64) float64 {
 matches := receiptNumber.FindAllString(s,-1)
 for _, m := range matches {
  v := receiptFloat(m)
  if v >= low && v <= high { return v }
 }
 return 0
}
func parseReceipt(text string) gin.H {
 lines := strings.Split(text,"\n")
 fields := gin.H{"date":"","fillingStation":"","fuelQuantity":nil,"perUnitPrice":nil,"totalAmount":nil,"amountEstimated":false}
 var liters, price float64
 for _, line := range lines {
  line = strings.TrimSpace(line)
  if fields["date"] == "" && (strings.Contains(strings.ToLower(line),"datum") || strings.Contains(line, ".")) {
   if m := receiptDate.FindStringSubmatch(line); m != nil {
    d,_:=strconv.Atoi(m[1]); mo,_:=strconv.Atoi(m[2]); y,_:=strconv.Atoi(m[3])
    if y < 100 { y+=2000 }
    value:=time.Date(y,time.Month(mo),d,12,0,0,0,time.UTC)
    if value.Year()==y && int(value.Month())==mo && value.Day()==d && y>=2000 && y<=time.Now().Year()+1 {fields["date"]=value.Format("2006-01-02")}
   }
  }
  if fields["fillingStation"] == "" && receiptStation.MatchString(line) && len(line)<80 {
   fields["fillingStation"]=line
  }
  if liters==0 {
   if m:=receiptQuantity.FindStringSubmatch(line);m!=nil {
    v:=receiptFloat(m[1]); if v>0 && v<500 {liters=v}
   }
  }
  if price==0 {
   if m:=receiptUnitPrice.FindStringSubmatch(line);m!=nil {
    v:=receiptFloat(m[1]); if v>=0.5 && v<=5 {price=v}
   } else if receiptPriceLabel.MatchString(line) {
    price=receiptLineValue(line,0.5,5)
   }
  }
 }
 if liters>0 { fields["fuelQuantity"]=liters }
 if price>0 { fields["perUnitPrice"]=price }
 // The fuel price is distinct from the basket total (drinks, deposits etc.).
 // Calculate only from recognized fuel quantity and unit price. A person must review it.
 if liters>0 && price>0 {
  fields["totalAmount"]=math.Round(liters*price*100)/100
  fields["amountEstimated"]=true
 } else {
  // Only use an explicitly labelled total when no fuel cost can be calculated.
  for _, line:=range lines {
   if receiptTotal.MatchString(line) {
    matches:=receiptNumber.FindAllString(line,-1)
    for i:=len(matches)-1;i>=0;i-- {
     v:=receiptFloat(matches[i]); if v>=1 && v<=1000 {fields["totalAmount"]=v; break}
    }
    if fields["totalAmount"] != nil {break}
   }
  }
 }
 return fields
}

func RegisterReceiptController(router *gin.RouterGroup) {
 router.POST("/receipts/recognize",func(c *gin.Context){
  c.Request.Body=http.MaxBytesReader(c.Writer,c.Request.Body,6<<20)
  file,err:=c.FormFile("file")
  if err!=nil || file.Size<=0 || file.Size>5<<20 {c.JSON(400,gin.H{"error":"Bitte JPG oder PNG bis 5 MB auswählen"});return}
  f,err:=file.Open();if err!=nil {c.JSON(400,gin.H{"error":"Bild konnte nicht gelesen werden"});return}
  defer f.Close()
  head:=make([]byte,512)
  n,_:=io.ReadFull(f,head)
  mime:=http.DetectContentType(head[:n])
  if mime!="image/png" && mime!="image/jpeg" {c.JSON(415,gin.H{"error":"Nur PNG und JPEG unterstützt"});return}
  suffix:=".png";if mime=="image/jpeg" {suffix=".jpg"}
  input,err:=os.CreateTemp("","hammond-bon-*"+suffix)
  if err!=nil {c.Status(500);return}
  defer os.Remove(input.Name())
  if _,err=input.Write(head[:n]);err==nil {_,err=io.Copy(input,io.LimitReader(f,5<<20))}
  input.Close()
  if err!=nil {c.JSON(500,gin.H{"error":"Beleg konnte nicht verarbeitet werden"});return}
  processed,err:=os.CreateTemp("","hammond-bon-*.png")
  if err!=nil {c.Status(500);return}
  processed.Close();defer os.Remove(processed.Name())
  ctx,cancel:=context.WithTimeout(c.Request.Context(),30*time.Second);defer cancel()
  // Reduce shadows and enlarge receipt text before OCR. No outside service is contacted.
  preprocess:=exec.CommandContext(ctx,"convert",input.Name(),"-colorspace","Gray","-auto-level","-resize","200%","-sharpen","0x1",processed.Name())
  if err=preprocess.Run();err!=nil {c.JSON(422,gin.H{"error":"Bildaufbereitung fehlgeschlagen"});return}
  ocr:=exec.CommandContext(ctx,"tesseract",processed.Name(),"stdout","-l","deu+eng","--psm","6")
  var stderr bytes.Buffer
  ocr.Stderr=&stderr
  output,err:=ocr.Output()
  if err!=nil {c.JSON(422,gin.H{"error":"Texterkennung fehlgeschlagen. Bitte manuell eintragen."});return}
  c.JSON(200,gin.H{"suggestions":parseReceipt(string(output)),"notice":"Erkannte Daten und Kraftstoffbetrag anhand des Originalbons prüfen. Das Bild wird nach der Verarbeitung gelöscht."})
 })
}
