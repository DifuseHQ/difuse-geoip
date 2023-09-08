package services

import (
	"encoding/csv"
	"fmt"
	"github.com/mholt/archiver/v3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("Error closing body: %v", err)
		}
	}(resp.Body)

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Fatalf("Error closing file: %v", err)
		}
	}(out)

	_, err = io.Copy(out, resp.Body)
	return err
}

func compressGeoIPData() error {
	archiveFilePath := "data/geoip2.tar.gz"
	folderToArchive := "data/geoip2"

	err := archiver.Archive([]string{folderToArchive}, archiveFilePath)
	if err != nil {
		return fmt.Errorf("failed to archive directory: %w", err)
	}

	return nil
}

func DownloadAndExtract() {
	url := "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country-CSV&license_key=" + os.Getenv("LICENSE_KEY") + "&suffix=zip"
	zipFilePath := "data.zip"
	outputFolder := "data"

	err := downloadFile(zipFilePath, url)
	if err != nil {
		log.Fatalf("Error downloading file: %v", err)
	}

	if _, err := os.Stat(outputFolder); !os.IsNotExist(err) {
		err = os.RemoveAll(outputFolder)
		if err != nil {
			log.Fatalf("Error deleting output folder: %v", err)
		}
	}

	err = archiver.NewZip().Unarchive(zipFilePath, outputFolder)
	if err != nil {
		log.Fatalf("Error extracting file: %v", err)
	}

	err = os.Remove(zipFilePath)
	if err != nil {
		log.Fatalf("Error removing zip file: %v", err)
	}

	directoryContent, err := ioutil.ReadDir(outputFolder)
	if err != nil {
		log.Fatalf("Error reading data directory: %v", err)
	}

	var extractedFolderName string
	for _, item := range directoryContent {
		if item.IsDir() && strings.HasPrefix(item.Name(), "GeoLite2-Country-CSV_") {
			extractedFolderName = item.Name()
			break
		}
	}

	if extractedFolderName == "" {
		log.Fatalf("Extracted folder not found")
	}

	extractedFolderPath := filepath.Join(outputFolder, extractedFolderName)
	err = extractDataByCountry(
		filepath.Join(extractedFolderPath, "GeoLite2-Country-Blocks-IPv4.csv"),
		filepath.Join(extractedFolderPath, "GeoLite2-Country-Blocks-IPv6.csv"),
		filepath.Join(extractedFolderPath, "GeoLite2-Country-Locations-en.csv"),
	)
	if err != nil {
		log.Fatalf("Error extracting data by country: %v", err)
	}

	err = compressGeoIPData()
	if err != nil {
		log.Fatalf("Error compressing data: %v", err)
	}

	err = os.RemoveAll("data/geoip2")
	if err != nil {
		log.Fatalf("Error deleting geoip2 folder: %v", err)
	}

	err = os.RemoveAll(filepath.Join("data", extractedFolderName))
	if err != nil {
		log.Fatalf("Error deleting GeoLite2 folder: %v", err)
	}
}

func extractDataByCountry(ipv4File, ipv6File, locationFile string) error {
	geonameMap := make(map[string]string)
	extractedDataFolder := "data/geoip2"

	file, err := os.Open(locationFile)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		geonameMap[record[0]] = record[4]
	}

	err = extractDataByCountryFromFile(ipv4File, geonameMap, filepath.Join(extractedDataFolder, "ipv4"))
	if err != nil {
		return err
	}

	err = extractDataByCountryFromFile(ipv6File, geonameMap, filepath.Join(extractedDataFolder, "ipv6"))
	if err != nil {
		return err
	}

	return nil
}

func extractDataByCountryFromFile(filePath string, geonameMap map[string]string, outputFolder string) error {
	err := os.MkdirAll(outputFolder, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Error closing file: %v", err)
		}
	}(file)

	reader := csv.NewReader(file)

	countryDataMap := make(map[string]*os.File)
	defer func() {
		for _, f := range countryDataMap {
			err := f.Close()
			if err != nil {
				log.Fatalf("Error closing file: %v", err)
			}
		}
	}()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		countryCode := geonameMap[record[1]]
		if countryCode == "" || countryCode == "country_iso_code" {
			continue
		}

		countryFile, ok := countryDataMap[countryCode]

		if !ok {
			countryFile, err = os.Create(filepath.Join(outputFolder, countryCode+".zone"))
			if err != nil {
				return err
			}
			countryDataMap[countryCode] = countryFile
		}

		_, err = countryFile.WriteString(record[0] + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}
