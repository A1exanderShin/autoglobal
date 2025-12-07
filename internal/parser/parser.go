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

	// ==== HTTP REQUEST ====
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	uaList := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3 Safari/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0",
	}
	req.Header.Set("User-Agent", uaList[rand.Intn(len(uaList))])

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("[parser] status code:", resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read body error: %w", err)
	}

	decoder := charmap.Windows1251.NewDecoder()
	utf8Body, err := decoder.Bytes(bodyBytes)
	if err != nil {
		return 0, fmt.Errorf("decode error: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(utf8Body))
	if err != nil {
		return 0, fmt.Errorf("goquery error: %w", err)
	}

	// ==== CARD SELECTOR ====
	cards := doc.Find("div[data-ftid='bulls-list_bull']")
	fmt.Println("[parser] found cards:", cards.Length())

	if cards.Length() == 0 {
		fmt.Println("[parser] NO CARDS FOUND — HTML structure changed")
		return 0, nil
	}

	cards.Each(func(i int, s *goquery.Selection) {
		// ==== URL ====
		link, ok := s.Find("a[data-ftid='bull_title']").Attr("href")
		if !ok || link == "" {
			fmt.Println("[SKIP] no URL")
			return
		}

		// ==== TITLE ====
		title := strings.TrimSpace(s.Find("a[data-ftid='bull_title']").Text())
		if title == "" {
			fmt.Println("[SKIP] no title")
			return
		}

		// ==== PRICE ====
		priceStr := strings.TrimSpace(s.Find("span[data-ftid='bull_price']").Text())
		if priceStr == "" {
			fmt.Println("[SKIP] no price")
			return
		}

		// Parse price
		raw := strings.ReplaceAll(priceStr, " ", "")
		raw = strings.ReplaceAll(raw, "\u00A0", "")
		raw = strings.ReplaceAll(raw, "₽", "")
		price, _ := strconv.Atoi(raw)
		if price <= 0 {
			fmt.Println("[SKIP] invalid price")
			return
		}

		// ==== Parse brand/model/year ====
		// Example: "Volkswagen Tiguan, 2021"
		parts := strings.Split(title, ",")
		if len(parts) < 2 {
			fmt.Println("[SKIP] title format invalid:", title)
			return
		}

		main := strings.TrimSpace(parts[0])
		yearStr := strings.TrimSpace(parts[1])

		year, _ := strconv.Atoi(yearStr)
		if year == 0 {
			fmt.Println("[SKIP] invalid year:", yearStr)
			return
		}

		name := strings.Split(main, " ")
		if len(name) < 2 {
			fmt.Println("[SKIP] cannot split brand/model:", main)
			return
		}

		brand := name[0]
		model := strings.Join(name[1:], " ")

		// ==== DEDUP BY URL ====
		exists, err := p.carService.ExistsByURL(ctx, link)
		if err != nil {
			fmt.Println("[ERROR] ExistsByURL:", err)
			return
		}
		if exists {
			fmt.Println("[SKIP] duplicate:", link)
			return
		}

		fmt.Println("=== NEW CAR ===")
		fmt.Println("URL:", link)
		fmt.Println("TITLE:", title)
		fmt.Println("BRAND:", brand)
		fmt.Println("MODEL:", model)
		fmt.Println("YEAR:", year)
		fmt.Println("PRICE:", price)

		// ==== SAVE ====
		req := dto.CreateCarRequest{
			Brand: brand,
			Model: model,
			Year:  year,
			Price: price,
			URL:   link,
		}

		id, err := p.carService.CreateCar(ctx, req)
		if err != nil {
			fmt.Println("[ERROR] save:", err)
			return
		}

		fmt.Println("[OK] saved:", id)
		count++
	})

	return count, nil
}

func (p *Parser) ParseAll(ctx context.Context, baseURL string, maxPages int) error {
	base := strings.TrimSuffix(baseURL, "/")

	for page := 1; page <= maxPages; page++ {

		url := base
		if page > 1 {
			url = fmt.Sprintf("%s/page%d/", base, page)
		}

		fmt.Println("[parser] parsing:", url)

		count, err := p.ParsePage(ctx, url)
		if err != nil {
			return err
		}

		if count == 0 {
			fmt.Println("[parser] STOP — no cars on page")
			break
		}

		time.Sleep(time.Duration(300+rand.Intn(1200)) * time.Millisecond)
	}

	fmt.Println("[parser] finished")
	return nil
}
