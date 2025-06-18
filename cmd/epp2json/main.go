package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/janexpl/epp2json"
)

func main() {
	var inputFile, outputFile string
	var onlyFZ, onlyFS bool

	flag.StringVar(&inputFile, "input", "eksport.epp", "Ścieżka do pliku wejściowego")
	flag.StringVar(&outputFile, "output", "faktury.json", "Ścieżka do pliku wyjściowego")
	flag.BoolVar(&onlyFZ, "fz-only", false, "Parsuj tylko faktury zakupowe (FZ)")
	flag.BoolVar(&onlyFS, "fs-only", false, "Parsuj tylko faktury sprzedażowe (FS)")
	flag.Parse()

	// Ustaw opcje parsowania
	options := epp2json.DefaultParseOptions()
	if onlyFZ {
		options.IncludeFS = false
	}
	if onlyFS {
		options.IncludeFZ = false
	}

	// Konwertuj plik
	jsonData, err := epp2json.ConvertEPPToJSON(inputFile, options)
	if err != nil {
		log.Fatal("Błąd konwersji:", err)
	}
	epp2json.WriteJSONToFile(jsonData, outputFile)

	// Pobierz statystyki
	eppData, err := epp2json.ParseEPPFile(inputFile, options)
	if err != nil {
		log.Fatal("Błąd podczas odczytu pliku dla statystyk:", err)
	}

	fzCount, fsCount := epp2json.GetInvoiceStats(eppData)
	totalCount := len(eppData.Invoices)

	fmt.Printf("Konwersja zakończona pomyślnie!\n")
	fmt.Printf("Przetworzono %d faktur\n", totalCount)
	fmt.Printf("Wynik zapisano do pliku: %s\n", outputFile)
	fmt.Printf("Faktury zakupowe (FZ): %d\n", fzCount)
	fmt.Printf("Faktury sprzedażowe (FS): %d\n", fsCount)
}
