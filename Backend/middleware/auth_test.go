package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"controle-estoque/config"
	"controle-estoque/database"

	"github.com/golang-jwt/jwt/v5"
	_ "modernc.org/sqlite"
)

const segredoTesteJWT = "12345678901234567890123456789012"

func TestAutenticarRejeitaClaimsInvalidos(t *testing.T) {
	t.Setenv("JWT_SECRET", segredoTesteJWT)

	tests := []struct {
		nome   string
		claims jwt.MapClaims
	}{
		{nome: "sem usuario", claims: jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}},
		{nome: "usuario decimal", claims: jwt.MapClaims{"usuario_id": 1.5, "exp": time.Now().Add(time.Hour).Unix()}},
		{nome: "expirado", claims: jwt.MapClaims{"usuario_id": 1, "exp": time.Now().Add(-time.Hour).Unix()}},
		{nome: "emissor incorreto", claims: jwt.MapClaims{
			"usuario_id": 1, "token_versao": 0, "iss": "outro-servico",
			"aud": config.AudienciaJWT, "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix(),
		}},
		{nome: "audiencia incorreta", claims: jwt.MapClaims{
			"usuario_id": 1, "token_versao": 0, "iss": config.EmissorJWT,
			"aud": "outro-cliente", "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix(),
		}},
	}

	for _, teste := range tests {
		t.Run(teste.nome, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, teste.claims)
			texto, err := token.SignedString(config.ChaveSecreta())
			if err != nil {
				t.Fatal(err)
			}

			requisicao := httptest.NewRequest(http.MethodGet, "/itens", nil)
			requisicao.Header.Set("Authorization", "Bearer "+texto)
			resposta := httptest.NewRecorder()
			Autenticar(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})(resposta, requisicao)

			if resposta.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d; esperado 401", resposta.Code)
			}
		})
	}
}

func TestAutenticarRejeitaAlgoritmoNaoPermitido(t *testing.T) {
	t.Setenv("JWT_SECRET", segredoTesteJWT)
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"usuario_id": 1,
		"exp":        time.Now().Add(time.Hour).Unix(),
	})
	texto, err := token.SignedString(config.ChaveSecreta())
	if err != nil {
		t.Fatal(err)
	}

	requisicao := httptest.NewRequest(http.MethodGet, "/itens", nil)
	requisicao.Header.Set("Authorization", "Bearer "+texto)
	resposta := httptest.NewRecorder()
	Autenticar(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})(resposta, requisicao)

	if resposta.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; esperado 401", resposta.Code)
	}
}

func TestAutenticarInvalidaTokenAposTrocaDeSenha(t *testing.T) {
	t.Setenv("JWT_SECRET", segredoTesteJWT)

	banco, err := sql.Open("sqlite", t.TempDir()+"/teste.db")
	if err != nil {
		t.Fatal(err)
	}
	defer banco.Close()
	if _, err := banco.Exec(`CREATE TABLE usuarios (id INTEGER PRIMARY KEY, token_versao INTEGER NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := banco.Exec(`INSERT INTO usuarios (id, token_versao) VALUES (1, 0)`); err != nil {
		t.Fatal(err)
	}
	database.DB = banco

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"usuario_id":   1,
		"token_versao": 0,
		"iss":          config.EmissorJWT,
		"aud":          config.AudienciaJWT,
		"iat":          time.Now().Unix(),
		"exp":          time.Now().Add(time.Hour).Unix(),
	})
	texto, err := token.SignedString(config.ChaveSecreta())
	if err != nil {
		t.Fatal(err)
	}

	status := autenticarStatus(texto)
	if status != http.StatusNoContent {
		t.Fatalf("token atual retornou status %d; esperado 204", status)
	}
	if _, err := database.DB.Exec("UPDATE usuarios SET token_versao = 1 WHERE id = 1"); err != nil {
		t.Fatal(err)
	}
	if status := autenticarStatus(texto); status != http.StatusUnauthorized {
		t.Fatalf("token antigo retornou status %d; esperado 401", status)
	}
}

func autenticarStatus(texto string) int {
	requisicao := httptest.NewRequest(http.MethodGet, "/itens", nil)
	requisicao.Header.Set("Authorization", "Bearer "+texto)
	resposta := httptest.NewRecorder()
	Autenticar(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})(resposta, requisicao)
	return resposta.Code
}
