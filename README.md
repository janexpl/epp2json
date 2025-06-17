# Biblioteka epp2json

**Biblioteka Go do konwersji plików EPP (Export Plus) do formatu JSON**

Program został opracowany na podstawie specyfikacji **Komunikacja EDI++ wersja 1.11** i służy do konwersji plików eksportowych z systemu Subiekt GT do formatu JSON. Może być używany zarówno jako aplikacja CLI, jak i jako biblioteka w innych projektach Go.

## 📦 Instalacja

### Jako aplikacja CLI

```bash
go install github.com/janexpl/epp2json/cmd/epp2json@latest
```

### Jako biblioteka w projekcie Go

```bash
go get github.com/janexpl/epp2json
```

## 🚀 Użycie

### Aplikacja CLI

```bash
# Podstawowe użycie - konwersja wszystkich faktur
epp2json -input eksport.epp -output faktury.json

# Tylko faktury zakupowe (FZ)
epp2json -input eksport.epp -output faktury_fz.json -fz-only

# Tylko faktury sprzedażowe (FS)
epp2json -input eksport.epp -output faktury_fs.json -fs-only

# Pomoc
epp2json -help
```

### Biblioteka w kodzie Go

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/janexpl/epp2json"
)

func main() {
    // Parsowanie wszystkich faktur
    eppData, err := epp2json.ParseEPPFile("eksport.epp", epp2json.DefaultParseOptions())
    if err != nil {
        log.Fatal(err)
    }
    
    // Statystyki
    fzCount, fsCount := epp2json.GetInvoiceStats(eppData)
    fmt.Printf("FZ: %d, FS: %d\n", fzCount, fsCount)
    
    // Filtrowanie według kontrahenta
    pkoInvoices := epp2json.GetInvoicesByContractor(eppData.Invoices, "PKO")
    
    // Grupowanie według miesięcy
    monthlyGroups := epp2json.GroupInvoicesByMonth(eppData.Invoices)
    
    // Obliczanie sum
    net, vat, gross := epp2json.CalculateTotalAmount(eppData.Invoices)
    fmt.Printf("Łączna kwota: %.2f PLN\n", gross)
}
```

## 📊 API Biblioteki

### Główne typy danych

```go
type Invoice struct {
    Type               string        `json:"typ"`
    Number             string        `json:"numer"`
    ContractorName     string        `json:"nazwa_kontrahenta"`
    NetAmount          float64       `json:"kwota_netto"`
    VatAmount          float64       `json:"kwota_vat"`
    GrossAmount        float64       `json:"kwota_brutto"`
    IssueDate          time.Time     `json:"data_wystawienia"`
    Items              []InvoiceItem `json:"pozycje"`
    // ... więcej pól
}

type ParseOptions struct {
    IncludeFZ bool // Faktury zakupowe
    IncludeFS bool // Faktury sprzedażowe
}
```

### Główne funkcje

#### Parsowanie

- `ParseEPPFile(filename, options)` - parsuje plik EPP
- `ParseEPPFromString(content, options)` - parsuje zawartość EPP ze stringa
- `ConvertEPPToJSON(inputFile, outputFile, options)` - konwertuje plik EPP do JSON

#### Filtrowanie i analiza

- `FilterInvoices(invoices, type)` - filtruje faktury według typu
- `GetInvoicesByContractor(invoices, contractor)` - faktury według kontrahenta
- `GetInvoicesByDateRange(invoices, start, end)` - faktury z zakresu dat
- `GroupInvoicesByMonth(invoices)` - grupuje faktury według miesięcy

#### Obliczenia

- `CalculateTotalAmount(invoices)` - oblicza sumy netto, VAT, brutto
- `GetInvoiceStats(eppData)` - statystyki liczby faktur FZ/FS

#### Opcje

- `DefaultParseOptions()` - domyślne opcje (wszystkie faktury)

## 🏗️ Struktura projektu

```
epp2json/
├── epp2json.go          # Główna biblioteka
├── cmd/
│   └── epp2json/
│       └── main.go      # Aplikacja CLI
├── examples/
│   └── example_usage.go # Przykłady użycia
├── go.mod
├── go.sum
└── README.md
```

## 🔧 Funkcjonalności

### ✅ Obsługa formatów
- **Kodowanie Windows-1250** - poprawne odczytywanie polskich znaków
- **Format CSV z cudzysłowami** - prawidłowe parsowanie pól zawierających przecinki
- **Daty w formacie YYYYMMDDHHMMSS** - automatyczna konwersja do `time.Time`

### ✅ Typy faktur
- **FZ** - Faktury zakupowe
- **FS** - Faktury sprzedażowe

### ✅ Struktura danych
- **Nagłówki faktur** - wszystkie podstawowe informacje o fakturze
- **Pozycje faktur** - szczegółowe informacje o pozycjach
- **Dane kontrahentów** - pełne informacje adresowe i NIP
- **Kwoty** - netto, VAT, brutto z automatycznym parsowaniem

### ✅ Funkcje analityczne
- Statystyki liczby faktur
- Filtrowanie według różnych kryteriów
- Grupowanie czasowe
- Obliczenia sum

## 📋 Przykłady użycia

Sprawdź folder `examples/` gdzie znajdziesz kompletne przykłady demonstrujące:
- Parsowanie tylko określonych typów faktur
- Filtrowanie według kontrahentów
- Analizę według miesięcy
- Niestandardowy eksport JSON
- Walidację danych
- Analizę według typów faktur

```bash
cd examples && go run example_usage.go
```

## 🔍 Wymagania

- **Go 1.21+**
- **golang.org/x/text** - obsługa kodowania Windows-1250

## 📄 Licencja

Program na licencji APACHE 2.0

## 🐛 Zgłaszanie błędów

W przypadku problemów lub propozycji ulepszeń, proszę o utworzenie issue w repozytorium projektu.

---

**Program został przetestowany na rzeczywistych plikach EPP z systemu Subiekt GT.** 