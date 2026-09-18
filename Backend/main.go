package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"controle-estoque/config"
	"controle-estoque/database"
	"controle-estoque/handlers"
	"controle-estoque/middleware"

	"github.com/joho/godotenv"
)

func main() {
	// Carrega variáveis de ambiente do .env (ex: JWT_SECRET) antes de qualquer outra coisa.
	godotenv.Load()
	if err := config.Validar(); err != nil {
		log.Fatal("configuração inválida: ", err)
	}

	database.Conectar()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Autenticação, confirmação de e-mail e recuperação de senha não exigem token.
	// Login/cadastro e códigos passam por LimitarTentativas: 5 falhas por IP e conta
	// a cada 5 min.
	mux.HandleFunc("POST /cadastro", middleware.LimitarTentativas(handlers.Cadastrar))
	mux.HandleFunc("POST /cadastro/verificar-email", middleware.LimitarTentativas(handlers.VerificarEmailCadastro))
	mux.HandleFunc("POST /cadastro/reenviar-verificacao", middleware.LimitarTentativas(handlers.ReenviarVerificacaoEmail))
	mux.HandleFunc("POST /login", middleware.LimitarTentativas(handlers.Login))
	mux.HandleFunc("GET /sessao", middleware.Autenticar(handlers.SessaoAtual))
	mux.HandleFunc("POST /logout", middleware.Autenticar(handlers.EncerrarSessao))
	mux.HandleFunc("POST /recuperar-senha/solicitar", middleware.LimitarTentativas(handlers.SolicitarRecuperacaoSenha))
	mux.HandleFunc("POST /recuperar-senha", middleware.LimitarTentativas(handlers.RedefinirSenha))
	mux.HandleFunc("GET /conta/email", middleware.Autenticar(handlers.ObterEmailConta))
	mux.HandleFunc("POST /conta/email", middleware.Autenticar(handlers.SolicitarEmailConta))
	mux.HandleFunc("POST /conta/email/verificar", middleware.Autenticar(handlers.VerificarEmailConta))

	// Itens (todas exigem token válido via middleware.Autenticar).
	mux.HandleFunc("GET /itens", middleware.Autenticar(handlers.ListarItens))
	mux.HandleFunc("POST /itens", middleware.Autenticar(handlers.AdicionarItem))
	mux.HandleFunc("PUT /itens/{id}", middleware.Autenticar(handlers.EditarItem))
	mux.HandleFunc("DELETE /itens/{id}", middleware.Autenticar(handlers.ExcluirItem))
	mux.HandleFunc("POST /itens/{id}/retirar", middleware.Autenticar(handlers.RetirarItem))
	mux.HandleFunc("GET /itens/estoque-baixo", middleware.Autenticar(handlers.ItensEstoqueBaixo))

	// Importação de nota fiscal. "importar" (Fase 1, via upload de XML),
	// "importar-foto" (Fase 3, via OCR de foto/print da nota) e
	// "importar-qrcode" (Fase 2, desativada — ver services/nfce_scraper.go)
	// devolvem a mesma prévia — "confirmar" é compartilhado pelos fluxos.
	// Todas exigem login.
	// Importação de nota fiscal. "importar" (Fase 1, via upload de XML),
	// "importar-foto" (Fase 3, via Tesseract local — melhor pra print de
	// tela), "importar-foto-papel" (Fase 3b, via OCR na nuvem — melhor pra
	// foto de cupom físico) e "importar-qrcode" (Fase 2, desativada — ver
	// services/nfce_scraper.go) devolvem a mesma prévia — "confirmar" é
	// compartilhado pelos fluxos. Todas exigem login.
	mux.HandleFunc("POST /notas-fiscais/importar", middleware.Autenticar(handlers.ImportarNotaFiscal))
	mux.HandleFunc("POST /notas-fiscais/importar-foto", middleware.Autenticar(handlers.ImportarNotaFiscalPorFoto))
	mux.HandleFunc("POST /notas-fiscais/importar-foto-papel", middleware.Autenticar(handlers.ImportarNotaFiscalPorFotoDePapel))
	mux.HandleFunc("POST /notas-fiscais/importar-qrcode", middleware.Autenticar(handlers.ImportarNotaFiscalPorQRCode))
	mux.HandleFunc("POST /notas-fiscais/confirmar", middleware.Autenticar(handlers.ConfirmarImportacao))

	// Timeouts explícitos em vez de usar os defaults do net/http (que são
	// SEM LIMITE — uma conexão travada ou muito lenta ficaria presa
	// indefinidamente do lado do Go, mesmo que algum proxy no meio do
	// caminho já tenha desistido dela há muito tempo). WriteTimeout um
	// pouco generoso porque o fluxo de OCR na nuvem (foto de papel) chama
	// uma API externa, que pode legitimamente levar alguns segundos.
	servidor := &http.Server{
		Addr:              ":8080",
		Handler:           middleware.LimitarCorpoEstruturado(corsMiddleware(protegerCSRF(mux))),
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      45 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Println("servidor rodando na porta 8080")
	if err := servidor.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// protegerCSRF exige a origem configurada para operações mutáveis que usam o
// cookie de sessão. Chamadas sem cookie (por exemplo, login ou clientes que
// ainda usam Authorization) mantêm o comportamento anterior; o cookie também
// usa SameSite=Lax como segunda camada.
func protegerCSRF(proximo http.Handler) http.Handler {
	origemEsperada := os.Getenv("CORS_ORIGIN")
	if origemEsperada == "" {
		origemEsperada = "http://localhost:5173"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutavel := r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete
		_, erroCookie := r.Cookie("estoque_sessao")
		temCookie := erroCookie == nil
		if mutavel && temCookie && strings.TrimSpace(r.Header.Get("Origin")) != origemEsperada {
			http.Error(w, "origem não permitida", http.StatusForbidden)
			return
		}
		proximo.ServeHTTP(w, r)
	})
}

// corsMiddleware libera as requisições vindas do frontend. A origem permitida vem da
// variável de ambiente CORS_ORIGIN (definida no docker-compose.yml em produção); em
// desenvolvimento local, se a variável não existir, cai de volta pro Vite (porta 5173).
func corsMiddleware(proximo http.Handler) http.Handler {
	origem := os.Getenv("CORS_ORIGIN")
	if origem == "" {
		origem = "http://localhost:5173"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origem)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		proximo.ServeHTTP(w, r)
	})
}
