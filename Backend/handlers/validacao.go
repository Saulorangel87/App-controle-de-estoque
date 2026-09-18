package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"
	"unicode/utf8"

	"controle-estoque/models"
)

// decodificarJSON aceita exatamente um valor JSON e rejeita conteúdo extra.
// Com o middleware LimitarCorpoEstruturado, a segunda leitura também detecta
// corpos que começam com um objeto válido e continuam além do limite.
func decodificarJSON(r *http.Request, destino any) error {
	decodificador := json.NewDecoder(r.Body)
	if err := decodificador.Decode(destino); err != nil {
		return err
	}

	var extra any
	err := decodificador.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return errors.New("corpo JSON contém mais de um valor")
	}
	return err
}

const (
	tamanhoMaximoNome    = 200
	tamanhoMaximoUnidade = 30
	valorMaximoEstoque   = 1_000_000_000
)

var locaisPermitidos = map[string]struct{}{
	"Despensa":  {},
	"Geladeira": {},
	"Freezer":   {},
	"Armário":   {},
}

func validarNumeroEstoque(valor float64, campo string) error {
	if math.IsNaN(valor) || math.IsInf(valor, 0) {
		return errors.New(campo + " deve ser um número válido")
	}
	if valor < 0 {
		return errors.New(campo + " não pode ser negativo")
	}
	if valor > valorMaximoEstoque {
		return errors.New(campo + " excede o limite permitido")
	}
	return nil
}

func validarItem(it models.Item) error {
	it.Nome = strings.TrimSpace(it.Nome)
	it.Unidade = strings.TrimSpace(it.Unidade)
	if it.Nome == "" || it.Unidade == "" || it.Local == "" {
		return errors.New("nome, unidade e local são obrigatórios")
	}
	if utf8.RuneCountInString(it.Nome) > tamanhoMaximoNome {
		return errors.New("nome do item excede o limite permitido")
	}
	if utf8.RuneCountInString(it.Unidade) > tamanhoMaximoUnidade {
		return errors.New("unidade excede o limite permitido")
	}
	if _, ok := locaisPermitidos[it.Local]; !ok {
		return errors.New("localização inválida")
	}
	if err := validarNumeroEstoque(it.Quantidade, "quantidade"); err != nil {
		return err
	}
	return validarNumeroEstoque(it.EstoqueMinimo, "estoque mínimo")
}

func validarEntradaConfirmada(entrada models.EntradaConfirmada) error {
	if err := validarNumeroEstoque(entrada.Quantidade, "quantidade"); err != nil {
		return err
	}
	if entrada.Quantidade <= 0 {
		return errors.New("quantidade deve ser maior que zero")
	}

	if entrada.ItemID != nil {
		if *entrada.ItemID <= 0 {
			return errors.New("item inválido")
		}
		return nil
	}

	return validarItem(models.Item{
		Nome:          entrada.Nome,
		Quantidade:    entrada.Quantidade,
		Unidade:       entrada.Unidade,
		Local:         entrada.Local,
		EstoqueMinimo: entrada.EstoqueMinimo,
	})
}
