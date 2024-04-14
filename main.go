package main

import (
	"encoding/csv"
	"fmt"
	"github.com/gocolly/colly"
	"os"
	"strings"
	"sync"
)

type Commune struct {
	name         string
	link         string
	contactEmail string
}
type Province struct {
	name string
	link string
}

var ProvinceList = []Province{
	{
		name: "Como",
		link: "http://www.comuni-italiani.it/013/",
	},
	{
		name: "Bergamo",
		link: "http://www.comuni-italiani.it/016/",
	},
	{
		name: "Brescia",
		link: "http://www.comuni-italiani.it/017/",
	},
	{
		name: "Monza_e_della_Brianza",
		link: "http://www.comuni-italiani.it/108/",
	},
	{
		name: "Varese",
		link: "http://www.comuni-italiani.it/012/",
	},
	{
		name: "Sondrio",
		link: "http://www.comuni-italiani.it/014/",
	},
	{
		name: "Verbano_Cusio_Ossola",
		link: "http://www.comuni-italiani.it/103/",
	},
	{
		name: "Vercelli",
		link: "http://www.comuni-italiani.it/002/",
	},
}

func main() {
	// for _, province := range ProvinceList {
	// 	csvName, done := createCommunesCSV(province)
	// 	if done {
	// 		fromCSVComunesExtractEmailIFExist(csvName, province)
	// 	}
	// }
	var wg sync.WaitGroup

	for _, province := range ProvinceList {
		csvName, done := createCommunesCSV(province)
		if done {
			wg.Add(1)
			go func(csvName string, province Province) {
				defer wg.Done()
				fromCSVComunesExtractEmailIFExist(csvName, province)
			}(csvName, province)
		}
	}

	wg.Wait()
}

func createCommunesCSV(p Province) (string, bool) {

	c := colly.NewCollector()
	var communeTable []Commune

	c.OnHTML("tbody > tr:nth-child(2) > td > table", func(e *colly.HTMLElement) {
		count := 0
		e.ForEach("tr", func(i int, e *colly.HTMLElement) {

			if count == 0 {
				count++
				return
			}
			e.ForEach("td", func(i int, e *colly.HTMLElement) {
				fmt.Println(e.Text)
				fmt.Println(e.ChildAttr("a", "href"))
				commune := Commune{
					name: e.ChildText("a"),
					link: p.link + e.ChildAttr("a", "href"),
				}
				fmt.Println("Commune:", commune)
				communeTable = append(communeTable, commune)
			})
		})
	})
	err := c.Visit(p.link)
	if err != nil {
		return "", false
	}
	fileName := p.name + ".csv"
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error:", err)
		return fileName, true
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error:", err)
		}
	}(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, comune := range communeTable {
		err := writer.Write([]string{comune.name, comune.link})
		if err != nil {
			return "", false
		}
	}
	return fileName, true
}

func fromCSVComunesExtractEmailIFExist(filename string, province Province) {
	// create a new colly collector

	// read the csv file
	csvFile, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// iterate over the records
	newFile, err := os.Create(province.name + "_" + "with_email.csv")
	if err != nil {
		fmt.Println("Error:", err)
	}
	writer := csv.NewWriter(newFile)
	defer writer.Flush()
	for _, record := range records {
		comune := Commune{
			name:         record[0],
			link:         record[1],
			contactEmail: "",
		}
		cE := colly.NewCollector()
		cE.OnHTML("a:contains('Email Comune')", func(e *colly.HTMLElement) {
			href := e.Attr("href")
			fmt.Println("", href)
			trimedHref := strings.TrimPrefix(href, "mailto:")
			fmt.Println(trimedHref)
			comune.contactEmail = trimedHref
			record = []string{comune.name, comune.link, comune.contactEmail}
			fmt.Println("Record:", record)
			err := writer.Write(record)
			if err != nil {
				return
			}
		})

		err := cE.Visit(comune.link)
		if err != nil {
			return
		}
	}
}
