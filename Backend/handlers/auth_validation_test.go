package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidarSenhaExigeMinimoApenasParaNovasCredenciais(t *testing.T) {
	if err := validarSenha("1234567"); err == nil {
		t.Fatal("senha com 7 caracteres foi aceita")
	}
	if err := validarSenha("12345678"); err != nil {
		t.Fatalf("senha com 8 caracteres foi rejeitada: %v", err)
	}
}

func TestValidarSenhaRespeitaLimiteDoBcryptEmBytes(t *testing.T) {
	senha := strings.Repeat("á", 36) // 72 bytes, embora tenha 36 caracteres
	if err := validarSenha(senha); err != nil {
		t.Fatalf("senha com 72 bytes foi rejeitada: %v", err)
	}

	if err := validarSenha(senha + "a"); err == nil {
		t.Fatal("senha acima de 72 bytes foi aceita")
	}
}

func TestValidarCampoAutenticacaoLimitaTamanhoEBrancos(t *testing.T) {
	if err := validarCampoAutenticacao("   ", "nome", tamanhoMaximoNomeUsuario); err == nil {
		t.Fatal("campo formado só por espaços foi aceito")
	}
	if err := validarCampoAutenticacao(strings.Repeat("a", tamanhoMaximoNomeUsuario+1), "nome", tamanhoMaximoNomeUsuario); err == nil {
		t.Fatal("campo acima do limite foi aceito")
	}
}

func TestDecodificarJSONRejeitaConteudoExtra(t *testing.T) {
	requisicao := httptest.NewRequest(
		http.MethodPost,
		"/teste",
		strings.NewReader(`{"nome":"usuario"}{"nome":"outro"}`),
	)

	var destino struct {
		Nome string `json:"nome"`
	}
	if err := decodificarJSON(requisicao, &destino); err == nil {
		t.Fatal("mais de um valor JSON foi aceito")
	}
}
