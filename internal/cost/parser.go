package cost

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type csvColumns struct {
	service  int
	start    int
	end      int
	cost     int
	currency int
}

var (
	serviceHeaders = map[string]bool{
		"lineitemproductcode": true,
		"product":             true,
		"productname":         true,
		"productservicecode":  true,
		"service":             true,
		"servicecode":         true,
		"servicename":         true,
	}
	startHeaders = map[string]bool{
		"billingperiodstartdate": true,
		"date":                   true,
		"lineitemusagestartdate": true,
		"month":                  true,
		"start":                  true,
		"startdate":              true,
		"usagestartdate":         true,
		"usagestarttime":         true,
	}
	endHeaders = map[string]bool{
		"billingperiodenddate": true,
		"end":                  true,
		"enddate":              true,
		"lineitemusageenddate": true,
		"usageenddate":         true,
		"usageendtime":         true,
	}
	costHeaders = map[string]bool{
		"amount":                      true,
		"amortizedcost":               true,
		"blendedcost":                 true,
		"cost":                        true,
		"costusd":                     true,
		"lineitemblendedcost":         true,
		"lineitemnetunblendedcost":    true,
		"lineitemunblendedcost":       true,
		"netamortizedcost":            true,
		"netunblendedcost":            true,
		"reservationamortizedupfront": true,
		"totalcost":                   true,
		"unblendedcost":               true,
	}
	currencyHeaders = map[string]bool{
		"currency":                true,
		"currencycode":            true,
		"lineitemcurrencycode":    true,
		"pricingcurrency":         true,
		"pricingcurrencycode":     true,
		"pricingtermcurrencycode": true,
	}
)

func ParseCostExplorerCSV(r io.Reader) ([]Record, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("cost CSV is empty")
		}
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	columns, err := detectColumns(headers)
	if err != nil {
		return nil, err
	}

	var records []Record
	line := 1
	for {
		row, err := reader.Read()
		line++
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read CSV line %d: %w", line, err)
		}
		if isBlankRow(row) {
			continue
		}

		record, err := parseRecord(row, columns)
		if err != nil {
			return nil, fmt.Errorf("parse CSV line %d: %w", line, err)
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		return nil, errors.New("cost CSV contains no data rows")
	}

	return records, nil
}

func detectColumns(headers []string) (csvColumns, error) {
	columns := csvColumns{
		service:  -1,
		start:    -1,
		end:      -1,
		cost:     -1,
		currency: -1,
	}

	for index, header := range headers {
		normalized := normalizeHeader(header)
		switch {
		case serviceHeaders[normalized] && columns.service == -1:
			columns.service = index
		case startHeaders[normalized] && columns.start == -1:
			columns.start = index
		case endHeaders[normalized] && columns.end == -1:
			columns.end = index
		case costHeaders[normalized] && columns.cost == -1:
			columns.cost = index
		case currencyHeaders[normalized] && columns.currency == -1:
			columns.currency = index
		}
	}

	var missing []string
	if columns.service == -1 {
		missing = append(missing, "service")
	}
	if columns.start == -1 {
		missing = append(missing, "start date")
	}
	if columns.cost == -1 {
		missing = append(missing, "cost")
	}
	if len(missing) > 0 {
		return columns, fmt.Errorf("missing required column(s): %s", strings.Join(missing, ", "))
	}

	return columns, nil
}

func parseRecord(row []string, columns csvColumns) (Record, error) {
	service := strings.TrimSpace(cell(row, columns.service))
	if service == "" {
		return Record{}, errors.New("service is empty")
	}

	startDate, err := parseDate(cell(row, columns.start))
	if err != nil {
		return Record{}, fmt.Errorf("invalid start date: %w", err)
	}

	var endDate time.Time
	if columns.end != -1 && strings.TrimSpace(cell(row, columns.end)) != "" {
		endDate, err = parseDate(cell(row, columns.end))
		if err != nil {
			return Record{}, fmt.Errorf("invalid end date: %w", err)
		}
	}

	cost, err := parseCostValue(cell(row, columns.cost))
	if err != nil {
		return Record{}, fmt.Errorf("invalid cost: %w", err)
	}

	currency := "USD"
	if columns.currency != -1 && strings.TrimSpace(cell(row, columns.currency)) != "" {
		currency = strings.ToUpper(strings.TrimSpace(cell(row, columns.currency)))
	}

	return Record{
		Service:   service,
		StartDate: startDate,
		EndDate:   endDate,
		Cost:      cost,
		Currency:  currency,
	}, nil
}

func parseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("date is empty")
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02",
		"2006/01/02 15:04:05",
		"1/2/2006",
		"1/2/2006 15:04",
		"01/02/2006",
		"2006-01",
		"Jan 2006",
		"January 2006",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date %q", value)
}

func parseCostValue(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, errors.New("cost is empty")
	}

	negative := strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")")
	value = strings.Trim(value, "()")
	value = strings.ReplaceAll(value, ",", "")
	value = strings.ReplaceAll(value, "$", "")
	value = strings.ReplaceAll(value, "€", "")
	value = strings.ReplaceAll(value, "£", "")
	value = strings.TrimSpace(value)

	fields := strings.Fields(value)
	if len(fields) == 2 {
		switch {
		case strings.EqualFold(fields[0], "USD"):
			value = fields[1]
		case strings.EqualFold(fields[1], "USD"):
			value = fields[0]
		}
	}

	amount, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	if negative {
		amount = -amount
	}

	return amount, nil
}

func normalizeHeader(value string) string {
	var builder strings.Builder
	for _, char := range strings.ToLower(value) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func cell(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}

func isBlankRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
