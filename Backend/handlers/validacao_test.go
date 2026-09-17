package handlers

import (
	"math"
	"testing"

	"controle-estoque/models"
)

func TestValidarItemRejeitaDadosInvalidos(t *testing.T) {
	tests := []struct {
		nome string
		item models.Item
	}{
		{nome: "quantidade negativa", item: models.Item{Nome: "Arroz", Quantidade: -1, Unidade: "kg", Local: "Despensa"}},
		{nome: "mínimo infinito", item: models.Item{Nome: "Arroz", Quantidade: 1, Unidade: "kg", Local: "Despensa", EstoqueMinimo: math.Inf(1)}},
		{nome: "local inválido", item: models.Item{Nome: "Arroz", Quantidade: 1, Unidade: "kg", Local: "Garagem"}},
		{nome: "nome vazio", item: models.Item{Quantidade: 1, Unidade: "kg", Local: "Despensa"}},
	}

	for _, teste := range tests {
		t.Run(teste.nome, func(t *testing.T) {
			if err := validarItem(teste.item); err == nil {
				t.Fatal("validarItem() aceitou dados inválidos")
			}
		})
	}
}

func TestValidarItemAceitaDadosValidos(t *testing.T) {
	item := models.Item{
		Nome:          "Arroz",
		Quantidade:    5,
		Unidade:       "kg",
		Local:         "Despensa",
		EstoqueMinimo: 1,
	}
	if err := validarItem(item); err != nil {
		t.Fatalf("validarItem() rejeitou item válido: %v", err)
	}
}
