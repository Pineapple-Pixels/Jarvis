package clients

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetsClient struct {
	service       *sheets.Service
	spreadsheetID string
}

func NewSheetsClient(credentialsFile, spreadsheetID string) (*SheetsClient, error) {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("sheets: create service: %w", err)
	}

	return &SheetsClient{
		service:       srv,
		spreadsheetID: spreadsheetID,
	}, nil
}

// AppendRow appends a row of values to the specified sheet.
func (c *SheetsClient) AppendRow(sheetName string, values []any) error {
	rangeStr := sheetName + "!A:Z"

	row := &sheets.ValueRange{
		Values: [][]any{values},
	}

	_, err := c.service.Spreadsheets.Values.Append(
		c.spreadsheetID,
		rangeStr,
		row,
	).ValueInputOption("USER_ENTERED").Do()

	if err != nil {
		return fmt.Errorf("sheets: append row: %w", err)
	}

	return nil
}

// ReadRange reads values from the specified range.
func (c *SheetsClient) ReadRange(sheetRange string) ([][]any, error) {
	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, sheetRange).Do()
	if err != nil {
		return nil, fmt.Errorf("sheets: read range: %w", err)
	}

	return resp.Values, nil
}
