package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"controle-estoque/database"
	"controle-estoque/middleware"
)

func TestRetirarItemProcessaRetiradasConcorrentesSemPerderMovimentacao(t *testing.T) {
	configurarBancoImportacaoTeste(t)
	database.DB.SetMaxOpenConns(1)
	if _, err := database.DB.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		t.Fatal(err)
	}

	retirar := func() int {
		corpo, err := json.Marshal(map[string]float64{"quantidade": 3})
		if err != nil {
			t.Error(err)
			return http.StatusInternalServerError
		}
		requisicao := httptest.NewRequest(http.MethodPost, "/itens/1/retirar", bytes.NewReader(corpo))
		requisicao.SetPathValue("id", "1")
		ctx := context.WithValue(requisicao.Context(), middleware.UsuarioIDContexto, 1)
		resposta := httptest.NewRecorder()
		RetirarItem(resposta, requisicao.WithContext(ctx))
		return resposta.Code
	}

	status := make(chan int, 2)
	var grupo sync.WaitGroup
	grupo.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer grupo.Done()
			status <- retirar()
		}()
	}
	grupo.Wait()
	close(status)

	for codigo := range status {
		if codigo != http.StatusOK {
			t.Fatalf("retirada concorrente retornou status %d; esperado 200", codigo)
		}
	}

	var quantidade float64
	if err := database.DB.QueryRow("SELECT quantidade FROM itens WHERE id = 1").Scan(&quantidade); err != nil {
		t.Fatal(err)
	}
	if quantidade != 0 {
		t.Fatalf("quantidade final = %v; esperada 0", quantidade)
	}
}
