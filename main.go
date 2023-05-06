package main

import (
	"os"
	"strings"

	"github.com/jayespace/jobScrapper/scrapper"
	"github.com/labstack/echo/v4"
)

const jobFile string = "jobs.csv"

func handleHome(c echo.Context) error {
	return c.File("home.html")
}

func handleScrap(c echo.Context) error {
	defer os.Remove(jobFile)
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrap(term)
	return c.Attachment(jobFile, jobFile)
}

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrap", handleScrap)
	e.Logger.Fatal(e.Start(":1323"))
}
