package parse_test

import (
	"strings"
	"testing"

	"github.com/ChrisVandoo/budgetbuddy/internal/parse"
	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

func TestParseCents_Simple(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"50.00", 5000},
		{"0.00", 0},
		{"100", 10000},
		{"99.99", 9999},
		{"0.01", 1},
		{"-50.00", -5000},
		{"-100", -10000},
		{"$50.00", 5000},
		{"-$50.00", -5000},
		{"1,234.56", 123456},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parse.ParseCents(tt.input)
			if err != nil {
				t.Fatalf("ParseCents(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseCents(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseCents_Invalid(t *testing.T) {
	invalid := []string{"abc", "10.00.00", ""}
	for _, s := range invalid {
		got, err := parse.ParseCents(s)
		if s == "" {
			if err != nil {
				t.Fatalf("ParseCents(%q) should not error for empty", s)
			}
			if got != 0 {
				t.Errorf("ParseCents(%q) = %d, want 0", s, got)
			}
			continue
		}
		if err == nil {
			t.Errorf("ParseCents(%q) expected error, got %d", s, got)
		}
	}
}

func TestNormalizeAmount_SingleColumnPositive(t *testing.T) {
	mapping := types.AmountMapping{
		SingleColumn:      true,
		IsPositiveMoneyIn: true,
	}
	val, err := parse.NormalizeAmount("50.00", mapping)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 5000 {
		t.Errorf("expected 5000, got %d", val)
	}
}

func TestNormalizeAmount_SingleColumnNegative(t *testing.T) {
	mapping := types.AmountMapping{
		SingleColumn:      true,
		IsPositiveMoneyIn: false,
	}
	val, err := parse.NormalizeAmount("50.00", mapping)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != -5000 {
		t.Errorf("expected -5000, got %d", val)
	}
}

func TestNormalizeAmount_NegativeCSVValue(t *testing.T) {
	mapping := types.AmountMapping{
		SingleColumn:      true,
		IsPositiveMoneyIn: false,
	}
	val, err := parse.NormalizeAmount("-50.00", mapping)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 5000 {
		t.Errorf("expected 5000, got %d", val)
	}
}

func TestDetectSource(t *testing.T) {
	sources := &types.SourcesYAML{
		Sources: map[string]types.SourceConfig{
			"Date,Transaction Details,Funds Out,Funds In": {
				Name: "Simplii",
			},
			"Transaction Date,Description,Amount": {
				Name: "BMO",
			},
		},
	}

	key, config, found := parse.DetectSource([]string{"Date", "Transaction Details", "Funds Out", "Funds In"}, sources)
	if !found {
		t.Fatal("expected to detect source")
	}
	if config.Name != "Simplii" {
		t.Errorf("expected Simplii, got %s", config.Name)
	}
	_ = key

	_, _, found = parse.DetectSource([]string{"Unknown", "Headers"}, sources)
	if found {
		t.Fatal("expected not to detect unknown source")
	}
}

func TestDetectSourceCaseInsensitive(t *testing.T) {
	sources := &types.SourcesYAML{
		Sources: map[string]types.SourceConfig{
			"Date,Transaction Details,Funds Out,Funds In": {
				Name: "Simplii",
			},
		},
	}

	_, _, found := parse.DetectSource([]string{"date", "transaction details", "funds out", "funds in"}, sources)
	if !found {
		t.Fatal("expected case-insensitive match")
	}
}

func TestParseCSV_SingleColumn(t *testing.T) {
	mapping := types.SourceMapping{
		Date:        types.DateMapping{Header: "Transaction Date", Format: "2006-01-02"},
		Description: types.DescriptionMapping{Header: "Description"},
		Amount: types.AmountMapping{
			SingleColumn:      true,
			IsPositiveMoneyIn: false,
			HeaderOut:         "Amount",
			HeaderIn:          "Amount",
		},
	}

	parser := &parse.Parser{
		HeaderRow: []string{"Transaction Date", "Description", "Amount"},
		Records: [][]string{
			{"2026-01-15", "AMAZON.CA", "50.00"},
			{"2026-01-16", "WALMART", "25.50"},
		},
	}
	txns, err := parser.ParseRecords(mapping, "Test Source")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(txns) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txns))
	}
	if txns[0].AmountCents != -5000 {
		t.Errorf("expected -5000, got %d", txns[0].AmountCents)
	}
	if txns[1].AmountCents != -2550 {
		t.Errorf("expected -2550, got %d", txns[1].AmountCents)
	}
	if txns[0].Description != "AMAZON.CA" {
		t.Errorf("expected AMAZON.CA, got %s", txns[0].Description)
	}
	if txns[0].Source != "Test Source" {
		t.Errorf("expected Test Source, got %s", txns[0].Source)
	}
}

