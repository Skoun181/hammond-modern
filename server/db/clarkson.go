package db

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"gorm.io/driver/mysql"

	"gorm.io/gorm"
)

func MigrateClarkson(connectionString, adminEmail, adminPassword string) error {
	if connectionString == "" || adminEmail == "" || len(adminPassword) < 12 || len(adminPassword) > 72 {
		return errors.New("CLARKSON_DSN, CLARKSON_ADMIN_EMAIL und ein neues Administratorpasswort (12–72 Bytes) sind erforderlich")
	}
	canInitialize, err := CanInitializeSystem()
	if err != nil {
		return err
	}
	if !canInitialize {
		return errors.New("Clarkson kann nur in eine leere Instanz importiert werden")
	}
	source, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		return errors.New("Verbindung zur Clarkson-Datenbank fehlgeschlagen")
	}
	connection, err := source.DB()
	if err != nil {
		return err
	}
	defer connection.Close()
	/////Models
	type CUser struct {
		ID                  string `gorm:"column:id"`
		Email               string `gorm:"column:email"`
		Username            string `gorm:"column:username"`
		Admin               bool   `gorm:"column:admin"`
		FuelUnit            int    `gorm:"column:fuelUnit"`
		DistanceUnit        int    `gorm:"column:distanceUnit"`
		FuelConsumptionUnit int    `gorm:"column:fuelConsumptionUnit"`
		CurrencyUnit        int    `gorm:"column:currencyUnit"`
	}

	type CVehicle struct {
		ID                string `gorm:"column:id"`
		User              string
		Name              string
		Registration      string
		Make              string
		Model             string
		YearOfManufacture int `gorm:"column:yearOfManufacture"`
		Vin               string
		EngineSizeCC      int `gorm:"column:engineSizeCC"`
		FuelType          int `gorm:"column:fuelType"`
	}

	type CFuel struct {
		ID              string    `gorm:"column:id"`
		Vehicle         string    `gorm:"column:vehicle"`
		Date            time.Time `gorm:"column:date"`
		FuelAmount      float32   `gorm:"column:fuelAmount"`
		TotalCost       float32   `gorm:"column:totalCost"`
		FuelUnitCost    float32   `gorm:"column:fuelUnitCost"`
		OdometerReading int       `gorm:"column:odometerReading"`
		Notes           string    `gorm:"column:notes"`
		FullTank        bool      `gorm:"column:fullTank"`
		MissedFillup    bool      `gorm:"column:missedFillUp"`
	}

	distanceUnitMap := map[int]DistanceUnit{
		1: MILES,
		2: KILOMETERS,
	}

	fuelTypeMap := map[int]FuelType{
		1: PETROL,
		2: DIESEL,
		3: ETHANOL,
		4: LPG,
	}

	fuelUnitsMap := map[int]FuelUnit{
		1: LITRE,
		2: GALLON,
		3: US_GALLON,
	}
	currencyMap := map[int]string{
		1: "GBP",
		2: "USD",
		3: "EUR",
		4: "AUD",
		5: "CAD",
	}

	var allUsers []CUser
	var allVehicles []CVehicle
	var allFillups []CFuel
	if err := source.Table("Users").Find(&allUsers).Error; err != nil {
		return errors.New("Clarkson-Benutzer konnten nicht gelesen werden")
	}
	if err := source.Table("Vehicles").Find(&allVehicles).Error; err != nil {
		return errors.New("Clarkson-Fahrzeuge konnten nicht gelesen werden")
	}
	if err := source.Table("Fuel").Find(&allFillups).Error; err != nil {
		return errors.New("Clarkson-Tankvorgänge konnten nicht gelesen werden")
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Unscoped().Model(&User{}).Count(&count).Error; err != nil {
			return err
		}
		if count != 0 {
			return errors.New("Clarkson kann nur in eine leere Instanz importiert werden")
		}
		newUsers := make(map[string]User)
		oldUsers := make(map[string]CUser)
		hasAdmin := false
		for _, original := range allUsers {
			user := User{
				Email: strings.ToLower(original.Email), Currency: currencyMap[original.CurrencyUnit],
				DistanceUnit: distanceUnitMap[original.DistanceUnit], Role: USER, Name: original.Username,
				DateFormat: "MM/dd/yyyy",
			}
			if original.Admin {
				user.Role = ADMIN
			}
			var randomPassword [24]byte
			if _, err := rand.Read(randomPassword[:]); err != nil {
				return err
			}
			password := hex.EncodeToString(randomPassword[:])
			if original.Admin && strings.EqualFold(adminEmail, user.Email) {
				hasAdmin = true
				password = adminPassword
				var setting Setting
				if err := tx.First(&setting).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				setting.Currency, setting.DistanceUnit = user.Currency, user.DistanceUnit
				if err := tx.Save(&setting).Error; err != nil {
					return err
				}
			}
			if err := user.SetPassword(password); err != nil {
				return err
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			newUsers[original.ID], oldUsers[original.ID] = user, original
		}
		if !hasAdmin {
			return errors.New("Die angegebene Administrator-E-Mail wurde in Clarkson nicht gefunden")
		}
		vehicles := make(map[string]Vehicle)
		owners := make(map[string]User)
		for _, original := range allVehicles {
			owner, ok := newUsers[original.User]
			if !ok {
				return errors.New("Clarkson-Fahrzeug ohne gültigen Besitzer")
			}
			vehicle := Vehicle{
				Nickname: original.Name, Registration: original.Registration, Model: original.Model,
				Make: original.Make, YearOfManufacture: original.YearOfManufacture, EngineSize: float32(original.EngineSizeCC),
				FuelUnit: fuelUnitsMap[oldUsers[original.User].FuelUnit], FuelType: fuelTypeMap[original.FuelType],
			}
			if err := tx.Create(&vehicle).Error; err != nil {
				return err
			}
			if err := tx.Create(&UserVehicle{UserID: owner.ID, VehicleID: vehicle.ID, IsOwner: true}).Error; err != nil {
				return err
			}
			vehicles[original.ID], owners[original.ID] = vehicle, owner
		}
		for _, original := range allFillups {
			vehicle, ok := vehicles[original.Vehicle]
			if !ok {
				return errors.New("Clarkson-Tankvorgang ohne gültiges Fahrzeug")
			}
			owner := owners[original.Vehicle]
			fillup := Fillup{VehicleID: vehicle.ID, FuelUnit: vehicle.FuelUnit,
				FuelQuantity: original.FuelAmount, PerUnitPrice: original.FuelUnitCost,
				TotalAmount: original.TotalCost, OdoReading: original.OdometerReading,
				IsTankFull: &original.FullTank, HasMissedFillup: &original.MissedFillup, Comments: original.Notes,
				UserID: owner.ID, Date: original.Date, Currency: owner.Currency, DistanceUnit: owner.DistanceUnit}
			if err := tx.Create(&fillup).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
