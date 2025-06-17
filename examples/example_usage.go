// Package main demonstruje użycie biblioteki epp2json w innym programie
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/janexpl/epp2json"
)

// Przykład 1: Parsowanie tylko faktur zakupowych (FZ)
func parseOnlyPurchaseInvoices() {
	fmt.Println("=== Przykład 1: Tylko faktury zakupowe (FZ) ===")

	options := epp2json.ParseOptions{
		IncludeFZ: true,
		IncludeFS: false,
	}

	eppData, err := epp2json.ParseEPPFile("../eksport.epp", options)
	if err != nil {
		log.Fatal("Błąd parsowania:", err)
	}

	fzCount, fsCount := epp2json.GetInvoiceStats(eppData)
	fmt.Printf("Faktury FZ: %d, FS: %d\n", fzCount, fsCount)

	// Wyświetl pierwszą fakturę zakupową
	if len(eppData.Invoices) > 0 {
		firstInvoice := eppData.Invoices[0]
		fmt.Printf("Pierwsza faktura: %s - %s (%.2f PLN)\n",
			firstInvoice.Number, firstInvoice.ContractorName, firstInvoice.GrossAmount)
	}
}

// Przykład 2: Filtrowanie faktur według kontrahenta
func filterByContractor() {
	fmt.Println("\n=== Przykład 2: Faktury dla konkretnego kontrahenta ===")

	eppData, err := epp2json.ParseEPPFile("../eksport.epp", epp2json.DefaultParseOptions())
	if err != nil {
		log.Fatal("Błąd parsowania:", err)
	}

	// Znajdź faktury dla PKO LEASING
	contractorInvoices := epp2json.GetInvoicesByContractor(eppData.Invoices, "PKO LEASING")

	fmt.Printf("Znaleziono %d faktur dla PKO LEASING\n", len(contractorInvoices))

	totalAmount := 0.0
	for _, invoice := range contractorInvoices {
		totalAmount += invoice.GrossAmount
		fmt.Printf("- %s: %.2f PLN (%s)\n",
			invoice.Number, invoice.GrossAmount, invoice.Type)
	}

	fmt.Printf("Łączna kwota: %.2f PLN\n", totalAmount)
}

// Przykład 3: Analiza faktur według miesięcy
func analyzeByMonth() {
	fmt.Println("\n=== Przykład 3: Analiza według miesięcy ===")

	eppData, err := epp2json.ParseEPPFile("../eksport.epp", epp2json.DefaultParseOptions())
	if err != nil {
		log.Fatal("Błąd parsowania:", err)
	}

	grouped := epp2json.GroupInvoicesByMonth(eppData.Invoices)

	fmt.Println("Liczba faktur według miesięcy:")
	for month, invoices := range grouped {
		_, _, total := epp2json.CalculateTotalAmount(invoices)
		fmt.Printf("- %s: %d faktur (%.2f PLN)\n", month, len(invoices), total)
	}
}

// Przykład 4: Eksport do niestandardowego formatu JSON
func customJSONExport() {
	fmt.Println("\n=== Przykład 4: Niestandardowy eksport ===")

	eppData, err := epp2json.ParseEPPFile("../eksport.epp", epp2json.DefaultParseOptions())
	if err != nil {
		log.Fatal("Błąd parsowania:", err)
	}

	// Utwórz uproszczoną strukturę
	type SimpleInvoice struct {
		Type       string    `json:"typ"`
		Number     string    `json:"numer"`
		Contractor string    `json:"kontrahent"`
		Amount     float64   `json:"kwota"`
		Date       time.Time `json:"data"`
		ItemsCount int       `json:"liczba_pozycji"`
	}

	var simpleInvoices []SimpleInvoice

	for _, invoice := range eppData.Invoices {
		simple := SimpleInvoice{
			Type:       invoice.Type,
			Number:     invoice.Number,
			Contractor: invoice.ContractorName,
			Amount:     invoice.GrossAmount,
			Date:       invoice.IssueDate,
			ItemsCount: len(invoice.Items),
		}
		simpleInvoices = append(simpleInvoices, simple)
	}

	// Zapisz do pliku
	jsonData, err := json.MarshalIndent(simpleInvoices, "", "  ")
	if err != nil {
		log.Fatal("Błąd JSON:", err)
	}

	err = os.WriteFile("uproszczone_faktury.json", jsonData, 0644)
	if err != nil {
		log.Fatal("Błąd zapisu:", err)
	}

	fmt.Printf("Zapisano %d uproszczonych faktur do pliku 'uproszczone_faktury.json'\n",
		len(simpleInvoices))
}

