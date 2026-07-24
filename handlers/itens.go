package handlers

import (
	"encoding/json"
	"net/http"

	"controle-estoque/database"
	"controle-estoque/middleware"
	"controle-estoque/models"
)

func ListarItens(w http.ResponseWriter, r *http.Request) {
	usuarioID := r.Context().Value(middleware.UsuarioIDContexto).(int)

	rows, err := database.DB.Query(
		"SELECT id, usuario_id, nome, quantidade, unidade, local, estoque_minimo FROM itens WHERE usuario_id = ?",
		usuarioID,
	)
	if err != nil {
		http.Error(w, "erro ao buscar itens", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	itens := []models.Item{}
	for rows.Next() {
		var it models.Item
		if err := rows.Scan(&it.ID, &it.UsuarioID, &it.Nome, &it.Quantidade, &it.Unidade, &it.Local, &it.EstoqueMinimo); err != nil {
			http.Error(w, "erro ao ler itens", http.StatusInternalServerError)
			return
		}
		itens = append(itens, it)
	}

	json.NewEncoder(w).Encode(itens)
}

func AdicionarItem(w http.ResponseWriter, r *http.Request) {
	usuarioID := r.Context().Value(middleware.UsuarioIDContexto).(int)

	var it models.Item
	if err := json.NewDecoder(r.Body).Decode(&it); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}

	if it.Nome == "" || it.Unidade == "" || it.Local == "" {
		http.Error(w, "nome, unidade e local são obrigatórios", http.StatusBadRequest)
		return
	}

	resultado, err := database.DB.Exec(
		`INSERT INTO itens (usuario_id, nome, quantidade, unidade, local, estoque_minimo)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		usuarioID, it.Nome, it.Quantidade, it.Unidade, it.Local, it.EstoqueMinimo,
	)
	if err != nil {
		http.Error(w, "erro ao adicionar item", http.StatusInternalServerError)
		return
	}

	id, _ := resultado.LastInsertId()
	it.ID = int(id)
	it.UsuarioID = usuarioID

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(it)
}

type retirada struct {
	Quantidade float64 `json:"quantidade"`
}

func RetirarItem(w http.ResponseWriter, r *http.Request) {
	usuarioID := r.Context().Value(middleware.UsuarioIDContexto).(int)
	id := r.PathValue("id")

	var ret retirada
	if err := json.NewDecoder(r.Body).Decode(&ret); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}

	if ret.Quantidade <= 0 {
		http.Error(w, "quantidade deve ser maior que zero", http.StatusBadRequest)
		return
	}

	var quantidadeAtual float64
	row := database.DB.QueryRow(
		"SELECT quantidade FROM itens WHERE id = ? AND usuario_id = ?", id, usuarioID,
	)
	if err := row.Scan(&quantidadeAtual); err != nil {
		http.Error(w, "item não encontrado", http.StatusNotFound)
		return
	}

	novaQuantidade := quantidadeAtual - ret.Quantidade
	if novaQuantidade < 0 {
		novaQuantidade = 0
	}

	_, err := database.DB.Exec(
		"UPDATE itens SET quantidade = ? WHERE id = ? AND usuario_id = ?",
		novaQuantidade, id, usuarioID,
	)
	if err != nil {
		http.Error(w, "erro ao atualizar item", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]float64{"quantidade": novaQuantidade})
}

func ItensEstoqueBaixo(w http.ResponseWriter, r *http.Request) {
	usuarioID := r.Context().Value(middleware.UsuarioIDContexto).(int)

	rows, err := database.DB.Query(
		`SELECT id, usuario_id, nome, quantidade, unidade, local, estoque_minimo
		 FROM itens WHERE usuario_id = ? AND quantidade <= estoque_minimo`,
		usuarioID,
	)
	if err != nil {
		http.Error(w, "erro ao buscar itens", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	itens := []models.Item{}
	for rows.Next() {
		var it models.Item
		if err := rows.Scan(&it.ID, &it.UsuarioID, &it.Nome, &it.Quantidade, &it.Unidade, &it.Local, &it.EstoqueMinimo); err != nil {
			http.Error(w, "erro ao ler itens", http.StatusInternalServerError)
			return
		}
		itens = append(itens, it)
	}

	json.NewEncoder(w).Encode(itens)
}