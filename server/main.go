package main

import (
	"log"
	"net/http"
	"os"

	"github.com/akhilrex/hammond/controllers"
	"github.com/akhilrex/hammond/db"
	"github.com/akhilrex/hammond/service"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/jasonlvhit/gocron"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Recover any restore interrupted by container shutdown before opening SQLite.
	if err := service.RecoverPortableRestore(); err != nil {
		log.Fatalf("Restore-Recovery fehlgeschlagen: %v", err)
	}
	var err error
	db.DB, err = db.Init()
	if err != nil {
		log.Fatal(err)
	}
	db.Migrate()
	if len(os.Args) == 2 && os.Args[1] == "--import-clarkson" {
		if err := db.MigrateClarkson(os.Getenv("CLARKSON_DSN"), os.Getenv("CLARKSON_ADMIN_EMAIL"), os.Getenv("CLARKSON_ADMIN_PASSWORD")); err != nil {
			log.Fatal(err)
		}
		log.Println("Clarkson-Import abgeschlossen. Weitere Benutzer benötigen ein individuelles Passwort.")
		return
	}
	if err := service.OpenSessions(os.Getenv("CONFIG")); err != nil {
		log.Fatal(err)
	}
	r := gin.New()
	r.Use(gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		return params.Method + " " + params.Request.URL.Path + " " + http.StatusText(params.StatusCode) + "\n"
	}))

	// Serialize a full-instance restore against normal application requests.
	r.Use(controllers.PortableDataGuard())

	r.Use(setupSettings())
	r.Use(gin.CustomRecovery(func(c *gin.Context, _ interface{}) { c.AbortWithStatus(http.StatusInternalServerError) }))
	r.Use(location.Default())
	r.Use(static.Serve("/", static.LocalFile("./dist", true)))
	r.NoRoute(func(c *gin.Context) {
		//fmt.'Println(c.Request.URL.Path)
		c.File("dist/index.html")
	})
	router := r.Group("/api")

	controllers.RegisterAnonController(router)
	controllers.RegisterAnonMasterConroller(router)
	controllers.RegisterSetupController(router)
	// Dedicated read-only bearer authentication, independent of the normal JWT.
	controllers.RegisterHomeAssistantReadController(router)

	router.Use(controllers.AuthMiddleware(true))
	controllers.RegisterHomeAssistantKeyController(router)
	controllers.RegisterUserController(router)
	controllers.RegisterMastersController(router)
	controllers.RegisterAuthController(router)
	controllers.RegisterVehicleController(router)
	controllers.RegisterDriverController(router)
	controllers.RegisterTireController(router)
	controllers.RegisterStationController(router)
	controllers.RegisterVehicleCatalogController(router)
	controllers.RegisterInvoiceController(router)
	controllers.RegisterMaintenanceController(router)
	controllers.RegisterRecurringCostController(router)
	controllers.RegisterReceiptController(router)
	controllers.RegisterFilesController(router)
	controllers.RegisteImportController(router)
	controllers.RegisterPortableBackupController(router)
	controllers.RegisterReportsController(router)

	go assetEnv()
	go intiCron()

	r.Run(":3000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}
func setupSettings() gin.HandlerFunc {
	return func(c *gin.Context) {

		setting := db.GetOrCreateSetting()
		c.Set("setting", setting)
		c.Writer.Header().Set("X-Clacks-Overhead", "GNU Terry Pratchett")

		c.Next()
	}
}

func intiCron() {

	//gocron.Every(uint64(checkFrequency)).Minutes().Do(service.DownloadMissingEpisodes)
	service.RunRecurringCosts()
	gocron.Every(1).Minutes().Do(service.RunRecurringCosts)
	gocron.Every(5).Minutes().Do(controllers.RunMaintenanceReminders)
	gocron.Every(2).Days().Do(service.CreateBackup)
	<-gocron.Start()
}

func assetEnv() {
	log.Println("Config Dir: ", os.Getenv("CONFIG"))
	log.Println("Assets Dir: ", os.Getenv("DATA"))
}
