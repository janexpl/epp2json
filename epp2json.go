// Package epp2json provides functionality to convert EPP files to JSON format
package epp2json

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Invoice reprezentuje pojedynczą fakturę (FZ lub FS)
type Invoice struct {
	Type               string        `json:"typ"`
	Number             string        `json:"numer"`
	InternalNumber     string        `json:"numer_wewnetrzny"`
	DocumentNumber     string        `json:"numer_dokumentu"`
	Date               time.Time     `json:"data"`
	IssueDate          time.Time     `json:"data_wystawienia"`
	SaleDate           time.Time     `json:"data_sprzedazy"`
	ContractorCode     string        `json:"kod_kontrahenta"`
	ContractorName     string        `json:"nazwa_kontrahenta"`
	ContractorFullName string        `json:"pelna_nazwa_kontrahenta"`
	City               string        `json:"miasto"`
	PostalCode         string        `json:"kod_pocztowy"`
	Address            string        `json:"adres"`
	NIP                string        `json:"nip"`
	Category           string        `json:"category"`
	NetAmount          float64       `json:"kwota_netto"`
	VatAmount          float64       `json:"kwota_vat"`
	GrossAmount        float64       `json:"kwota_brutto"`
	Currency           string        `json:"waluta"`
	PaymentDate        time.Time     `json:"termin_platnosci"`
	Registrar          string        `json:"rejestrator"`
	Items              []InvoiceItem `json:"pozycje"`
}

// InvoiceItem reprezentuje pozycję faktury
type InvoiceItem struct {
	VatRate    string  `json:"stawka_vat"`
	Quantity   float64 `json:"ilosc"`
	NetPrice   float64 `json:"cena_netto"`
	VatAmount  float64 `json:"kwota_vat"`
	GrossPrice float64 `json:"cena_brutto"`
	NetTotal   float64 `json:"wartosc_netto"`
	VatTotal   float64 `json:"wartosc_vat"`
	GrossTotal float64 `json:"wartosc_brutto"`
}

// EPPData reprezentuje kompletne dane z pliku EPP
type EPPData struct {
	Info     map[string]string `json:"info"`
	Invoices []Invoice         `json:"faktury"`
}

// ParseOptions zawiera opcje parsowania
type ParseOptions struct {
	IncludeFZ bool // Czy dołączać faktury zakupowe
	IncludeFS bool // Czy dołączać faktury sprzedażowe
}
type Section struct {
	Header  string
	Content string
}
type EPPSections struct {
	Info     string
	Sections []Section
}

// DefaultParseOptions zwraca domyślne opcje parsowania (wszystkie typy faktur)
func DefaultParseOptions() ParseOptions {
	return ParseOptions{
		IncludeFZ: true,
		IncludeFS: true,
	}
}

// ParseDate parsuje datę z formatu YYYYMMDDHHMMSS do time.Time
func ParseDate(dateStr string) time.Time {
	if len(dateStr) == 14 {
		// Format: YYYYMMDDHHMMSS
		if t, err := time.Parse("20060102150405", dateStr); err == nil {
			return t
		}
	}
	return time.Time{}
}

// ParseFloat parsuje string na float64
func ParseFloat(str string) float64 {
	if val, err := strconv.ParseFloat(str, 64); err == nil {
		return val
	}
	return 0.0
}
func ParseSections(input string) EPPSections {
	const (
		headerTag  = "[NAGLOWEK]"
		contentTag = "[ZAWARTOSC]"
		infoTag    = "[INFO]"
	)

	var sections []Section
	var info string
	// rozbijamy tekst po każdym wystąpieniu tagu [NAGLOWEK]
	blocks := strings.Split(input, headerTag)
	// pierwszy blok przed pierwszym [NAGLOWEK] odrzucamy
	if len(blocks) > 0 {
		prefix := blocks[0]
		if idx := strings.Index(prefix, infoTag); idx >= 0 {
			// wszystko po [INFO] do końca prefixu to nasz info
			info = strings.TrimSpace(prefix[idx+len(infoTag):])
		}
	}
	for _, block := range blocks[1:] {
		// szukamy w nim [ZAWARTOSC]
		idx := strings.Index(block, contentTag)
		if idx < 0 {
			// brak tagu [ZAWARTOSC] – pomijamy
			continue
		}
		// wszystko przed [ZAWARTOSC] to header
		header := strings.TrimSpace(block[:idx])
		// wszystko po [ZAWARTOSC] to content
		content := strings.TrimSpace(block[idx+len(contentTag):])
		sections = append(sections, Section{
			Header:  header,
			Content: content,
		})
	}

	return EPPSections{
		Info:     info,
		Sections: sections,
	}
}

