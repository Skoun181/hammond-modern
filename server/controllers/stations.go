package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var postalPattern = regexp.MustCompile(`^[0-9]{5}$`)
var stationCache = struct {
	sync.Mutex
	entries     map[string]cachedStations
	lastRequest time.Time
}{entries: map[string]cachedStations{}}

type cachedStations struct {
	data   interface{}
	expiry time.Time
}

var stationClient = &http.Client{Timeout: 8 * time.Second}

// Searches by German postcode. No request-per-keystroke; API key stays server-side.
func RegisterStationController(router *gin.RouterGroup) {
	router.GET("/stations", func(c *gin.Context) {
		code := c.Query("plz")
		if !postalPattern.MatchString(code) {
			c.JSON(400, gin.H{"error": "Bitte fünfstellige Postleitzahl eingeben"})
			return
		}
		key := os.Getenv("TANKERKOENIG_API_KEY")
		if key == "" {
			c.JSON(503, gin.H{"error": "Tankstellensuche nicht eingerichtet: TANKERKOENIG_API_KEY fehlt"})
			return
		}
		stationCache.Lock()
		if found, ok := stationCache.entries[code]; ok && time.Now().Before(found.expiry) {
			stationCache.Unlock()
			c.JSON(200, found.data)
			return
		}
		// One external search per minute; fail clearly rather than violate the public API's rate limit.
		if time.Since(stationCache.lastRequest) < time.Minute {
			stationCache.Unlock()
			c.JSON(429, gin.H{"error": "Bitte eine Minute warten oder dieselbe PLZ erneut abfragen"})
			return
		}
		stationCache.lastRequest = time.Now()
		stationCache.Unlock()
		geoURL := "https://nominatim.openstreetmap.org/search?countrycodes=de&postalcode=" + url.QueryEscape(code) + "&format=json&limit=1"
		req, err := http.NewRequest("GET", geoURL, nil)
		if err != nil {
			c.JSON(502, gin.H{"error": "Geocodierung fehlgeschlagen"})
			return
		}
		req.Header.Set("User-Agent", "HammondModern/1.0 (self-hosted vehicle administration)")
		resp, err := stationClient.Do(req)
		if err != nil {
			c.JSON(502, gin.H{"error": "Postleitzahl-Suche nicht erreichbar"})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			c.JSON(502, gin.H{"error": "Postleitzahl-Suche fehlgeschlagen"})
			return
		}
		var geo []struct {
			Lat string `json:"lat"`
			Lon string `json:"lon"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(c.Writer, resp.Body, 256*1024)).Decode(&geo); err != nil || len(geo) == 0 {
			c.JSON(404, gin.H{"error": "Postleitzahl nicht gefunden"})
			return
		}
		lat, err1 := strconv.ParseFloat(geo[0].Lat, 64)
		lon, err2 := strconv.ParseFloat(geo[0].Lon, 64)
		if err1 != nil || err2 != nil {
			c.JSON(502, gin.H{"error": "Ungültige Koordinaten"})
			return
		}
		u := fmt.Sprintf("https://creativecommons.tankerkoenig.de/json/list.php?lat=%f&lng=%f&rad=5&sort=dist&type=all&apikey=%s", lat, lon, url.QueryEscape(key))
		r, err := stationClient.Get(u)
		if err != nil {
			c.JSON(502, gin.H{"error": "Tankstellendienst nicht erreichbar"})
			return
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			c.JSON(502, gin.H{"error": "Tankstellendienst meldet Fehler"})
			return
		}
		var answer struct {
			OK       bool `json:"ok"`
			Stations []struct {
				ID     string  `json:"id"`
				Name   string  `json:"name"`
				Brand  string  `json:"brand"`
				Street string  `json:"street"`
				Place  string  `json:"place"`
				Dist   float64 `json:"dist"`
			} `json:"stations"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(c.Writer, r.Body, 1024*1024)).Decode(&answer); err != nil || !answer.OK {
			c.JSON(502, gin.H{"error": "Tankstellendaten nicht verfügbar"})
			return
		}
		stationCache.Lock()
		stationCache.entries[code] = cachedStations{data: answer.Stations, expiry: time.Now().Add(15 * time.Minute)}
		stationCache.Unlock()
		c.JSON(200, answer.Stations)
	})
}
