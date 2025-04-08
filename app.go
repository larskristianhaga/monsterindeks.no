package main

import (
	"crypto/tls"
	"database/sql"
	"embed"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

//go:embed templates/*
var resources embed.FS

var t = template.Must(template.ParseFS(resources, "templates/*"))

var domain = "https://www.monsterindeks.no"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("App live and listening on port:", port)

	http.HandleFunc("/", RootHandler)
	http.HandleFunc("/insert-latest-monster-price", InsertLatestMonsterPriceHandler)
	http.HandleFunc("/get-latest-monster-price", GetLatestMonsterPriceHandler)
	http.HandleFunc("/get-raw-data", GetRawDataHandler)
	http.HandleFunc("/health", HealthHandler)
	http.HandleFunc("/robots.txt", RobotsHandler)
	http.HandleFunc("/sitemap.xml", SitemapHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func RootHandler(w http.ResponseWriter, _ *http.Request) {
	db := OpenDatabase()

	rows, err := db.Query("SELECT id, gross_price, gross_price_normalised, created_at FROM monsters ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []MonsterRecord
	for rows.Next() {
		var rec MonsterRecord
		if err := rows.Scan(&rec.ID, &rec.GrossPrice, &rec.GrossPriceNormalised, &rec.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		records = append(records, rec)
	}
	_ = t.ExecuteTemplate(w, "index.html.tmpl", records)
}

type MonsterRecord struct {
	ID                   int       `json:"id"`
	GrossPrice           string    `json:"gross_price"`
	GrossPriceNormalised int       `json:"gross_price_normalised"`
	CreatedAt            time.Time `json:"created_at"`
}

func ConvertPriceToNormalisedInteger(price string) int {
	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		log.Fatal(err)
	}

	priceInt := int(priceFloat * 100)
	return priceInt
}

func InsertLatestMonsterPriceHandler(w http.ResponseWriter, _ *http.Request) {
	monsterPrice := getMonsterData().GrossPrice
	monsterPriceNormalised := ConvertPriceToNormalisedInteger(monsterPrice)
	db := OpenDatabase()

	log.Println("Inserting monster price into database")
	_, err := db.Exec("INSERT INTO monsters (gross_price, gross_price_normalised) VALUES (?, ?)", monsterPrice, monsterPriceNormalised)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Println("Monster price inserted into database")
}

func GetLatestMonsterPriceHandler(w http.ResponseWriter, _ *http.Request) {
	monsterPrice := getMonsterData().GrossPrice
	normalizedMonsterPrice := ConvertPriceToNormalisedInteger(monsterPrice)

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("Current Monster price: " + monsterPrice + "\n" + "Current Monster normalised price : " + strconv.Itoa(normalizedMonsterPrice)))
}

func GetRawDataHandler(w http.ResponseWriter, _ *http.Request) {
	db := OpenDatabase()

	rows, err := db.Query("SELECT id, gross_price, gross_price_normalised, created_at FROM monsters ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []MonsterRecord
	for rows.Next() {
		var rec MonsterRecord
		if err := rows.Scan(&rec.ID, &rec.GrossPrice, &rec.GrossPriceNormalised, &rec.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		records = append(records, rec)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(records)
}

func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("I'm healthy"))
}

func RobotsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	data := map[string]string{
		"URL": domain + "/sitemap.xml",
	}

	_ = t.ExecuteTemplate(w, "robots.txt.tmpl", data)
}

func SitemapHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/xml")

	data := map[string]string{
		"URL": domain,
	}

	_ = t.ExecuteTemplate(w, "sitemap.xml.tmpl", data)
}

func getMonsterData() Monster {
	odaMonsterEndpoint := "https://oda.com/tienda-web-api/v1/products/23300/"

	client := createInsecureHTTPClient()

	response, err := client.Get(odaMonsterEndpoint)
	if err != nil {
		log.Fatal(err.Error())
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err.Error())
	}

	var monster Monster
	err = json.Unmarshal(responseData, &monster)
	if err != nil {
		log.Fatal(err.Error())
	}

	return monster
}

type Monster struct {
	GrossPrice string `json:"gross_price"`
}

func OpenDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "/data/monsterdatabase.db")
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func createInsecureHTTPClient() *http.Client {
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: customTransport}
}