func TestParseCSV_DualColumn(t *testing.T) {
	mapping := types.SourceMapping{
		Date:        types.DateMapping{Header: "Date"},
		Description: types.DescriptionMapping{Header: "Transaction Details"},
		Amount: types.AmountMapping{
			SingleColumn:      false,
			IsPositiveMoneyIn: true,
			HeaderOut:         "Funds Out",
			HeaderIn:          "Funds In",
		},
	}

	parser := &parse.Parser{
		HeaderRow: []string{"Date", "Transaction Details", "Funds Out", "Funds In"},
		Records: [][]string{
			{"01/15/2026", "AMAZON.CA", "50.00", ""},
			{"01/16/2026", "PAYCHECK", "", "1000.00"},
		},
	}
	txns, err := parser.ParseRecords(mapping, "Simplii")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(txns) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txns))
	}
	if txns[0].AmountCents != -5000 {
		t.Errorf("expected -5000, got %d", txns[0].AmountCents)
	}
	if txns[1].AmountCents != 100000 {
		t.Errorf("expected 100000, got %d", txns[1].AmountCents)
	}
}

func TestParseCSV_EmptyRows(t *testing.T) {
	mapping := types.SourceMapping{
		Date:        types.DateMapping{Header: "Date"},
		Description: types.DescriptionMapping{Header: "Desc"},
		Amount: types.AmountMapping{
			SingleColumn:      true,
			IsPositiveMoneyIn: true,
			HeaderOut:         "Amount",
			HeaderIn:          "Amount",
		},
	}

	parser := &parse.Parser{
		HeaderRow: []string{"Date", "Amount", "Desc"},
		Records: [][]string{
			{"2026-01-15", "50.00", "Test1"},
			{"", "", ""},
		},
	}
	txns, err := parser.ParseRecords(mapping, "Test")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("expected 1 transaction (empty rows skipped), got %d", len(txns))
	}
}

func TestParseCSV_OnlyHeaders(t *testing.T) {
	mapping := types.SourceMapping{
		Date:        types.DateMapping{Header: "Date"},
		Description: types.DescriptionMapping{Header: "Desc"},
		Amount: types.AmountMapping{
			SingleColumn:      true,
			IsPositiveMoneyIn: true,
			HeaderOut:         "Amount",
			HeaderIn:          "Amount",
		},
	}

	parser := &parse.Parser{
		HeaderRow: []string{"Date", "Amount", "Desc"},
		Records:   [][]string{},
	}
	_, err := parser.ParseRecords(mapping, "Test")
	if err == nil {
		t.Fatal("expected error for csv with only headers")
	}
}

func TestParseCSV_MissingHeaders(t *testing.T) {
	mapping := types.SourceMapping{
		Date:        types.DateMapping{Header: "Date"},
		Description: types.DescriptionMapping{Header: "Missing"},
		Amount: types.AmountMapping{
			SingleColumn:      true,
			IsPositiveMoneyIn: true,
			HeaderOut:         "Amount",
			HeaderIn:          "Amount",
		},
	}

	parser := &parse.Parser{
		HeaderRow: []string{"Date", "Amount", "Desc"},
		Records: [][]string{
			{"2026-01-15", "50.00", "Test"},
		},
	}
	_, err := parser.ParseRecords(mapping, "Test")
	if err == nil {
		t.Fatal("expected error for missing headers")
	}
}

func TestReadCSVHeaders(t *testing.T) {
	csv := `Date,Amount,Description
2026-01-15,50.00,Test`

	parser := parse.NewParser()
	if err := parser.ReadCSVFile(strings.NewReader(csv)); err != nil {
		t.Fatalf("ReadCSVFile failed: %v", err)
	}
	headers, err := parser.GetHeaderRecord()
	if err != nil {
		t.Fatalf("GetHeaderRecord failed: %v", err)
	}
	if len(headers) != 3 {
		t.Fatalf("expected 3 headers, got %d", len(headers))
	}
	if headers[0] != "Date" {
		t.Errorf("expected Date, got %s", headers[0])
	}
}

