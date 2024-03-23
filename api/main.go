package main

import (
	"fmt"
	"log"
	"os"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttprouter"
    "database/sql"
    "github.com/lib/pq"
	"encoding/json"
	"reflect"
	"regexp"
)
var connection_data = "host="+os.Getenv("DATABASE_HOST")+" port="+os.Getenv("DATABASE_PORT")+" user="+os.Getenv("DATABASE_USERNAME")+" password="+os.Getenv("DATABASE_PASSWORD")+" dbname="+os.Getenv("DATABASE_NAME")+" sslmode=disable"
var basic_sql_query = "SELECT uuid,rendszam,tulajdonos,TO_CHAR(forgalmi_ervenyes,'YYYY-MM-DD'),adatok FROM jarmuvek"

type VehicleData struct {
	Uuid string `json:"uuid"`
    Rendszam string `json:"rendszam"`
    Tulajdonos string `json:"tulajdonos"`
	Adatok []string `json:"adatok"`
	ForgalmiErvenyes string `json:"forgalmi_ervenyes"`
}
func isStringEmptyOrNull(s string) bool {
    if s == "" {
        return true
    }
    return false
}
func JarmuvekPost(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params){
	body := ctx.PostBody()
	var vData VehicleData
	if err := json.Unmarshal(body, &vData); err != nil {
		ctx.Error("Error parsing JSON", fasthttp.StatusBadRequest)
		return
	}
	if isStringEmptyOrNull(vData.Rendszam) || isStringEmptyOrNull(vData.Tulajdonos) || isStringEmptyOrNull(vData.ForgalmiErvenyes) {
		ctx.Error("property_missing", fasthttp.StatusBadRequest)
		return
	}
	if reflect.TypeOf(vData.Adatok).Kind() != reflect.Slice{
		ctx.Error("adatok_is_not_an_array", fasthttp.StatusBadRequest)
		return
	}
	pattern := `^\d{4}-\d{2}-\d{2}$`
	regex := regexp.MustCompile(pattern)
	if !regex.MatchString(vData.ForgalmiErvenyes) {
		ctx.Error("forgalmi_ervenyes_format_error", fasthttp.StatusBadRequest)
		return
    }
	db, err := sql.Open("postgres", connection_data)
	if err != nil {
		ctx.Error("connection_error", fasthttp.StatusBadRequest)
		return
	}
	defer db.Close()
	var textArray = "{}"
	if len(vData.Adatok) != 0{
		textArray = "{" + "\"" + vData.Adatok[0] + "\""
		for _, v := range vData.Adatok[1:] {
			textArray += "," + "\"" + v + "\""
		}
		textArray += "}"
	}
	var uuid string
	sqlStatement := "INSERT INTO jarmuvek (rendszam, tulajdonos,forgalmi_ervenyes,adatok) VALUES ($1, $2, $3, $4) RETURNING uuid"
	err = db.QueryRow(sqlStatement, vData.Rendszam, vData.Tulajdonos,vData.ForgalmiErvenyes,textArray).Scan(&uuid)
	if err != nil {
		ctx.Error("Bad Request", fasthttp.StatusBadRequest)
	}
	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.Response.Header.Set("Location", "/jarmuvek/"+uuid)
    ctx.Response.Header.SetContentType("text/plain")
	ctx.Response.SetBodyString("success")
}
func JarmuvekKereses(ctx *fasthttp.RequestCtx, ps fasthttprouter.Params){
	var urlParam = string(ctx.FormValue("q"))
	if urlParam == "" {
		ctx.Response.Header.Set("Content-Type", "text/plain")
		ctx.Error("Empty query string", fasthttp.StatusBadRequest)
		return
	}
	allVData := []VehicleData{}
	db, err := sql.Open("postgres", connection_data)
	if err != nil {
		ctx.Response.Header.Set("Content-Type", "text/plain")
		ctx.Error("Cannot connect to pgsql", fasthttp.StatusBadRequest)
		return
	}
	defer db.Close()
	if len(urlParam) == 0{
		returnBody, marshalError := json.Marshal(allVData)
		if marshalError != nil {
			ctx.Error("Bad Request", fasthttp.StatusBadRequest)
			return
		}
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Write(returnBody)
		return;
	}
	rows, err := db.Query(basic_sql_query+" WHERE CONCAT(uuid,rendszam,tulajdonos) ILIKE '%"+urlParam+"%' OR adatok::text ILIKE '%"+urlParam+"%'")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    for rows.Next() {
		var adatokArray pq.StringArray
		var vData VehicleData
        if err := rows.Scan(&vData.Uuid, &vData.Rendszam, &vData.Tulajdonos,&vData.ForgalmiErvenyes,&adatokArray); err != nil {
            log.Fatal(err)
        }
		adatokStringArray := []string(adatokArray)
		vData.Adatok = make([]string, len(adatokStringArray))
		copy(vData.Adatok, adatokStringArray)
		allVData = append(allVData, vData)
    }
	returnBody, marshalError := json.Marshal(allVData)
	if marshalError != nil {
		ctx.Error("Bad Request", fasthttp.StatusBadRequest)
		return
	}
	ctx.Response.Header.Set("Content-Type", "application/json")
    ctx.Write(returnBody)
}
func JarmuvekUuid(ctx *fasthttp.RequestCtx, ps fasthttprouter.Params){
	var vData VehicleData
	db, err := sql.Open("postgres", connection_data)
	if err != nil {
		ctx.Error("sql_connection_error", fasthttp.StatusBadRequest)
	}
	defer db.Close()
	row := db.QueryRow(basic_sql_query + " WHERE uuid::text = '"+ps.ByName("uuid")+"'")
	var adatokArray pq.StringArray
	erra := row.Scan(&vData.Uuid,&vData.Rendszam,&vData.Tulajdonos,&vData.ForgalmiErvenyes,&adatokArray)
	if erra != nil {
		ctx.Error("zero",fasthttp.StatusNotFound)
		return
	}
	adatokStringArray := []string(adatokArray)
	vData.Adatok = make([]string, len(adatokStringArray))
	copy(vData.Adatok, adatokStringArray)
	jsonResponse, err := json.Marshal(vData)
    if err != nil {
        ctx.Error("first", fasthttp.StatusBadRequest)
        return
    }
    ctx.Response.Header.SetContentType("application/json")
    ctx.Write(jsonResponse)
}
func JarmuvekGet(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params){
	var count int
	db, err := sql.Open("postgres", connection_data)
	if err != nil {
		ctx.Response.SetBodyString("SQL CONNECTION ERROR")
	}
	defer db.Close()
	row := db.QueryRow("SELECT COUNT(*) FROM jarmuvek")
	erra := row.Scan(&count)
	if erra != nil {
		ctx.Error(fmt.Sprintf("Error: %v", erra), fasthttp.StatusBadRequest)
		return
	}
    ctx.Response.Header.SetContentType("text/plain")
    ctx.Response.SetBodyString(fmt.Sprint(count))
}
func Index(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params) {
	fmt.Println("Hello from:", os.Getenv("SERVERTITLE"))
	fmt.Fprint(ctx, "Welcome!\n")
}
func main() {
	r := fasthttprouter.New()
	r.GET("/", Index)
	r.GET("/jarmuvek", JarmuvekGet)
	r.GET("/kereses", JarmuvekKereses)
	r.POST("/jarmuvek", JarmuvekPost)
	r.GET("/jarmuvek/:uuid", JarmuvekUuid)
	fmt.Println("Server:" + os.Getenv("SERVERTITLE") + " started at :5000")
	fasthttp.ListenAndServe(":5000", r.Handler)
}