// ParseCSVLine parsuje linię CSV z obsługą cudzysłowów
func ParseCSVLine(line string) ([]string, error) {
	var result []string
	reader := csv.NewReader(strings.NewReader(line))
	result, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("błąd czytania CSV: %v", err)
	}

	return result, nil
}

// ParseHeader parsuje nagłówek faktury z pól CSV
func ParseHeader(fields []string) Invoice {
	invoice := Invoice{}

	if len(fields) > 0 {
		invoice.Type = fields[0]
	}
	if len(fields) > 4 {
		invoice.Number = fields[4]
	}
	if len(fields) > 6 {
		invoice.InternalNumber = fields[6]
	}
	if len(fields) > 11 {
		invoice.ContractorCode = fields[11]
	}
	if len(fields) > 12 {
		invoice.ContractorName = fields[12]
	}
	if len(fields) > 13 {
		invoice.ContractorFullName = fields[13]
	}
	if len(fields) > 14 {
		invoice.City = fields[14]
	}
	if len(fields) > 15 {
		invoice.PostalCode = fields[15]
	}
	if len(fields) > 16 {
		invoice.Address = fields[16]
	}
	if len(fields) > 17 {
		invoice.NIP = fields[17]
	}
	if len(fields) > 18 {
		invoice.Category = fields[19]
	}
	if len(fields) > 21 {
		invoice.Date = ParseDate(fields[21])
	}
	if len(fields) > 22 {
		invoice.IssueDate = ParseDate(fields[22])
	}
	if len(fields) > 23 {
		invoice.SaleDate = ParseDate(fields[23])
	}
	if len(fields) > 27 {
		invoice.NetAmount = ParseFloat(fields[27])
	}
	if len(fields) > 28 {
		invoice.VatAmount = ParseFloat(fields[28])
	}
	if len(fields) > 29 {
		invoice.GrossAmount = ParseFloat(fields[29])
	}
	if len(fields) > 34 {
		invoice.PaymentDate = ParseDate(fields[34])
	}
	if len(fields) > 41 {
		invoice.Registrar = fields[41]
	}
	if len(fields) > 46 {
		invoice.Currency = fields[46]
	}

	return invoice
}

// ParseItem parsuje pozycję faktury z pól CSV
func ParseItem(fields []string) InvoiceItem {
	item := InvoiceItem{}

	if len(fields) > 0 {
		item.VatRate = fields[0]
	}
	if len(fields) > 1 {
		item.Quantity = ParseFloat(fields[1])
	}
	if len(fields) > 2 {
		item.NetPrice = ParseFloat(fields[2])
	}
	if len(fields) > 3 {
		item.VatAmount = ParseFloat(fields[3])
	}
	if len(fields) > 4 {
		item.GrossPrice = ParseFloat(fields[4])
	}
	if len(fields) > 5 {
		item.NetTotal = ParseFloat(fields[5])
	}
	if len(fields) > 6 {
		item.VatTotal = ParseFloat(fields[6])
	}
	if len(fields) > 7 {
		item.GrossTotal = ParseFloat(fields[7])
	}

	return item
}

// ParseEPPFromString parsuje zawartość pliku EPP z stringa
func ParseEPPFromString(content string, options ParseOptions) (*EPPData, error) {
	sections := ParseSections(content)
	eppData := &EPPData{
		Info:     make(map[string]string),
		Invoices: []Invoice{},
	}

	currentInvoice := Invoice{}
	info := sections.Info
	fields, err := ParseCSVLine(info)
	if err != nil {
		return nil, fmt.Errorf("błąd podczas parsowania info: %v", err)
	}
	// Parse info
	if len(fields) >= 2 {
		eppData.Info["version"] = fields[0]
		if len(fields) > 3 {
			eppData.Info["system"] = fields[3]
		}
		if len(fields) > 5 {
			eppData.Info["company"] = fields[5]
		}
	}

	for _, section := range sections.Sections {
		line := strings.TrimSpace(section.Header)
		if line == "" {
			continue
		}
		fields, err := ParseCSVLine(line)
		if err != nil {
			return nil, fmt.Errorf("błąd podczas parsowania nagłówka: %v", err)
		}
		// Sprawdź czy to faktura FZ lub FS
		if len(fields) > 0 {
			invoiceType := fields[0]
			shouldInclude := (invoiceType == "FZ" || invoiceType == "KFZ" && options.IncludeFZ) ||
				(invoiceType == "FS" || invoiceType == "KFS" && options.IncludeFS)

			if shouldInclude {
				// Jeśli już mamy fakturę, dodaj ją do listy
				if currentInvoice.Type != "" {
					eppData.Invoices = append(eppData.Invoices, currentInvoice)
				}

				// Parsuj nowy nagłówek
				currentInvoice = ParseHeader(fields)
				currentInvoice.Items = []InvoiceItem{}
				fields, err = ParseCSVLine(section.Content)
				if err != nil {
					return nil, fmt.Errorf("błąd podczas parsowania pozycji: %v", err)
				}
				// Parsuj pozycje faktury
				if currentInvoice.Type != "" {
					item := ParseItem(fields)
					currentInvoice.Items = append(currentInvoice.Items, item)
				}
			}

		}

	}

	// Dodaj ostatnią fakturę
	if currentInvoice.Type != "" {
		eppData.Invoices = append(eppData.Invoices, currentInvoice)
	}

	return eppData, nil
}

