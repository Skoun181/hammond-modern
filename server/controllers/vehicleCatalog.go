package controllers

import (
    "encoding/json"
    "net/http"
    "net/url"
    "strings"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
)

type vehicleModelCacheEntry struct {
    models []string
    expires time.Time
}

var vehicleModelCache = struct {
    sync.Mutex
    entries map[string]vehicleModelCacheEntry
}{entries: make(map[string]vehicleModelCacheEntry)}

var vehicleCatalogClient = &http.Client{Timeout: 12 * time.Second}

// Model lookup is server-side so browsers do not depend on third-party CORS.
func RegisterVehicleCatalogController(router *gin.RouterGroup) {
    router.GET("/vehicle-models", func(c *gin.Context) {
        makeName := strings.TrimSpace(c.Query("make"))
        if len(makeName) < 2 || len(makeName) > 80 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Bitte eine gültige Automarke eingeben."})
            return
        }
        cacheKey := strings.ToLower(makeName)
        vehicleModelCache.Lock()
        if item, ok := vehicleModelCache.entries[cacheKey]; ok && time.Now().Before(item.expires) {
            vehicleModelCache.Unlock()
            c.JSON(http.StatusOK, gin.H{"models": item.models, "source": "cache"})
            return
        }
        vehicleModelCache.Unlock()

        endpoint := "https://vpic.nhtsa.dot.gov/api/vehicles/GetModelsForMake/" + url.PathEscape(makeName) + "?format=json"
        req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, endpoint, nil)
        if err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": "Modellabfrage konnte nicht vorbereitet werden."})
            return
        }
        req.Header.Set("Accept", "application/json")
        resp, err := vehicleCatalogClient.Do(req)
        if err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": "Online-Modellkatalog gerade nicht erreichbar. Modell kann manuell eingegeben werden."})
            return
        }
        defer resp.Body.Close()
        if resp.StatusCode != http.StatusOK {
            c.JSON(http.StatusBadGateway, gin.H{"error": "Der Modellkatalog antwortet derzeit nicht."})
            return
        }
        var payload struct {
            Results []struct {
                ModelName string `json:"Model_Name"`
            } `json:"Results"`
        }
        if err := json.NewDecoder(http.MaxBytesReader(c.Writer, resp.Body, 2<<20)).Decode(&payload); err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": "Der Modellkatalog hat ungültige Daten geliefert."})
            return
        }
        unique := map[string]bool{}
        models := make([]string, 0, len(payload.Results))
        for _, item := range payload.Results {
            model := strings.TrimSpace(item.ModelName)
            if model == "" || unique[strings.ToLower(model)] { continue }
            unique[strings.ToLower(model)] = true
            models = append(models, model)
        }
        // Cache both successful and empty results to avoid unnecessary requests.
        vehicleModelCache.Lock()
        vehicleModelCache.entries[cacheKey] = vehicleModelCacheEntry{models: models, expires: time.Now().Add(24 * time.Hour)}
        vehicleModelCache.Unlock()
        c.JSON(http.StatusOK, gin.H{"models": models, "source": "nhtsa"})
    })
}
