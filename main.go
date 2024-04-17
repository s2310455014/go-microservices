// main.go

package main

func main() {
	a := App{}
	//a.Initialize(
	//	os.Getenv("APP_DB_USERNAME"),
	//	os.Getenv("APP_DB_PASSWORD"),
	//	os.Getenv("APP_DB_NAME"))
	a.Initialize(
		"elehna",
		"elehna",
		"postgres")

	//export APP_DB_USERNAME=postgres
	//export APP_DB_PASSWORD=<whatever password you use>
	//export APP_DB_NAME=postgres

	//List env variables (bash): printenv
	a.Run(":8010")
}
