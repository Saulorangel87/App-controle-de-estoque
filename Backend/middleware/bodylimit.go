package middleware

import (
	"net/http"
	"strings"
)

// TamanhoMaximoCorpoEstruturado limita requisições que não são multipart.
// Os endpoints de upload aplicam um limite próprio, mais amplo, antes de
// processar os arquivos. O limite evita que JSON ou corpos desconhecidos
// ocupem memória indefinidamente durante o decode.
const TamanhoMaximoCorpoEstruturado = 1 << 20 // 1 MiB

// LimitarCorpoEstruturado aplica um limite aos corpos que não são multipart.
// Isso cobre todos os endpoints JSON atuais, inclusive os que forem
// adicionados ao mux no futuro. O handler continua responsável por devolver a
// mensagem/status adequados quando o decoder encontrar o limite.
func LimitarCorpoEstruturado(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tipoConteudo := strings.ToLower(r.Header.Get("Content-Type"))
		if strings.HasPrefix(tipoConteudo, "multipart/form-data") {
			proximo.ServeHTTP(w, r)
			return
		}
		if r.ContentLength > TamanhoMaximoCorpoEstruturado {
			http.Error(w, "corpo da requisição muito grande", http.StatusRequestEntityTooLarge)
			return
		}
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, TamanhoMaximoCorpoEstruturado)
		}
		proximo.ServeHTTP(w, r)
	})
}