func TestReadCSVFile_ExtraLinesBeforeHeader(t *testing.T) {
	csv := `Account Statement
Period: January 2026
Date,Amount,Desc
2026-01-15,50.00,Test1`

	parser := parse.NewParser()
	if err := parser.ReadCSVFile(strings.NewReader(csv)); err != nil {
		t.Fatalf("ReadCSVFile failed: %v", err)
	}

	headers, err := parser.GetHeaderRecord()
	if err != nil {
		t.Fatalf("GetHeaderRecord failed: %v", err)
	}
	if len(headers) != 3 {
		t.Fatalf("expected 3 headers, got %d", len(headers))
	}
	if headers[0] != "Date" {
		t.Errorf("expected Date, got %s", headers[0])
	}
	if headers[1] != "Amount" {
		t.Errorf("expected Amount, got %s", headers[1])
	}
	if headers[2] != "Desc" {
		t.Errorf("expected Desc, got %s", headers[2])
	}

	if len(parser.Records) != 1 {
		t.Fatalf("expected 1 data record, got %d", len(parser.Records))
	}
	if parser.Records[0][0] != "2026-01-15" {
		t.Errorf("expected 2026-01-15, got %s", parser.Records[0][0])
	}
	if parser.Records[0][1] != "50.00" {
		t.Errorf("expected 50.00, got %s", parser.Records[0][1])
	}
}

func TestReadCSVFile_Normal(t *testing.T) {
	csv := `Date,Amount,Desc
2026-01-15,50.00,Test1
2026-01-16,25.50,Test2`

	parser := parse.NewParser()
	if err := parser.ReadCSVFile(strings.NewReader(csv)); err != nil {
		t.Fatalf("ReadCSVFile failed: %v", err)
	}

	headers, err := parser.GetHeaderRecord()
	if err != nil {
		t.Fatalf("GetHeaderRecord failed: %v", err)
	}
	if len(headers) != 3 {
		t.Fatalf("expected 3 headers, got %d", len(headers))
	}

	if len(parser.Records) != 2 {
		t.Fatalf("expected 2 data records, got %d", len(parser.Records))
	}
	if parser.Records[0][0] != "2026-01-15" {
		t.Errorf("expected 2026-01-15, got %s", parser.Records[0][0])
	}
	if parser.Records[1][2] != "Test2" {
		t.Errorf("expected Test2, got %s", parser.Records[1][2])
	}
}

func TestReadCSVFile_OnlyHeader(t *testing.T) {
	csv := `Date,Amount,Desc`

	parser := parse.NewParser()
	if err := parser.ReadCSVFile(strings.NewReader(csv)); err != nil {
		t.Fatalf("ReadCSVFile failed: %v", err)
	}

	headers, err := parser.GetHeaderRecord()
	if err != nil {
		t.Fatalf("GetHeaderRecord failed: %v", err)
	}
	if len(headers) != 3 {
		t.Fatalf("expected 3 headers, got %d", len(headers))
	}

	if len(parser.Records) != 0 {
		t.Fatalf("expected 0 records (header excluded), got %d", len(parser.Records))
	}
}

func TestReadCSVFile_NoValidHeader(t *testing.T) {
	csv := `one,two
foo,bar`

	parser := parse.NewParser()
	err := parser.ReadCSVFile(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for CSV without a valid header row")
	}
}

func TestReadCSVFile_Empty(t *testing.T) {
	csv := ``

	parser := parse.NewParser()
	err := parser.ReadCSVFile(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for empty CSV")
	}
}

func TestLoadSources_NonExistent(t *testing.T) {
	sources, err := parse.LoadSources("/tmp/nonexistent/sources.yaml")
	if err != nil {
		t.Fatalf("LoadSources should not error for missing file: %v", err)
	}
	if sources == nil || sources.Sources == nil {
		t.Fatal("expected non-nil sources")
	}
	if len(sources.Sources) != 0 {
		t.Errorf("expected 0 sources, got %d", len(sources.Sources))
	}
}
