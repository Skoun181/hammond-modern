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
  good:=[]string{"photo.jpg","rechnung.pdf","file-123.png","report.xlsx"}
  for _,name:=range good {if !portableAssetName(name){t.Errorf("valid filename rejected: %s",name)}}
  bad:=[]string{"../x.pdf","folder/receipt.pdf","..\\x.pdf",".secret.txt","malware.html","file.exe",""}
  for _,name:=range bad{if portableAssetName(name){t.Errorf("unsafe filename accepted: %s",name)}}
}

func TestPortableZIPRejectsTraversal(t *testing.T) {
  dir:=t.TempDir()
  archive:=filepath.Join(dir,"input.zip")
  file,err:=os.Create(archive);if err!=nil{t.Fatal(err)}
  zw:=zip.NewWriter(file)
  w,err:=zw.Create("../outside.txt");if err!=nil{t.Fatal(err)}
  w.Write([]byte("bad"))
  if err:=zw.Close();err!=nil{t.Fatal(err)}
  file.Close()
  _,_,err=portableExtract(archive,filepath.Join(dir,"stage"))
  if err==nil{t.Fatal("ZIP with path traversal was accepted")}
}

func TestPortableBackupRoundTripAndPasswordRedaction(t *testing.T) {
  root:=t.TempDir()
  config:=filepath.Join(root,"config")
  assets:=filepath.Join(root,"assets")
  if err:=os.MkdirAll(config,0700);err!=nil{t.Fatal(err)}
  if err:=os.MkdirAll(assets,0700);err!=nil{t.Fatal(err)}
  t.Setenv("CONFIG",config)
  t.Setenv("DATA",assets)

  original:=db.DB
  t.Cleanup(func(){db.DB=original})
  g,err:=gorm.Open(sqlite.Open(filepath.Join(config,"hammond.db")),&gorm.Config{DisableForeignKeyConstraintWhenMigrating:true})
  if err!=nil{t.Fatal(err)}
  db.DB=g
  t.Cleanup(func(){raw,_:=g.DB();if raw!=nil{raw.Close()}})
  if err=g.AutoMigrate(&db.User{},&db.Vehicle{},&db.Attachment{});err!=nil{t.Fatal(err)}
  user:=db.User{Email:"admin@example.test",Role:db.ADMIN,Name:"Altes Profil"}
  if err=user.SetPassword("secure-example-password");err!=nil{t.Fatal(err)}
  if err=g.Create(&user).Error;err!=nil{t.Fatal(err)}
  attachmentName:="receipt.pdf"
  attachmentContent:=[]byte("%PDF-1.4 test document")
  if err=os.WriteFile(filepath.Join(assets,attachmentName),attachmentContent,0600);err!=nil{t.Fatal(err)}
  attachment:=db.Attachment{Path:filepath.Join(assets,attachmentName),OriginalName:"receipt.pdf",Size:int64(len(attachmentContent)),ContentType:"application/pdf",UserID:user.ID}
  if err=g.Create(&attachment).Error;err!=nil{t.Fatal(err)}

  zipFile,summary,err:=CreatePortableZIP()
  if err!=nil{t.Fatalf("export failed: %v",err)}
  defer os.Remove(zipFile)
  if summary.AssetCount!=1||summary.Manifest.Counts["users"]!=1{t.Fatalf("bad summary: %#v",summary)}

  reader,err:=zip.OpenReader(zipFile);if err!=nil{t.Fatal(err)}
  restoredSnapshot:=filepath.Join(root,"sanitised.sqlite")
  found:=false
  for _,entry:=range reader.File {
    if entry.Name!="data.sqlite"{continue}
    found=true
    in,err:=entry.Open();if err!=nil{t.Fatal(err)}
    out,err:=os.Create(restoredSnapshot);if err!=nil{t.Fatal(err)}
    if _,err=io.Copy(out,in);err!=nil{t.Fatal(err)}
    out.Close();in.Close()
  }
  reader.Close()
  if !found{t.Fatal("SQLite snapshot missing")}
  snap,err:=sql.Open("sqlite3",restoredSnapshot);if err!=nil{t.Fatal(err)}
  var stored string
  if err=snap.QueryRow("SELECT password FROM users").Scan(&stored);err!=nil{t.Fatal(err)}
  snap.Close()
  if stored!=""{t.Fatal("password hash leaked into export")}

  inspected,err:=InspectPortableZIP(zipFile,user)
  if err!=nil{t.Fatalf("preview failed: %v",err)}
  if !inspected.AdminEmailMatched{t.Fatal("admin not found in backup")}

  if err=g.Model(&db.User{}).Where("id=?",user.ID).Update("name","Neu").Error;err!=nil{t.Fatal(err)}
  if err=os.Remove(filepath.Join(assets,attachmentName));err!=nil{t.Fatal(err)}
  // Restores require the existing attachment to be present for the local
  // safety snapshot. Replace its contents instead of deleting it.
  if err=os.WriteFile(filepath.Join(assets,attachmentName),[]byte("changed"),0600);err!=nil{t.Fatal(err)}
  result,err:=RestorePortableZIP(zipFile,user)
  if err!=nil{t.Fatalf("restore failed: %v",err)}
  if result.SafetyBackup==""{t.Fatal("missing safety snapshot")}
  var after db.User
  if err=g.Where("email=?",user.Email).First(&after).Error;err!=nil{t.Fatal(err)}
  if after.Name!="Altes Profil" || after.Password!=user.Password {
    t.Fatalf("restored user mismatch: name=%q password matched=%t",after.Name,after.Password==user.Password)
  }
  fileContent,err:=os.ReadFile(filepath.Join(assets,attachmentName))
  if err!=nil{t.Fatal(err)}
  if string(fileContent)!=string(attachmentContent){t.Fatal("asset contents were not restored")}
  var backupPass string
  rescue,err:=sql.Open("sqlite3",filepath.Join(result.SafetyBackup,"hammond.db"));if err!=nil{t.Fatal(err)}
  if err=rescue.QueryRow("SELECT password FROM users").Scan(&backupPass);err!=nil{t.Fatal(err)}
  rescue.Close()
  if backupPass==""{t.Fatal("local safety snapshot did not preserve existing credentials")}
  // Simulate a container interruption after the recovery journal was created:
  // the following startup must restore the pre-restore snapshot consistently.
  if err=portableStartJournal(config,result.SafetyBackup);err!=nil{t.Fatal(err)}
  if err=g.Model(&db.User{}).Where("id=?",user.ID).Update("name","Interrupted").Error;err!=nil{t.Fatal(err)}
  if err=os.WriteFile(filepath.Join(assets,attachmentName),[]byte("interrupted asset"),0600);err!=nil{t.Fatal(err)}
  live,_:=g.DB()
  if err=live.Close();err!=nil{t.Fatal(err)}
  if err=RecoverPortableRestore();err!=nil{t.Fatalf("startup recovery failed: %v",err)}
  recovered,err:=sql.Open("sqlite3",filepath.Join(config,"hammond.db"))
  if err!=nil{t.Fatal(err)}
  var name string
  if err=recovered.QueryRow("SELECT name FROM users WHERE email=?",user.Email).Scan(&name);err!=nil{t.Fatal(err)}
  recovered.Close()
  if name!="Neu"{t.Fatalf("recovery did not restore the safety snapshot: %s",name)}
  recoveredFile,err:=os.ReadFile(filepath.Join(assets,attachmentName))
  if err!=nil{t.Fatal(err)}
  if string(recoveredFile)!="changed"{t.Fatal("recovery did not restore previous attachment")}
  if _,err=os.Stat(portableJournalPath(config));!os.IsNotExist(err){t.Fatal("recovery journal still exists")}

}
