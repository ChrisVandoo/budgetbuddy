package parse

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

type ParsedTransaction struct {
	Source      string
	Date        string
	Description string
	AmountCents int64
}

type Parser struct {
	SourceName string
	Mapping    types.SourceMapping
}

func NewParser(sourceName string, mapping types.SourceMapping) *Parser {
	return &Parser{
		SourceName: sourceName,
		Mapping:    mapping,
	}
}

func (p *Parser) ParseFile(path string) ([]ParsedTransaction, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	return p.Parse(f)
}

func (p *Parser) Parse(r io.Reader) ([]ParsedTransaction, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("csv must have at least a header row and one data row")
	}

	headers := records[0]
	dataRows := records[1:]

	dateIdx := findHeaderIndex(headers, p.Mapping.Date.Header)
	descIdx := findHeaderIndex(headers, p.Mapping.Description.Header)
	if dateIdx == -1 || descIdx == -1 {
		return nil, fmt.Errorf("required headers not found in CSV")
	}

	var amountIdx int
	var inIdx, outIdx int
	if p.Mapping.Amount.SingleColumn {
		amountIdx = findHeaderIndex(headers, p.Mapping.Amount.HeaderOut)
		if amountIdx == -1 {
			return nil, fmt.Errorf("amount header %q not found", p.Mapping.Amount.HeaderOut)
		}
	} else {
		inIdx = findHeaderIndex(headers, p.Mapping.Amount.HeaderIn)
		outIdx = findHeaderIndex(headers, p.Mapping.Amount.HeaderOut)
		if inIdx == -1 || outIdx == -1 {
			return nil, fmt.Errorf("amount headers not found (in: %q, out: %q)",
				p.Mapping.Amount.HeaderIn, p.Mapping.Amount.HeaderOut)
		}
	}

	var transactions []ParsedTransaction
	for _, row := range dataRows {
		if len(row) <= dateIdx || len(row) <= descIdx {
			continue
		}
		if !p.Mapping.Amount.SingleColumn && (len(row) <= inIdx || len(row) <= outIdx) {
			continue
		}

		dateVal := strings.TrimSpace(row[dateIdx])
		descVal := strings.TrimSpace(row[descIdx])
		if dateVal == "" || descVal == "" {
			continue
		}

		var amountCents int64
		if p.Mapping.Amount.SingleColumn {
			amt, err := NormalizeAmount(row[amountIdx], p.Mapping.Amount)
			if err != nil {
				return nil, fmt.Errorf("parse amount on row %q: %w", descVal, err)
			}
			amountCents = amt
		} else {
			inVal, err := ParseCents(row[inIdx])
			if err != nil {
				return nil, fmt.Errorf("parse money in on row %q: %w", descVal, err)
			}
			outVal, err := ParseCents(row[outIdx])
			if err != nil {
				return nil, fmt.Errorf("parse money out on row %q: %w", descVal, err)
			}
			amountCents = inVal - outVal
		}

		transactions = append(transactions, ParsedTransaction{
			Source:      p.SourceName,
			Date:        dateVal,
			Description: descVal,
			AmountCents: amountCents,
		})
	}

	return transactions, nil
}

func findHeaderIndex(headers []string, target string) int {
	target = strings.TrimSpace(strings.ToLower(target))
	for i, h := range headers {
		if strings.TrimSpace(strings.ToLower(h)) == target {
			return i
		}
	}
	return -1
}

func ReadCSVHeaders(r io.Reader) ([]string, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv headers: %w", err)
	}

	return headers, nil
}