// Przykład 5: Walidacja danych
func validateData() {
	fmt.Println("\n=== Przykład 5: Walidacja danych ===")

	eppData, err := epp2json.ParseEPPFile("../eksport.epp", epp2json.DefaultParseOptions())
	if err != nil {
		log.Fatal("Błąd parsowania:", err)
	}

	var invalidInvoices []string

	for _, invoice := range eppData.Invoices {
		// Sprawdź czy faktura ma prawidłowe dane
		if invoice.Number == "" {
			invalidInvoices = append(invalidInvoices,
				fmt.Sprintf("Brak numeru faktury (typ: %s)", invoice.Type))
		}

		if invoice.GrossAmount < 0 {
			invalidInvoices = append(invalidInvoices,
				fmt.Sprintf("Ujemna kwota w fakturze %s", invoice.Number))
		}

		if invoice.ContractorName == "" {
			invalidInvoices = append(invalidInvoices,
				fmt.Sprintf("Brak nazwy kontrahenta w fakturze %s", invoice.Number))
		}
	}

	if len(invalidInvoices) == 0 {
		fmt.Println("Wszystkie faktury są prawidłowe!")
	} else {
		fmt.Printf("Znaleziono %d problemów:\n", len(invalidInvoices))
		for _, problem := range invalidInvoices {
			fmt.Printf("- %s\n", problem)
		}
	}
}

// Przykład 6: Analiza według typów faktur
func analyzeByType() {
	fmt.Println("\n=== Przykład 6: Analiza według typów ===")

	eppData, err := epp2json.ParseEPPFile("../eksport.epp", epp2json.DefaultParseOptions())
	if err != nil {
		log.Fatal("Błąd parsowania:", err)
	}

	// Filtruj faktury według typu
	fzInvoices := epp2json.FilterInvoices(eppData.Invoices, "FZ")
	fsInvoices := epp2json.FilterInvoices(eppData.Invoices, "FS")

	// Oblicz sumy
	fzNet, fzVat, fzGross := epp2json.CalculateTotalAmount(fzInvoices)
	fsNet, fsVat, fsGross := epp2json.CalculateTotalAmount(fsInvoices)

	fmt.Printf("Faktury zakupowe (FZ): %d faktur\n", len(fzInvoices))
	fmt.Printf("- Netto: %.2f PLN\n", fzNet)
	fmt.Printf("- VAT: %.2f PLN\n", fzVat)
	fmt.Printf("- Brutto: %.2f PLN\n", fzGross)

	fmt.Printf("\nFaktury sprzedażowe (FS): %d faktur\n", len(fsInvoices))
	fmt.Printf("- Netto: %.2f PLN\n", fsNet)
	fmt.Printf("- VAT: %.2f PLN\n", fsVat)
	fmt.Printf("- Brutto: %.2f PLN\n", fsGross)

	fmt.Printf("\nRazem: %.2f PLN brutto\n", fzGross+fsGross)
}

func main() {
	fmt.Println("Demo użycia biblioteki epp2json\n")

	// Sprawdź czy plik istnieje
	if _, err := os.Stat("../eksport.epp"); os.IsNotExist(err) {
		log.Fatal("Plik eksport.epp nie istnieje w katalogu nadrzędnym")
	}

	// Uruchom przykłady
	parseOnlyPurchaseInvoices()
	filterByContractor()
	analyzeByMonth()
	customJSONExport()
	validateData()
	analyzeByType()

	fmt.Println("\nDemo zakończone!")
}
