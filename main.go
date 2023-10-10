package main

import (
	"github.com/DifuseHQ/difuse-geoip/src/services"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"log"
	"sync"
	"time"
)

var mu sync.Mutex
var isDownloading bool

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		mu.Lock()
		isDownloading = true
		mu.Unlock()

		c.Set("Content-Disposition", "attachment; filename=geoip2.tar.gz")
		err := c.SendFile("data/geoip2.tar.gz", true)

		mu.Lock()
		isDownloading = false
		mu.Unlock()

		return err
	})

	app.Get("/mmdb", func(c *fiber.Ctx) error {
		mu.Lock()
		isDownloading = true
		mu.Unlock()

		c.Set("Content-Disposition", "attachment; filename=geoip.mmdb")
		err := c.SendFile("data_mmdb/geoip2_mmdb/GeoLite2-Country.mmdb", true)

		mu.Lock()
		isDownloading = false
		mu.Unlock()

		return err
	})

	go services.DownloadAndExtractMMDB()
	go services.DownloadAndExtract()

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for {
			<-ticker.C
			mu.Lock()
			if !isDownloading {
				services.DownloadAndExtractMMDB()
				services.DownloadAndExtract()
			}
			mu.Unlock()
		}
	}()

	log.Fatal(app.Listen(":3000"))
}
