package service

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/akhilrex/hammond/db"
	"github.com/akhilrex/hammond/internal/sanitize"
	"github.com/akhilrex/hammond/models"
	uuid "github.com/satori/go.uuid"
)

func CreateAttachment(path, originalName string, size int64, contentType, userId string) (*db.Attachment, error) {
	model := &db.Attachment{
		Path:         path,
		OriginalName: originalName,
		Size:         size,
		ContentType:  contentType,
		UserID:       userId,
	}
	tx := db.DB.Create(&model)

	if tx.Error != nil {
		return nil, tx.Error
	}
	return model, nil
}

func CreateQuickEntry(model models.CreateQuickEntryModel, attachmentId, userId string) (*db.QuickEntry, error) {
	toCreate := &db.QuickEntry{
		AttachmentID: attachmentId,
		UserID:       userId,
		Comments:     model.Comments,
	}
	tx := db.DB.Create(&toCreate)

	if tx.Error != nil {
		return nil, tx.Error
	}
	return toCreate, nil
}
func GetAllQuickEntries(sorting string) (*[]db.QuickEntry, error) {
	return db.GetAllQuickEntries(sorting)
}
func GetQuickEntriesForUser(userId, sorting string) (*[]db.QuickEntry, error) {
	return db.GetQuickEntriesForUser(userId, sorting)
}
func GetQuickEntryById(id string) (*db.QuickEntry, error) {
	return db.GetQuickEntryById(id)
}
func DeleteQuickEntryById(id string) error {
	return db.DeleteQuickEntryById(id)
}
func SetQuickEntryAsProcessed(id string) error {
	return db.SetQuickEntryAsProcessed(id, time.Now())

}

func GetAttachmentById(id string) (*db.Attachment, error) {
	return db.GetAttachmentById(id)
}

func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func GetFilePath(originalName string) string {
	dataPath := os.Getenv("DATA")
	return path.Join(dataPath, getFileName(originalName))
}
func getFileName(orig string) string {

	ext := filepath.Ext(orig)
	return uuid.NewV4().String() + ext

}

func CreateBackup() (string, error) {
	PortableDataLock.Lock()
	defer PortableDataLock.Unlock()
	archive, _, err := CreatePortableZIP()
	if err != nil {
		return "", err
	}
	defer os.Remove(archive)
	folder := filepath.Join(os.Getenv("CONFIG"), "backups")
	if err := os.MkdirAll(folder, 0700); err != nil {
		return "", err
	}
	name := "hammond-portable-" + time.Now().UTC().Format("20060102-150405.000000000") + ".zip"
	if err := os.Rename(archive, filepath.Join(folder, name)); err != nil {
		return "", err
	}
	files, err := filepath.Glob(filepath.Join(folder, "hammond-portable-*.zip"))
	if err != nil {
		return name, err
	}
	sort.Strings(files)
	if len(files) > 5 {
		for _, path := range files[:len(files)-5] {
			if err := os.Remove(path); err != nil {
				return name, err
			}
		}
	}
	return name, nil
}

func httpClient() *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			//	r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	return &client
}

func createFolder(folder string, parent string) string {
	folder = cleanFileName(folder)
	//str := stringy.New(folder)
	folderPath := path.Join(parent, folder)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		os.MkdirAll(folderPath, 0777)
		changeOwnership(folderPath)
	}
	return folderPath
}

func createDataFolderIfNotExists(folder string) string {
	dataPath := os.Getenv("DATA")
	return createFolder(folder, dataPath)
}
func createConfigFolderIfNotExists(folder string) string {
	dataPath := os.Getenv("CONFIG")
	return createFolder(folder, dataPath)
}

func cleanFileName(original string) string {
	return sanitize.Name(original)
}
func changeOwnership(path string) {
	uid, err1 := strconv.Atoi(os.Getenv("PUID"))
	gid, err2 := strconv.Atoi(os.Getenv("PGID"))
	fmt.Println(path)
	if err1 == nil && err2 == nil {
		fmt.Println(path + " : Attempting change")
		os.Chown(path, uid, gid)
	}

}
func DeleteFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(filePath); err != nil {
		return err
	}
	return nil
}
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
