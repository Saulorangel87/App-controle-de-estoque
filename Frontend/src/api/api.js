// Endereço do backend Go. Em produção, troque por uma variável de ambiente do Vite
// (import.meta.env.VITE_API_URL) apontando para o domínio real da API.
const URL_BASE = "http://localhost:8080";

/**
 * Função central de requisição. Todas as chamadas à API passam por aqui,
 * o que evita repetir a lógica de headers/token/erro em cada tela.
 *
 * @param {string} caminho - ex: "/itens"
 * @param {object} opcoes - opções extras do fetch (method, body, etc.)
 * @param {string|null} token - token JWT do usuário logado
 */
async function requisitar(caminho, opcoes = {}, token = null) {
  const cabecalhos = {
    "Content-Type": "application/json",
    ...opcoes.headers,
  };

  if (token) {
    cabecalhos.Authorization = `Bearer ${token}`;
  }

  const resposta = await fetch(`${URL_BASE}${caminho}`, {
    ...opcoes,
    headers: cabecalhos,
  });

  // O backend responde erros como texto simples (http.Error), não JSON.
  if (!resposta.ok) {
    const textoErro = await resposta.text();
    throw new Error(textoErro || "Erro na requisição");
  }

  // Respostas 204 (excluir item) não têm corpo — não tenta fazer parse de JSON vazio.
  if (resposta.status === 204) {
    return null;
  }

  return resposta.json();
}

// ---------- Autenticação ----------

export function cadastrar(nome, senha) {
  return requisitar("/cadastro", {
    method: "POST",
    body: JSON.stringify({ nome, senha }),
  });
}

export function login(nome, senha) {
  return requisitar("/login", {
    method: "POST",
    body: JSON.stringify({ nome, senha }),
  });
}

// ---------- Itens ----------

export function listarItens(token) {
  return requisitar("/itens", { method: "GET" }, token);
}

export function itensEstoqueBaixo(token) {
  return requisitar("/itens/estoque-baixo", { method: "GET" }, token);
}

export function adicionarItem(item, token) {
  return requisitar(
    "/itens",
    { method: "POST", body: JSON.stringify(item) },
    token
  );
}

export function editarItem(id, item, token) {
  return requisitar(
    `/itens/${id}`,
    { method: "PUT", body: JSON.stringify(item) },
    token
  );
}

export function excluirItem(id, token) {
  return requisitar(`/itens/${id}`, { method: "DELETE" }, token);
}

export function retirarItem(id, quantidade, token) {
  return requisitar(
    `/itens/${id}/retirar`,
    { method: "POST", body: JSON.stringify({ quantidade }) },
    token
  );
}
