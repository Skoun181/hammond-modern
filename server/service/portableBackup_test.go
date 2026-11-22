package service

import (
	"archive/zip"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/akhilrex/hammond/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPortableAssetNames(t *testing.T) {
	good := []string{"photo.jpg", "rechnung.pdf", "file-123.png", "report.xlsx"}
	for _, name := range good {
		if !portableAssetName(name) {
			t.Errorf("valid filename rejected: %s", name)
		}
	}
	bad := []string{"../x.pdf", "folder/receipt.pdf", "..\\x.pdf", ".secret.txt", "malware.html", "file.exe", ""}
	for _, name := range bad {
		if portableAssetName(name) {
			t.Errorf("unsafe filename accepted: %s", name)
		}
	}
}

func TestPortableZIPRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "input.zip")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(file)
	w, err := zw.Create("../outside.txt")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("bad"))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	file.Close()
	_, _, err = portableExtract(archive, filepath.Join(dir, "stage"))
	if err == nil {
		t.Fatal("ZIP with path traversal was accepted")
	}
}

func TestPortableBackupRoundTripAndPasswordRedaction(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "config")
	assets := filepath.Join(root, "assets")
	if err := os.MkdirAll(config, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(assets, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG", config)
	t.Setenv("DATA", assets)

	original := db.DB
	t.Cleanup(func() { db.DB = original })
	g, err := gorm.Open(sqlite.Open(filepath.Join(config, "hammond.db")), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatal(err)
	}
	db.DB = g
	t.Cleanup(func() {
		raw, _ := g.DB()
		if raw != nil {
			raw.Close()
		}
	})
	if err = g.AutoMigrate(&db.User{}, &db.Vehicle{}, &db.Attachment{}); err != nil {
		t.Fatal(err)
	}
	user := db.User{Email: "admin@example.test", Role: db.ADMIN, Name: "Altes Profil"}
	if err = user.SetPassword("secure-example-password"); err != nil {
		t.Fatal(err)
	}
	if err = g.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	attachmentName := "receipt.pdf"
	attachmentContent := []byte("%PDF-1.4 test document")
	if err = os.WriteFile(filepath.Join(assets, attachmentName), attachmentContent, 0600); err != nil {
		t.Fatal(err)
	}
	attachment := db.Attachment{Path: filepath.Join(assets, attachmentName), OriginalName: "receipt.pdf", Size: int64(len(attachmentContent)), ContentType: "application/pdf", UserID: user.ID}
	if err = g.Create(&attachment).Error; err != nil {
		t.Fatal(err)
	}

	// Export a legacy snapshot before adding the tire table.
	legacyZIP, _, err := CreatePortableZIP()
	if err != nil { t.Fatal(err) }
	defer os.Remove(legacyZIP)
	if err = g.AutoMigrate(&db.Tire{}); err != nil { t.Fatal(err) }
	depth:=6.5
	tire:=db.Tire{ID:"test-tire",OwnerUserID:user.ID,TireDetails:db.TireDetails{Manufacturer:"Test",Size:"225/45 R17",Season:"winter",DOT:"2325",Status:"stored",TreadDepth:&depth}}
	if err = g.Create(&tire).Error; err != nil { t.Fatal(err) }
	preMileageZIP,_,err:=CreatePortableZIP()
	if err!=nil{t.Fatal(err)}
	defer os.Remove(preMileageZIP)
	if err=g.AutoMigrate(&db.TireMileageState{});err!=nil{t.Fatal(err)}
	if _,err=InspectPortableZIP(preMileageZIP,user);err!=nil{t.Fatalf("v1.1.7 preview: %v",err)}
	if err=g.Model(&db.Tire{}).Where("id = ?",tire.ID).Updates(map[string]interface{}{"status":"mounted","position":"front-left","vehicle_id":"test-vehicle"}).Error;err!=nil{t.Fatal(err)}
	mount:=db.TireMileageState{TireID:tire.ID,MountOdo:1000,MountedOn:"2026-03-01"}
	if err=g.Create(&mount).Error;err!=nil{t.Fatal(err)}
	if legacy, previewErr := InspectPortableZIP(legacyZIP,user); previewErr != nil || legacy.Manifest.Counts["tires"] != 0 { t.Fatalf("legacy preview: %v",previewErr) }

	if err=g.AutoMigrate(&db.RecurringCost{});err!=nil{t.Fatal(err)}
	plan:=db.RecurringCost{OwnerUserID:user.ID,VehicleID:"test-vehicle",Title:"Finanzierung",Category:"Finanzierung",AmountCents:30000,Currency:"EUR",StartDate:"2026-01-01",IntervalMonths:1,Installments:48,Generated:3,BookedCents:90000}
	if err=g.Create(&plan).Error;err!=nil{t.Fatal(err)}
	oldFinanceZIP,_,err:=CreatePortableZIP()
	if err!=nil{t.Fatal(err)}
	defer os.Remove(oldFinanceZIP)
	if err=g.AutoMigrate(&db.FinancingTerms{});err!=nil{t.Fatal(err)}
	principal:=int64(1200000);nominal:=450;effective:=480
	terms:=db.FinancingTerms{PlanID:plan.ID,DownPaymentCents:200000,DownPaymentDate:"2025-12-15",AlreadyRecorded:true,PrincipalCents:&principal,NominalBps:&nominal,EffectiveBps:&effective}
	if err=g.Create(&terms).Error;err!=nil{t.Fatal(err)}
	if _,err=InspectPortableZIP(oldFinanceZIP,user);err!=nil{t.Fatalf("pre-finance preview: %v",err)}
	preCalcZIP,_,err:=CreatePortableZIP()
	if err!=nil{t.Fatal(err)}
	defer os.Remove(preCalcZIP)
	if err=g.AutoMigrate(&db.FinancingCalculation{});err!=nil{t.Fatal(err)}
	monthly:=3750;first:=int64(6500)
	calc:=db.FinancingCalculation{PlanID:plan.ID,MonthlyRatePPM:&monthly,FirstInterestCents:&first}
	if err=g.Create(&calc).Error;err!=nil{t.Fatal(err)}
	if _,err=InspectPortableZIP(preCalcZIP,user);err!=nil{t.Fatalf("pre-calculation preview: %v",err)}
	if err=g.AutoMigrate(&db.HomeAssistantKey{});err!=nil{t.Fatal(err)}
	if err=g.Create(&db.HomeAssistantKey{OwnerUserID:user.ID,TokenHash:"private-test-key-hash"}).Error;err!=nil{t.Fatal(err)}
	zipFile, summary, err := CreatePortableZIP()
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	defer os.Remove(zipFile)
	if summary.Manifest.Counts["home_assistant_keys"] != 0 {t.Fatal("Home Assistant API key leaked into portable backup")}
	if summary.AssetCount != 1 || summary.Manifest.Counts["users"] != 1 || summary.Manifest.Counts["tires"] != 1 || summary.Manifest.Counts["tire_mileage_states"] != 1 {
		t.Fatalf("bad summary: %#v", summary)
	}

	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		t.Fatal(err)
	}
	restoredSnapshot := filepath.Join(root, "sanitised.sqlite")
	found := false
	for _, entry := range reader.File {
		if entry.Name != "data.sqlite" {
			continue
		}
		found = true
		in, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		out, err := os.Create(restoredSnapshot)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = io.Copy(out, in); err != nil {
			t.Fatal(err)
		}
		out.Close()
		in.Close()
	}
	reader.Close()
	if !found {
		t.Fatal("SQLite snapshot missing")
	}
	snap, err := sql.Open("sqlite3", restoredSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	var stored string
	if err = snap.QueryRow("SELECT password FROM users").Scan(&stored); err != nil {
		t.Fatal(err)
	}
	snap.Close()
	if stored != "" {
		t.Fatal("password hash leaked into export")
	}

	inspected, err := InspectPortableZIP(zipFile, user)
	if err != nil {
		t.Fatalf("preview failed: %v", err)
	}
	if !inspected.AdminEmailMatched {
		t.Fatal("admin not found in backup")
	}

	if _,err=RestorePortableZIP(preCalcZIP,user);err!=nil{t.Fatalf("pre-calculation restore: %v",err)}
	var calculationCount int64
	if err=g.Model(&db.FinancingCalculation{}).Count(&calculationCount).Error;err!=nil||calculationCount!=0{t.Fatal("older backup must clear calculation settings")}
	if _,err=RestorePortableZIP(oldFinanceZIP,user);err!=nil{t.Fatalf("pre-finance restore: %v",err)}
	var financeCount int64
	if err=g.Model(&db.FinancingTerms{}).Count(&financeCount).Error;err!=nil||financeCount!=0{t.Fatal("older backup must clear financing details")}
	// Verify v1.1.7 archive clears mounting states and retains tire records.
	if _,err=RestorePortableZIP(preMileageZIP,user);err!=nil{t.Fatalf("v1.1.7 restore: %v",err)}
	var mountCount int64
	if err=g.Model(&db.TireMileageState{}).Count(&mountCount).Error;err!=nil||mountCount!=0{t.Fatal("legacy archive did not clear mounting states")}
	// Verify old archive clears current tires, then newer archive restores the tire.
	if _, err = RestorePortableZIP(legacyZIP,user);err != nil {t.Fatalf("legacy restore failed: %v",err)}
	var tireCount int64
	if err=g.Model(&db.Tire{}).Count(&tireCount).Error;err!=nil||tireCount!=0{t.Fatal("legacy archive did not clear tires")}
	if err = g.Model(&db.User{}).Where("id=?", user.ID).Update("name", "Neu").Error; err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(filepath.Join(assets, attachmentName)); err != nil {
		t.Fatal(err)
	}
	// Restores require the existing attachment to be present for the local
	// safety snapshot. Replace its contents instead of deleting it.
	if err = os.WriteFile(filepath.Join(assets, attachmentName), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := RestorePortableZIP(zipFile, user)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	var restoredTire db.Tire
	if err = g.Where("id=?",tire.ID).First(&restoredTire).Error;err!=nil || restoredTire.DOT!="2325" || restoredTire.OwnerUserID!=user.ID || restoredTire.TreadDepth==nil || *restoredTire.TreadDepth!=6.5 { t.Fatalf("tire not restored: %#v %v",restoredTire,err) }
	var restoredMount db.TireMileageState
	if err=g.Where("tire_id = ?",tire.ID).First(&restoredMount).Error;err!=nil||restoredMount.MountOdo!=1000||restoredMount.MountedOn!="2026-03-01"{t.Fatalf("mounting interval not restored: %#v %v",restoredMount,err)}
	var restoredTerms db.FinancingTerms
	if err=g.Where("plan_id = ?",plan.ID).First(&restoredTerms).Error;err!=nil||restoredTerms.DownPaymentCents!=200000||restoredTerms.NominalBps==nil||*restoredTerms.NominalBps!=450||!restoredTerms.AlreadyRecorded{t.Fatalf("financing terms not restored: %#v %v",restoredTerms,err)}
	var restoredCalc db.FinancingCalculation
	if err=g.Where("plan_id = ?",plan.ID).First(&restoredCalc).Error;err!=nil||restoredCalc.MonthlyRatePPM==nil||*restoredCalc.MonthlyRatePPM!=3750||restoredCalc.FirstInterestCents==nil||*restoredCalc.FirstInterestCents!=6500{t.Fatalf("calculation not restored: %#v %v",restoredCalc,err)}
	var restoredKeyCount int64
	if err=g.Model(&db.HomeAssistantKey{}).Count(&restoredKeyCount).Error;err!=nil||restoredKeyCount!=0{t.Fatal("Home Assistant token unexpectedly restored")}
	if result.SafetyBackup == "" {
		t.Fatal("missing safety snapshot")
	}
	var after db.User
	if err = g.Where("email=?", user.Email).First(&after).Error; err != nil {
		t.Fatal(err)
	}
	if after.Name != "Altes Profil" || after.Password != user.Password {
		t.Fatalf("restored user mismatch: name=%q password matched=%t", after.Name, after.Password == user.Password)
	}
	fileContent, err := os.ReadFile(filepath.Join(assets, attachmentName))
	if err != nil {
		t.Fatal(err)
	}
	if string(fileContent) != string(attachmentContent) {
		t.Fatal("asset contents were not restored")
	}
	var backupPass string
	rescue, err := sql.Open("sqlite3", filepath.Join(result.SafetyBackup, "hammond.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err = rescue.QueryRow("SELECT password FROM users").Scan(&backupPass); err != nil {
		t.Fatal(err)
	}
	rescue.Close()
	if backupPass == "" {
		t.Fatal("local safety snapshot did not preserve existing credentials")
	}
	// Simulate a container interruption after the recovery journal was created:
	// the following startup must restore the pre-restore snapshot consistently.
	if err = portableStartJournal(config, result.SafetyBackup); err != nil {
		t.Fatal(err)
	}
	if err = g.Model(&db.User{}).Where("id=?", user.ID).Update("name", "Interrupted").Error; err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(assets, attachmentName), []byte("interrupted asset"), 0600); err != nil {
		t.Fatal(err)
	}
	live, _ := g.DB()
	if err = live.Close(); err != nil {
		t.Fatal(err)
	}
	if err = RecoverPortableRestore(); err != nil {
		t.Fatalf("startup recovery failed: %v", err)
	}
	recovered, err := sql.Open("sqlite3", filepath.Join(config, "hammond.db"))
	if err != nil {
		t.Fatal(err)
	}
	var name string
	if err = recovered.QueryRow("SELECT name FROM users WHERE email=?", user.Email).Scan(&name); err != nil {
		t.Fatal(err)
	}
	recovered.Close()
	if name != "Neu" {
		t.Fatalf("recovery did not restore the safety snapshot: %s", name)
	}
	recoveredFile, err := os.ReadFile(filepath.Join(assets, attachmentName))
	if err != nil {
		t.Fatal(err)
	}
	if string(recoveredFile) != "changed" {
		t.Fatal("recovery did not restore previous attachment")
	}
	if _, err = os.Stat(portableJournalPath(config)); !os.IsNotExist(err) {
		t.Fatal("recovery journal still exists")
	}

}
