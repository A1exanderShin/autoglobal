package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/A1exanderShin/autoglobal/internal/cars/dto"
	"github.com/A1exanderShin/autoglobal/internal/cars/service"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
)

type Parser struct {
	httpClient *http.Client
	carService *service.Service
}

func NewParser(carService *service.Service) *Parser {
	return &Parser{
		httpClient: &http.Client{},
		carService: carService,
	}
}

func (p *Parser) ParsePage(ctx context.Context, url string) (int, error) {
	count := 0

	fmt.Println("[parser] start:", url)

	// ---- 1. HTTP запрос ----
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	// --- REAL USER-AGENTS ---
	uaList := []string{
		// Windows Chrome
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",

		// macOS Safari
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3 Safari/605.1.15",

		// Linux Chrome
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",

		// Windows Firefox
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0",
	}

	// Ставим случайный настоящий User-Agent
	req.Header.Set("User-Agent", uaList[rand.Intn(len(uaList))])

	// ---- FULL BROWSER HEADERS ----
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("[parser] status code:", resp.StatusCode)

	// ---- 2. Чтение тела ----
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read body error: %w", err)
	}

	// ---- 3. Декодировка Windows-1251 → UTF-8 ----
	decoder := charmap.Windows1251.NewDecoder()
	utf8Body, err := decoder.Bytes(bodyBytes)
	if err != nil {
		return 0, fmt.Errorf("decode error: %w", err)
	}

	// ---- 4. Парсинг HTML ----
	reader := bytes.NewReader(utf8Body)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return 0, fmt.Errorf("goquery error: %w", err)
	}

	// ---- 5. Поиск карточек ----
	cards := doc.Find("[data-ftid='bulls-list_bull']").Length()
	fmt.Println("[parser] found cards:", cards)

	doc.Find("[data-ftid='bulls-list_bull']").Each(func(i int, s *goquery.Selection) {

		// ---- title + price ----
		title := strings.TrimSpace(s.Find("[data-ftid='bull_title']").Text())
		price := strings.TrimSpace(s.Find("[data-ftid='bull_price']").Text())

		if title == "" || price == "" {
			return
		}

		// ---- title: Brand Model, Year ----
		parts := strings.Split(title, ",")
		if len(parts) < 2 {
			return
		}

		main := strings.TrimSpace(parts[0])
		yearStr := strings.TrimSpace(parts[1])

		year, _ := strconv.Atoi(yearStr)
		if year == 0 {
			return
		}

		words := strings.Split(main, " ")
		if len(words) < 2 {
			return
		}

		brand := words[0]
		model := strings.Join(words[1:], " ")

		// ---- PRICE ----
		raw := strings.ReplaceAll(price, " ", "")
		raw = strings.ReplaceAll(raw, "\u00A0", "")
		raw = strings.ReplaceAll(raw, "₽", "")
		raw = strings.TrimSpace(raw)

		priceInt, _ := strconv.Atoi(raw)
		if priceInt <= 0 {
			return
		}

		// TODO: add deduplication (check if car exists in DB)

		fmt.Println("------ CAR ------")
		fmt.Println("TITLE:", title)
		fmt.Println("BRAND:", brand)
		fmt.Println("MODEL:", model)
		fmt.Println("YEAR:", year)
		fmt.Println("PRICE:", priceInt)

		// ---- recording ----
		count++

		req := dto.CreateCarRequest{
			Brand: brand,
			Model: model,
			Year:  year,
			Price: priceInt,
		}

		id, err := p.carService.CreateCar(ctx, req)
		if err != nil {
			fmt.Println("[parser] save error:", err)
			return
		}

		fmt.Println("[parser] saved car with ID:", id)
	})

	return count, nil
}

func (p *Parser) ParseAll(ctx context.Context, baseURL string, maxPages int) error {
	originalURL := baseURL

	for page := 1; page <= maxPages; page++ {

		var url string

		if page == 1 {
			url = originalURL
		} else {
			if strings.Contains(originalURL, "?") {
				url = originalURL + "&page=" + strconv.Itoa(page)
			} else {
				url = originalURL + "?page=" + strconv.Itoa(page)
			}
		}

		fmt.Println("[parser] parsing:", url)

		count, err := p.ParsePage(ctx, url)
		if err != nil {
			return err
		}

		// НЕТ объявлений → стоп
		if count == 0 {
			break
		}

		// ---- Progress bar ----
		progress := float64(page) / float64(maxPages)
		percent := int(progress * 100)
		bars := int(progress * 20)

		bar := strings.Repeat("#", bars) + strings.Repeat(".", 20-bars)

		fmt.Printf("\r[%s] %d%% (%d/%d страниц)", bar, percent, page, maxPages)

		time.Sleep(time.Duration(300+rand.Intn(1200)) * time.Millisecond)
	}

	fmt.Println("\n[parser] finished — no more pages")
	return nil
}