// ParseEPPFile parsuje plik EPP i zwraca strukturę danych
func ParseEPPFile(filename string, options ParseOptions) (*EPPData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("nie można otworzyć pliku %s: %v", filename, err)
	}
	defer file.Close()

	decoder := transform.NewReader(file, charmap.Windows1250.NewDecoder())

	var contentBuilder strings.Builder
	if _, err := io.Copy(&contentBuilder, decoder); err != nil {
		return nil, fmt.Errorf("błąd podczas odczytu pliku: %v", err)
	}

	content := contentBuilder.String()
	return ParseEPPFromString(content, options)
}

func ConvertEPPDataToJSON(eppData *EPPData) (jsonData []byte, err error) {
	jsonData, err = json.MarshalIndent(eppData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("błąd podczas konwersji do JSON: %v", err)
	}
	return jsonData, nil
}

// ConvertEPPToJSON konwertuje plik EPP na JSON i zapisuje do pliku
func ConvertEPPToJSON(inputFile string, options ParseOptions) (jsonData []byte, err error) {
	// Parsuj plik EPP
	eppData, err := ParseEPPFile(inputFile, options)
	if err != nil {
		return nil, fmt.Errorf("błąd podczas parsowania pliku EPP: %v", err)
	}

	// Konwertuj do JSON
	jsonData, err = json.MarshalIndent(eppData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("błąd podczas konwersji do JSON: %v", err)
	}

	return jsonData, nil
}
func WriteJSONToFile(jsonData []byte, outputFile string) error {
	// Zapisz do pliku
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("błąd podczas zapisu pliku: %v", err)
	}
	return nil
}

// GetInvoiceStats zwraca statystyki faktur
func GetInvoiceStats(eppData *EPPData) (fzCount, fsCount int) {
	for _, invoice := range eppData.Invoices {
		switch invoice.Type {
		case "FZ":
			fzCount++
		case "FS":
			fsCount++
		}
	}
	return fzCount, fsCount
}

// FilterInvoices filtruje faktury według typu
func FilterInvoices(invoices []Invoice, invoiceType string) []Invoice {
	var filtered []Invoice
	for _, invoice := range invoices {
		if invoice.Type == invoiceType {
			filtered = append(filtered, invoice)
		}
	}
	return filtered
}

// GetInvoicesByContractor zwraca faktury dla określonego kontrahenta
func GetInvoicesByContractor(invoices []Invoice, contractorCode string) []Invoice {
	var filtered []Invoice
	for _, invoice := range invoices {
		if strings.Contains(invoice.ContractorCode, contractorCode) {
			filtered = append(filtered, invoice)
		}
	}
	return filtered
}

// GetInvoicesByDateRange zwraca faktury z określonego zakresu dat
func GetInvoicesByDateRange(invoices []Invoice, start, end time.Time) []Invoice {
	var filtered []Invoice
	for _, invoice := range invoices {
		if !invoice.IssueDate.IsZero() &&
			invoice.IssueDate.After(start) &&
			invoice.IssueDate.Before(end) {
			filtered = append(filtered, invoice)
		}
	}
	return filtered
}

// CalculateTotalAmount oblicza łączną kwotę dla listy faktur
func CalculateTotalAmount(invoices []Invoice) (net, vat, gross float64) {
	for _, invoice := range invoices {
		net += invoice.NetAmount
		vat += invoice.VatAmount
		gross += invoice.GrossAmount
	}
	return net, vat, gross
}

// GroupInvoicesByMonth grupuje faktury według miesięcy
func GroupInvoicesByMonth(invoices []Invoice) map[string][]Invoice {
	grouped := make(map[string][]Invoice)

	for _, invoice := range invoices {
		if !invoice.IssueDate.IsZero() {
			month := invoice.IssueDate.Format("2006-01")
			grouped[month] = append(grouped[month], invoice)
		}
	}

	return grouped
}
