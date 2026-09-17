package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"controle-estoque/database"
	"controle-estoque/middleware"

	_ "modernc.org/sqlite"
)

func configurarBancoImportacaoTeste(t *testing.T) {
	t.Helper()
	banco, err := sql.Open("sqlite", t.TempDir()+"/importacao.db")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := banco.Exec(`
		CREATE TABLE itens (
			id INTEGER PRIMARY KEY,
			usuario_id INTEGER NOT NULL,
			nome TEXT NOT NULL,
			quantidade REAL NOT NULL,
			unidade TEXT NOT NULL,
			local TEXT NOT NULL,
			estoque_minimo REAL NOT NULL
		);
		CREATE TABLE confirmacoes_importacao (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			usuario_id INTEGER NOT NULL,
			chave TEXT NOT NULL,
			atualizados INTEGER NOT NULL,
			criados INTEGER NOT NULL,
			criada_em TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (usuario_id, chave)
		);
		INSERT INTO itens (id, usuario_id, nome, quantidade, unidade, local, estoque_minimo)
		VALUES (1, 1, 'Arroz', 5, 'kg', 'Despensa', 1),
		       (2, 2, 'Feijão', 8, 'kg', 'Despensa', 1);
	`); err != nil {
		banco.Close()
		t.Fatal(err)
	}
	database.DB = banco
	t.Cleanup(func() {
		banco.Close()
		database.DB = nil
	})
}

func requisicaoConfirmacaoTeste(t *testing.T, chave string, entradas any) *httptest.ResponseRecorder {
	t.Helper()
	corpo, err := json.Marshal(map[string]any{"chave": chave, "entradas": entradas})
	if err != nil {
		t.Fatal(err)
	}
	requisicao := httptest.NewRequest(http.MethodPost, "/notas-fiscais/confirmar", bytes.NewReader(corpo))
	contexto := context.WithValue(requisicao.Context(), middleware.UsuarioIDContexto, 1)
	requisicao = requisicao.WithContext(contexto)
	resposta := httptest.NewRecorder()
	ConfirmarImportacao(resposta, requisicao)
	return resposta
}

func TestConfirmarImportacaoEIdempotente(t *testing.T) {
	configurarBancoImportacaoTeste(t)
	entradas := []map[string]any{{
		"item_id": 1, "nome": "Arroz", "quantidade": 2,
		"unidade": "kg", "local": "Despensa", "estoque_minimo": 0,
	}}

	primeira := requisicaoConfirmacaoTeste(t, "nota-001", entradas)
	if primeira.Code != http.StatusOK {
		t.Fatalf("primeira confirmação retornou %d: %s", primeira.Code, primeira.Body.String())
	}
	segunda := requisicaoConfirmacaoTeste(t, "nota-001", entradas)
	if segunda.Code != http.StatusOK {
		t.Fatalf("repetição retornou %d: %s", segunda.Code, segunda.Body.String())
	}

	var quantidade float64
	if err := database.DB.QueryRow("SELECT quantidade FROM itens WHERE id = 1").Scan(&quantidade); err != nil {
		t.Fatal(err)
	}
	if quantidade != 7 {
		t.Fatalf("quantidade = %v; esperada 7 após uma única aplicação", quantidade)
	}
}

func TestConfirmarImportacaoFazRollbackQuandoItemNaoExiste(t *testing.T) {
	configurarBancoImportacaoTeste(t)
	entradas := []map[string]any{
		{"item_id": 1, "nome": "Arroz", "quantidade": 2, "unidade": "kg", "local": "Despensa"},
		{"item_id": 999, "nome": "Inexistente", "quantidade": 1, "unidade": "kg", "local": "Despensa"},
	}

	resposta := requisicaoConfirmacaoTeste(t, "nota-rollback", entradas)
	if resposta.Code != http.StatusNotFound {
		t.Fatalf("confirmação inválida retornou %d; esperado 404", resposta.Code)
	}

	var quantidade float64
	if err := database.DB.QueryRow("SELECT quantidade FROM itens WHERE id = 1").Scan(&quantidade); err != nil {
		t.Fatal(err)
	}
	if quantidade != 5 {
		t.Fatalf("quantidade = %v; esperada 5 após rollback", quantidade)
	}
}

func TestConfirmarImportacaoNaoAcessaItemDeOutroUsuario(t *testing.T) {
	configurarBancoImportacaoTeste(t)
	entradas := []map[string]any{{
		"item_id": 2, "nome": "Feijão", "quantidade": 2,
		"unidade": "kg", "local": "Despensa", "estoque_minimo": 0,
	}}

	resposta := requisicaoConfirmacaoTeste(t, "nota-outro-usuario", entradas)
	if resposta.Code != http.StatusNotFound {
		t.Fatalf("acesso cruzado retornou %d; esperado 404", resposta.Code)
	}

	var quantidade float64
	if err := database.DB.QueryRow("SELECT quantidade FROM itens WHERE id = 2").Scan(&quantidade); err != nil {
		t.Fatal(err)
	}
	if quantidade != 8 {
		t.Fatalf("quantidade do outro usuário = %v; esperada 8", quantidade)
	}
}
