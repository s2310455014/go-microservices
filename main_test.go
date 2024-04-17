// main_test.go

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
)

var a App

func TestMain(m *testing.M) {
	a.Initialize(
		"elehna",
		"elehna",
		"postgres")

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/product/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {

	clearTable()

	var jsonStr = []byte(`{"name":"test product", "price": 11.22}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}

	body, _ := ioutil.ReadAll(response.Body)
	fmt.Print("\t" + string(body) + "\n")

}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	body, _ := ioutil.ReadAll(response.Body)
	fmt.Print("\t" + string(body) + "\n")

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addProducts(count int) {
	fmt.Printf("Adding %d products into database.\n", count)

	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
	}
}

func TestUpdateProduct(t *testing.T) {

	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	var jsonStr = []byte(`{"name":"test product - updated name", "price": 11.22}`)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
	}

	if m["name"] == originalProduct["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
	}

	if m["price"] == originalProduct["price"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
	}

	body, _ := ioutil.ReadAll(response.Body)
	fmt.Print("\t" + string(body) + "\n")
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)

	body, _ := ioutil.ReadAll(response.Body)
	fmt.Print("\t" + string(body) + "\n")

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetAllProducts(t *testing.T) {
	clearTable()
	addProducts(3) // Fügt 3 Produkte hinzu

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("\tResponse Body:", string(body))

	var products []map[string]interface{}
	if err := json.Unmarshal(body, &products); err != nil {
		t.Errorf("Error parsing json response: %s", err)
	}

	if len(products) != 3 {
		t.Errorf("Expected 3 products. Got %d", len(products))
	}
}

func TestGetNrProducts(t *testing.T) {
	clearTable()
	addProducts(5) // Fügt 5 Produkte hinzu

	req, _ := http.NewRequest("GET", "/products/count", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var result map[string]int
	err := json.Unmarshal(response.Body.Bytes(), &result)
	if err != nil {
		t.Fatal("Cannot parse json response:", err)
	}

	expectedCount := 5
	if result["Number of products"] != expectedCount {
		t.Errorf("Expected number of products to be %d. Got %d", expectedCount, result["Number of products"])
	}
}

func TestGetMostExpensiveProduct(t *testing.T) {
	clearTable()
	addProducts(3) // Fügt Produkte mit ansteigenden Preisen hinzu

	req, _ := http.NewRequest("GET", "/products/expensiveProduct", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var product map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &product)
	if err != nil {
		t.Fatal("Cannot parse json response:", err)
	}

	if product["price"] != float64(30) {
		t.Errorf("Expected the price of the most expensive product to be 30.00. Got %.2f", product["price"])
	}
}

func TestGetCheapestProduct(t *testing.T) {
	clearTable()
	addProducts(3) // Fügt Produkte mit ansteigenden Preisen hinzu

	req, _ := http.NewRequest("GET", "/products/cheapProduct", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var product map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &product)
	if err != nil {
		t.Fatal("Cannot parse json response:", err)
	}

	if product["price"] != float64(10) {
		t.Errorf("Expected the price of the cheapest product to be 10.00. Got %.2f", product["price"])
	}
}
