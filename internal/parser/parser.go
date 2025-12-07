package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/A1exanderShin/autoglobal/internal/cars/dto"
	"github.com/A1exanderShin/autoglobal/internal/cars/service"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
)

// Parser — объект, который отвечает за скачивание HTML,
// парсинг данных и запись результатов через сервисный слой.
type Parser struct {
	httpClient *http.Client     // клиент для HTTP-запросов
	carService *service.Service // сервисный слой — сюда будем сохранять данные
}

// Конструктор Parser — создаёт экземпляр с нужными зависимостями.
func NewParser(carService *service.Service) *Parser {
	return &Parser{
		httpClient: &http.Client{}, // создаём http.Client
		carService: carService,     // передаём сервис машин
	}
}

// ParsePage — основной метод: скачивает HTML, парсит объявления, сохраняет их в БД.
func (p *Parser) ParsePage(ctx context.Context, url string) error {
	fmt.Println("[parser] start:", url)

	// 1. ДЕЛАЕМ HTTP-ЗАПРОС
	resp, err := p.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	// 2. ЧИТАЕМ HTML КАК []byte
	bodyBytes, _ := io.ReadAll(resp.Body)

	// 3. ИСПРАВЛЯЕМ КОДИРОВКУ Drom (Windows-1251 → UTF-8)
	decoder := charmap.Windows1251.NewDecoder()
	utf8Body, err := decoder.Bytes(bodyBytes)
	if err != nil {
		return fmt.Errorf("decode error: %w", err)
	}

	// 4. ПЕРЕДАЁМ HTML В goquery
	reader := bytes.NewReader(utf8Body)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return fmt.Errorf("goquery error: %w", err)
	}

	// 5. НАХОДИМ ВСЕ КАРТОЧКИ АВТО
	doc.Find("[data-ftid='bulls-list_bull']").Each(func(i int, s *goquery.Selection) {

		// 5.1 ТЕКСТ заголовка: "Toyota Camry, 2018"
		title := strings.TrimSpace(s.Find("[data-ftid='bull_title']").Text())

		// 5.2 Цена: "1 200 000 ₽"
		price := strings.TrimSpace(s.Find("[data-ftid='bull_price']").Text())

		if title == "" || price == "" {
			return // пропускаем битые элементы
		}

		// ---------------------------------
		// 6. ПАРСИНГ НАЗВАНИЯ: Brand + Model + Year
		// ---------------------------------

		parts := strings.Split(title, ",")
		if len(parts) < 2 {
			return // формат кривой
		}

		main := strings.TrimSpace(parts[0]) // например: "Toyota Camry"
		yearStr := strings.TrimSpace(parts[1])

		// год
		year, _ := strconv.Atoi(yearStr)

		// бренд и модель
		words := strings.Split(main, " ")
		if len(words) < 2 {
			return
		}

		brand := words[0]                     // Toyota
		model := strings.Join(words[1:], " ") // Camry

		// ---------------------------------
		// 7. ОЧИСТКА СТРОКИ С ЦЕНОЙ
		// ---------------------------------

		raw := price
		raw = strings.ReplaceAll(raw, " ", "")      // убираем пробелы
		raw = strings.ReplaceAll(raw, "\u00A0", "") // убираем неразрывный пробел
		raw = strings.ReplaceAll(raw, "₽", "")      // убираем знак рубля
		raw = strings.TrimSpace(raw)

		priceInt, _ := strconv.Atoi(raw)

		// ЛОГ ДЛЯ ОТЛАДКИ
		fmt.Println("------ CAR ------")
		fmt.Println("TITLE:", title)
		fmt.Println("BRAND:", brand)
		fmt.Println("MODEL:", model)
		fmt.Println("YEAR:", year)
		fmt.Println("PRICE:", priceInt)

		// ---------------------------------
		// 8. СОХРАНЕНИЕ В БД через сервис
		// ---------------------------------

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

	return nil
}
