// Decodifica a parte "payload" de um token JWT (segunda seção, separada por ponto)
// SEM verificar a assinatura — isso é papel exclusivo do backend. Aqui só lemos os
// dados públicos do token (nome, id) para exibir na interface.
export function decodificarToken(token) {
  try {
    const partePayload = token.split(".")[1];
    // JWT usa base64url (troca +/ por -_ e remove padding); atob espera base64 padrão.
    const base64 = partePayload.replace(/-/g, "+").replace(/_/g, "/");
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split("")
        .map((c) => "%" + c.charCodeAt(0).toString(16).padStart(2, "0"))
        .join("")
    );
    return JSON.parse(jsonPayload);
  } catch (erro) {
    return null;
  }
}
