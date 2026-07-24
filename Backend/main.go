package main

import (
	"log"
	"net/http"

	"controle-estoque/database"
	"controle-estoque/handlers"
	"controle-estoque/middleware"

	"github.com/joho/godotenv"
)

func main() {
	// Carrega variáveis de ambiente do .env (ex: JWT_SECRET) antes de qualquer outra coisa.
	godotenv.Load()

	database.Conectar()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Autenticação (não exige token).
	mux.HandleFunc("POST /cadastro", handlers.Cadastrar)
	mux.HandleFunc("POST /login", handlers.Login)

	// Itens (todas exigem token válido via middleware.Autenticar).
	mux.HandleFunc("GET /itens", middleware.Autenticar(handlers.ListarItens))
	mux.HandleFunc("POST /itens", middleware.Autenticar(handlers.AdicionarItem))
	mux.HandleFunc("PUT /itens/{id}", middleware.Autenticar(handlers.EditarItem))
	mux.HandleFunc("DELETE /itens/{id}", middleware.Autenticar(handlers.ExcluirItem))
	mux.HandleFunc("POST /itens/{id}/retirar", middleware.Autenticar(handlers.RetirarItem))
	mux.HandleFunc("GET /itens/estoque-baixo", middleware.Autenticar(handlers.ItensEstoqueBaixo))

	log.Println("servidor rodando na porta 8080")
	// Envolve o mux com o middleware de CORS para o frontend (rodando em outra porta) conseguir chamar a API.
	if err := http.ListenAndServe(":8080", corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

// corsMiddleware libera as requisições vindas do frontend em desenvolvimento (Vite, porta 5173)
// e trata a requisição de preflight (OPTIONS) que o navegador envia antes de POST/PUT/DELETE.
func corsMiddleware(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		proximo.ServeHTTP(w, r)
	})
}
