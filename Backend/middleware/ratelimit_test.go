package middleware

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func limparRateLimitTeste(t *testing.T) {
	t.Helper()
	tentativasPorChave.mu.Lock()
	defer tentativasPorChave.mu.Unlock()
	tentativasPorChave.tentativas = make(map[string][]time.Time)
}

func TestChavesDeRateLimitIncluemContaEPreservamCorpo(t *testing.T) {
	limparRateLimitTeste(t)
	requisicao := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"nome":" Usuário Teste ","senha":"segredo"}`),
	)
	requisicao.Header.Set("Content-Type", "application/json")

	chaves := chavesDeRateLimit(requisicao)
	if len(chaves) != 2 || chaves[1] != "conta:usuário teste" {
		t.Fatalf("chaves = %#v; esperadas chave de IP e conta normalizada", chaves)
	}

	corpo, err := io.ReadAll(requisicao.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(corpo) != `{"nome":" Usuário Teste ","senha":"segredo"}` {
		t.Fatalf("corpo restaurado = %q", corpo)
	}
}

func TestLimitarTentativasBloqueiaContaMesmoComIPsDiferentes(t *testing.T) {
	limparRateLimitTeste(t)
	handler := LimitarTentativas(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	for tentativa := 0; tentativa < maxTentativas; tentativa++ {
		requisicao := httptest.NewRequest(
			http.MethodPost,
			"/login",
			strings.NewReader(`{"nome":"mesma-conta","senha":"senha"}`),
		)
		requisicao.RemoteAddr = fmt.Sprintf("192.0.2.%d:1234", tentativa+1)
		requisicao.Header.Set("Content-Type", "application/json")
		resposta := httptest.NewRecorder()
		handler(resposta, requisicao)
		if resposta.Code != http.StatusUnauthorized {
			t.Fatalf("tentativa %d retornou %d; esperada 401", tentativa+1, resposta.Code)
		}
	}

	requisicao := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"nome":"mesma-conta","senha":"senha"}`),
	)
	requisicao.RemoteAddr = "192.0.2.99:1234"
	requisicao.Header.Set("Content-Type", "application/json")
	resposta := httptest.NewRecorder()
	handler(resposta, requisicao)

	if resposta.Code != http.StatusTooManyRequests {
		t.Fatalf("sexta tentativa retornou %d; esperada 429", resposta.Code)
	}
}
