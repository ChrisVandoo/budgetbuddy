package parse

import (
	"encoding/csv"
	"fmt"
	"io"
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
	HeaderRow []string
	Records   [][]string
}

func NewParser() *Parser {
	return &Parser{}
}

// Reads a CSV file, finds the header row and records.
func (p *Parser) ReadCSVFile(r io.Reader) error {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true
	// CSV file may have "bad" rows at the beginning of the file which we should ignore
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read csv: %w", err)
	}

	var headerIndex int
	for i, record := range records {
		// Find the first row that has at least 3 columns as the header row, we require at least a transaction date, description, and amount.
		// This makes a few assumptions, but should help filter out invalid lines.
		if len(record) > 2 {
			p.HeaderRow = record
			headerIndex = i
			break
		}
	}

	if p.HeaderRow == nil {
		return fmt.Errorf("failed to find headers in csv file, require at least date, description, and amount")
	}

	p.Records = records[headerIndex+1:]

	return nil
}

// Returns the header record/row so that we can find the correct source mapping
func (p *Parser) GetHeaderRecord() ([]string, error) {
	if p.HeaderRow == nil {
		return nil, fmt.Errorf("no header record found, read the CSV file before trying to get the header record")
	}
	return p.HeaderRow, nil
}

// Normalizes a list of CSV records based on a source mapping that maps a bank specific CSV format to a ParsedTransaction
func (p *Parser) ParseRecords(mapping types.SourceMapping, sourceName string) ([]ParsedTransaction, error) {
	if p.Records == nil || p.HeaderRow == nil {
		return nil, fmt.Errorf("no records or header row found, read the CSV file before trying to parse the records")
	}

	if len(p.Records) < 1 {
		return nil, fmt.Errorf("csv must have at one data row")
	}

	dateIdx := findHeaderIndex(p.HeaderRow, mapping.Date.Header)
	descIdx := findHeaderIndex(p.HeaderRow, mapping.Description.Header)
	if dateIdx == -1 || descIdx == -1 {
		return nil, fmt.Errorf("required headers not found in CSV")
	}

	var amountIdx int
	var inIdx, outIdx int
	if mapping.Amount.SingleColumn {
		amountIdx = findHeaderIndex(p.HeaderRow, mapping.Amount.HeaderOut)
		if amountIdx == -1 {
			return nil, fmt.Errorf("amount header %q not found", mapping.Amount.HeaderOut)
		}
	} else {
		inIdx = findHeaderIndex(p.HeaderRow, mapping.Amount.HeaderIn)
		outIdx = findHeaderIndex(p.HeaderRow, mapping.Amount.HeaderOut)
		if inIdx == -1 || outIdx == -1 {
			return nil, fmt.Errorf("amount headers not found (in: %q, out: %q)",
				mapping.Amount.HeaderIn, mapping.Amount.HeaderOut)
		}
	}

	var transactions []ParsedTransaction
	expectedFieldsPerRecord := len(p.HeaderRow)
	for _, row := range p.Records {
		if len(row) < expectedFieldsPerRecord {
			continue
		}

		dateVal := strings.TrimSpace(row[dateIdx])
		descVal := strings.TrimSpace(row[descIdx])
		// We may encounter multiple header rows in a single CSV file. If that is the case just skip to the next row.
		if dateVal == "" || descVal == "" || (dateVal == mapping.Date.Header && descVal == mapping.Description.Header) {
			continue
		}

		var amountCents int64
		if mapping.Amount.SingleColumn {
			amt, err := NormalizeAmount(row[amountIdx], mapping.Amount)
			if err != nil {
				return nil, fmt.Errorf("failed to parse amount on row %q: %w", descVal, err)
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
			Source:      sourceName,
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
