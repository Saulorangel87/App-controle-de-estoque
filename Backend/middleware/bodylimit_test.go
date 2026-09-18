package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLimitarCorpoEstruturadoRejeitaCorpoAcimaDoLimite(t *testing.T) {
	chamado := false
	handler := LimitarCorpoEstruturado(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		chamado = true
		_, _ = io.ReadAll(r.Body)
	}))

	requisicao := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(strings.Repeat("x", TamanhoMaximoCorpoEstruturado+1)),
	)
	requisicao.Header.Set("Content-Type", "application/json")
	resposta := httptest.NewRecorder()
	handler.ServeHTTP(resposta, requisicao)

	if resposta.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d; esperado 413", resposta.Code)
	}
	if chamado {
		t.Fatal("handler foi executado apesar do Content-Length acima do limite")
	}
}

func TestLimitarCorpoEstruturadoNaoInterfereEmMultipart(t *testing.T) {
	corpo := strings.Repeat("x", TamanhoMaximoCorpoEstruturado+1)
	var tamanhoLido int
	handler := LimitarCorpoEstruturado(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		conteudo, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("multipart não deveria ser limitado por este middleware: %v", err)
		}
		tamanhoLido = len(conteudo)
	}))

	requisicao := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(corpo))
	requisicao.Header.Set("Content-Type", "multipart/form-data; boundary=teste")
	handler.ServeHTTP(httptest.NewRecorder(), requisicao)

	if tamanhoLido != len(corpo) {
		t.Fatalf("tamanho lido = %d; esperado %d", tamanhoLido, len(corpo))
	}
}

func TestLimitarCorpoEstruturadoCortaCorpoSemContentLength(t *testing.T) {
	var erroLeitura error
	handler := LimitarCorpoEstruturado(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, erroLeitura = io.ReadAll(r.Body)
	}))

	requisicao := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(strings.Repeat("x", TamanhoMaximoCorpoEstruturado+1)),
	)
	requisicao.ContentLength = -1
	requisicao.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), requisicao)

	var erroLimite *http.MaxBytesError
	if !errors.As(erroLeitura, &erroLimite) {
		t.Fatalf("erro de leitura = %v; esperado http.MaxBytesError", erroLeitura)
	}
}
