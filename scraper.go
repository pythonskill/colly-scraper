package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/magiconair/properties"
)

func main() {
	log.Print("Start scraping")

	// Читаем настройки
	p := properties.MustLoadFile("scraper.properties", properties.UTF8)
	title, _ := p.Get("title")
	fmt.Print(title)

	chapterSizeCollector := colly.NewCollector(
		colly.AllowedDomains("mangabuff.ru"),
	)

	chapterSizeCollector.OnHTML("body > div.wrapper > div.main > div > div.row > div.col.l9.l12 > div > div.manga__middle > div.tabs > div.tabs__nav > button:nth-child(2)", func(h *colly.HTMLElement) {
		mangaSizeText := h.Text
		if len(mangaSizeText) == 0 {
			return
		}
		re := regexp.MustCompile("[0-9]+")

		mangaSize, err := strconv.Atoi(re.FindString(mangaSizeText))
		if err != nil {
			log.Fatal("Parsing manga size Error!!!")
			return
		}

		fmt.Printf("Total chapters count: %d", mangaSize)

		imageCollector := colly.NewCollector(
			colly.AllowedDomains("mangabuff.ru", "c2.mangabuff.ru"),
		)

		// Стащим все изображения по универсальному селектору, куда не попадают левые изображения
		imageCollector.OnHTML("body > div.reader > div.reader__container > div.reader__pages > div > img", func(h *colly.HTMLElement) {
			imgSrc := h.Attr("data-src")
			if len(imgSrc) == 0 {
				imgSrc = h.Attr("src")
			}
			imgSrc = h.Request.AbsoluteURL(imgSrc)

			fmt.Printf("Image found: %s\n", imgSrc)
			os.Mkdir("./images", os.ModePerm)
			fileName := "images/" + h.Attr("alt") + ".jpg"
			// err := c.Visit(imgSrc)
			// if err != nil {
			// 	log.Printf("Failed to download image %s: %s", imgSrc, err)
			// 	return
			// }
			downloadFile(imgSrc, fileName)
			// if errDownload != nil {
			// 	log.Fatal(err)
			// }
			// delay for "humanize" process of scraping, if dont need then set value = 0
			//defer time.Sleep(1 * time.Second)
		})
		imageCollector.OnRequest(func(r *colly.Request) {
			fmt.Println("Downloading them all...", r.URL.String())
		})
		// Пройдемся циклом по всем страницам глав
		for i := 1; i < mangaSize; i++ {
			imageCollector.Visit(fmt.Sprintf("https://mangabuff.ru/manga/%s/1/%d", title, i))
		}

	})

	chapterSizeCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	// Узнаем сколько глав в наличии и колбек на это действие образует контекст, в рамках которого уже новым
	// коллектором запустим скачивание ;-) ну ты понялпо ебанутому все, то есть я хотел сказать в стиле GO
	chapterSizeCollector.Visit(fmt.Sprintf("https://mangabuff.ru/manga/%s/", title))
}

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	log.Printf("Creating file %s", fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
