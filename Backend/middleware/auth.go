package middleware

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

	"controle-estoque/config"
	"controle-estoque/database"

	"github.com/golang-jwt/jwt/v5"
)

type chaveContexto string

const UsuarioIDContexto chaveContexto = "usuario_id"

func Autenticar(proximo http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cabecalho := r.Header.Get("Authorization")
		if cabecalho == "" {
			cookie, err := r.Cookie("estoque_sessao")
			if err != nil || cookie.Value == "" {
				http.Error(w, "token não informado", http.StatusUnauthorized)
				return
			}
			cabecalho = "Bearer " + cookie.Value
		}

		partes := strings.Split(cabecalho, " ")
		if len(partes) != 2 || partes[0] != "Bearer" {
			http.Error(w, "formato de token inválido", http.StatusUnauthorized)
			return
		}

		tokenString := partes[1]

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(t *jwt.Token) (interface{}, error) {
				if t.Method != jwt.SigningMethodHS256 {
					return nil, fmt.Errorf("algoritmo JWT não permitido: %s", t.Method.Alg())
				}
				return config.ChaveSecreta(), nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		)
		if err != nil || !token.Valid {
			http.Error(w, "token inválido ou expirado", http.StatusUnauthorized)
			return
		}

		claimsValidados, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}

		usuarioIDFloat, ok := claimsValidados["usuario_id"].(float64)
		if !ok || usuarioIDFloat <= 0 || math.Trunc(usuarioIDFloat) != usuarioIDFloat {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}
		usuarioID := int(usuarioIDFloat)
		if float64(usuarioID) != usuarioIDFloat {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}

		tokenVersao, ok := claimsValidados["token_versao"].(float64)
		if !ok || tokenVersao < 0 || math.Trunc(tokenVersao) != tokenVersao {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}
		var tokenVersaoAtual int
		if err := database.DB.QueryRow(
			"SELECT token_versao FROM usuarios WHERE id = ?", usuarioID,
		).Scan(&tokenVersaoAtual); err != nil || float64(tokenVersaoAtual) != tokenVersao {
			http.Error(w, "token inválido ou expirado", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UsuarioIDContexto, usuarioID)
		proximo(w, r.WithContext(ctx))
	}
}
