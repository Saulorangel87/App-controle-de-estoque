package middleware

import (
	"context"
	"net/http"
	"strings"

	"controle-estoque/config"

	"github.com/golang-jwt/jwt/v5"
)

type chaveContexto string

const UsuarioIDContexto chaveContexto = "usuario_id"

func Autenticar(proximo http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cabecalho := r.Header.Get("Authorization")
		if cabecalho == "" {
			http.Error(w, "token não informado", http.StatusUnauthorized)
			return
		}

		partes := strings.Split(cabecalho, " ")
		if len(partes) != 2 || partes[0] != "Bearer" {
			http.Error(w, "formato de token inválido", http.StatusUnauthorized)
			return
		}

		tokenString := partes[1]

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return config.ChaveSecreta(), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "token inválido ou expirado", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}

		usuarioID := int(claims["usuario_id"].(float64))

		ctx := context.WithValue(r.Context(), UsuarioIDContexto, usuarioID)
		proximo(w, r.WithContext(ctx))
	}
}
