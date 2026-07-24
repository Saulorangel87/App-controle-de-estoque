package main

import (
	"log"
	"net/http"
	"os"

	"controle-estoque/database"
	"controle-estoque/handlers"
	"controle-estoque/middleware"

	"github.com/joho/godotenv"
)

func main() {
	// Carrega variáveis de ambiente do .env (ex: JWT_SECRET) antes de qualquer outra coisa.
	// Em produção (Docker), as variáveis já vêm do docker-compose.yml e o .env não existe
	// dentro do container — godotenv.Load() simplesmente não encontra o arquivo e segue
	// em frente sem erro, então essa chamada é segura nos dois cenários.
	godotenv.Load()

	database.Conectar()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Autenticação e recuperação de senha (não exigem token — a pessoa ainda não está logada).
	mux.HandleFunc("POST /cadastro", handlers.Cadastrar)
	mux.HandleFunc("POST /login", handlers.Login)
	mux.HandleFunc("GET /recuperar-senha/pergunta", handlers.ObterPerguntaSeguranca)
	mux.HandleFunc("POST /recuperar-senha", handlers.RedefinirSenha)

	// Itens (todas exigem token válido via middleware.Autenticar).
	mux.HandleFunc("GET /itens", middleware.Autenticar(handlers.ListarItens))
	mux.HandleFunc("POST /itens", middleware.Autenticar(handlers.AdicionarItem))
	mux.HandleFunc("PUT /itens/{id}", middleware.Autenticar(handlers.EditarItem))
	mux.HandleFunc("DELETE /itens/{id}", middleware.Autenticar(handlers.ExcluirItem))
	mux.HandleFunc("POST /itens/{id}/retirar", middleware.Autenticar(handlers.RetirarItem))
	mux.HandleFunc("GET /itens/estoque-baixo", middleware.Autenticar(handlers.ItensEstoqueBaixo))

	log.Println("servidor rodando na porta 8080")
	if err := http.ListenAndServe(":8080", corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
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

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		proximo.ServeHTTP(w, r)
	})
}
