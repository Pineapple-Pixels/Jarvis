package usecase

import (
	"testing"

	"asistente/pkg/domain"

	"github.com/stretchr/testify/assert"
)

const (
	testDate = "2026-03-10"
)

func TestFormatNumber_Small(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"zero", 0, "0"},
		{"hundreds", 500, "500"},
		{"thousands", 5000, "5,000"},
		{"tens_of_thousands", 15000, "15,000"},
		{"hundreds_of_thousands", 150000, "150,000"},
		{"millions", 1500000, "1,500,000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatNumber(tt.input)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatExpenseResponse_ARS(t *testing.T) {
	expense := domain.ParsedExpense{
		Amount:      5000,
		Category:    "Supermercado",
		Description: "Compras",
		PaidBy:      "Sebas",
		Date:        testDate,
	}

	result := FormatExpenseResponse(expense)

	expected := "\U0001F6D2 Anotado!\nSupermercado — Compras\nSebas pago $5,000 el 10/03/2026"
	assert.Equal(t, expected, result)
}

func TestFormatExpenseResponse_USD(t *testing.T) {
	expense := domain.ParsedExpense{
		Amount:      0,
		AmountUSD:   20,
		Category:    "Entretenimiento",
		Description: "Netflix",
		PaidBy:      "Sebas",
		Date:        testDate,
	}

	result := FormatExpenseResponse(expense)

	expected := "\U0001F3AE Anotado!\nEntretenimiento — Netflix\nSebas pago US$20 el 10/03/2026"
	assert.Equal(t, expected, result)
}

func TestFormatExpenseResponse_CategoryEmoji(t *testing.T) {
	tests := []struct {
		category string
		emoji    string
	}{
		{"Supermercado", "\U0001F6D2"},
		{"Restaurante", "\U0001F354"},
		{"Transporte", "\U0001F697"},
		{"Servicios", "\U0001F4F1"},
		{"Salud", "\U0001F48A"},
		{"Ropa", "\U0001F455"},
		{"Entretenimiento", "\U0001F3AE"},
		{"Educacion", "\U0001F4DA"},
		{"Hogar", "\U0001F3E0"},
		{"Otro", "\U0001F4E6"},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			expense := domain.ParsedExpense{
				Amount:   1000,
				Category: tt.category,
				PaidBy:   "Test",
				Date:     testDate,
			}

			result := FormatExpenseResponse(expense)

			expected := tt.emoji + " Anotado!\n" + tt.category + " — \nTest pago $1,000 el 10/03/2026"
			assert.Equal(t, expected, result)
		})
	}
}

func TestFormatExpenseResponse_UnknownCategory_DefaultEmoji(t *testing.T) {
	expense := domain.ParsedExpense{
		Amount:   1000,
		Category: "CategoriaInventada",
		PaidBy:   "Test",
		Date:     testDate,
	}

	result := FormatExpenseResponse(expense)

	expected := "\U0001F4E6 Anotado!\nCategoriaInventada — \nTest pago $1,000 el 10/03/2026"
	assert.Equal(t, expected, result)
}

func TestFormatExpenseResponse_DateFormat(t *testing.T) {
	expense := domain.ParsedExpense{
		Amount:   1000,
		Category: "Otro",
		PaidBy:   "Sebas",
		Date:     "2026-03-10",
	}

	result := FormatExpenseResponse(expense)

	expected := "\U0001F4E6 Anotado!\nOtro — \nSebas pago $1,000 el 10/03/2026"
	assert.Equal(t, expected, result)
}
