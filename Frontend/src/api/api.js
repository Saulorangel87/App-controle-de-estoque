// Em produção, VITE_API_URL é definida no momento do build (docker-compose.yml usa
// build args para isso, já que variáveis do Vite são "gravadas" no bundle estático,
// não lidas em tempo de execução). Em desenvolvimento local, sem essa variável, cai
// de volta para o backend rodando em localhost:8080.
const URL_BASE = import.meta.env.VITE_API_URL || "http://localhost:8080";

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
    // Anexa o status ao erro (não só a mensagem) — quem chamar essa função
    // consegue diferenciar, por exemplo, "senha errada" (401) de "bloqueado
    // por excesso de tentativas" (429), que precisam de mensagens diferentes.
    const erro = new Error(textoErro || "Erro na requisição");
    erro.status = resposta.status;
    throw erro;
  }

  // Respostas 204 (excluir item) não têm corpo — não tenta fazer parse de JSON vazio.
  if (resposta.status === 204) {
    return null;
  }

  return resposta.json();
}

// ---------- Autenticação ----------

export function cadastrar(nome, senha, perguntaSeguranca, respostaSeguranca) {
  return requisitar("/cadastro", {
    method: "POST",
    body: JSON.stringify({
      nome,
      senha,
      pergunta_seguranca: perguntaSeguranca,
      resposta_seguranca: respostaSeguranca,
    }),
  });
}

export function login(nome, senha) {
  return requisitar("/login", {
    method: "POST",
    body: JSON.stringify({ nome, senha }),
  });
}

export function obterPerguntaSeguranca(nome) {
  return requisitar(`/recuperar-senha/pergunta?nome=${encodeURIComponent(nome)}`, {
    method: "GET",
  });
}

export function redefinirSenha(nome, resposta, novaSenha) {
  return requisitar("/recuperar-senha", {
    method: "POST",
    body: JSON.stringify({ nome, resposta, nova_senha: novaSenha }),
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

// ---------- Importação de nota fiscal ----------

// Não reaproveita requisitar() porque upload de arquivo precisa de
// multipart/form-data — o navegador define esse Content-Type (com o
// "boundary" correto) sozinho quando NÃO especificamos um manualmente.
export async function importarNotaFiscal(arquivo, token) {
  const formData = new FormData();
  formData.append("arquivo", arquivo);

  const resposta = await fetch(`${URL_BASE}/notas-fiscais/importar`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    body: formData,
  });

  if (!resposta.ok) {
    const textoErro = await resposta.text();
    const erro = new Error(textoErro || "Erro na requisição");
    erro.status = resposta.status;
    throw erro;
  }

  return resposta.json();
}

// Importação via foto/print da nota (Fase 3 — OCR). Mesmo padrão multipart
// do upload de XML: o navegador define o Content-Type sozinho.
export async function importarNotaFiscalPorFoto(arquivo, token) {
  const formData = new FormData();
  formData.append("arquivo", arquivo);

  const resposta = await fetch(`${URL_BASE}/notas-fiscais/importar-foto`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    body: formData,
  });

  if (!resposta.ok) {
    const textoErro = await resposta.text();
    const erro = new Error(textoErro || "Erro na requisição");
    erro.status = resposta.status;
    throw erro;
  }

  return resposta.json();
}

// Importação via foto de cupom de papel físico — usa OCR na nuvem no
// backend (OCR.space) em vez do Tesseract local, mesma mecânica de envio.
export async function importarNotaFiscalPorFotoDePapel(arquivo, token) {
  const formData = new FormData();
  formData.append("arquivo", arquivo);

  const resposta = await fetch(`${URL_BASE}/notas-fiscais/importar-foto-papel`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    body: formData,
  });

  if (!resposta.ok) {
    const textoErro = await resposta.text();
    const erro = new Error(textoErro || "Erro na requisição");
    erro.status = resposta.status;
    throw erro;
  }

  return resposta.json();
}

// Importação via QR Code (Fase 2) — envia a URL lida no QR Code da NFC-e
// (colada ou capturada pela câmera) para o backend consultar na SEFAZ.
// Diferente de importarNotaFiscal, aqui o corpo é JSON simples, não
// multipart — não tem arquivo, só um texto.
export function importarNotaFiscalPorQRCode(url, token) {
  return requisitar(
    "/notas-fiscais/importar-qrcode",
    { method: "POST", body: JSON.stringify({ url }) },
    token
  );
}

export function confirmarImportacaoNota(entradas, token) {
  return requisitar(
    "/notas-fiscais/confirmar",
    { method: "POST", body: JSON.stringify(entradas) },
    token
  );
}
