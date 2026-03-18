package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectIntent_Expense(t *testing.T) {
	tests := []struct {
		msg string
	}{
		{"gasté 500 en el super"},
		{"Gasté 1000 en nafta"},
		{"gaste 200 en farmacia"},
		{"pagué 3000 de luz"},
		{"compré ropa por 5 lucas"},
		{"cena con amigos 2000 pesos"},
		{"cambié 100 dólares"},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			assert.Equal(t, intentExpense, detectIntent(tt.msg))
		})
	}
}

func TestDetectIntent_Note(t *testing.T) {
	tests := []struct {
		msg string
	}{
		{"nota comprar leche"},
		{"nota: reunión el martes"},
		{"recordá llamar al médico"},
		{"recordame pagar la tarjeta"},
		{"acordate de comprar pan"},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			assert.Equal(t, intentNote, detectIntent(tt.msg))
		})
	}
}

func TestDetectIntent_Chat(t *testing.T) {
	tests := []struct {
		msg string
	}{
		{"hola cómo estás"},
		{"qué hora es"},
		{"contame un chiste"},
		{"explicame qué es kubernetes"},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			assert.Equal(t, intentChat, detectIntent(tt.msg))
		})
	}
}

func TestStripNotePrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"nota comprar leche", "comprar leche"},
		{"Nota: reunión martes", "reunión martes"},
		{"recordá llamar al médico", "llamar al médico"},
		{"recordame pagar tarjeta", "pagar tarjeta"},
		{"hola mundo", "hola mundo"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, stripNotePrefix(tt.input))
		})
	}
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
	assert.Equal(t, "hel...", truncate("hello world", 3))
}
