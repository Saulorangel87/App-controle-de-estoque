import { createContext, useContext, useState, useCallback, useMemo } from "react";
import { decodificarToken } from "../utils/jwt.js";

const AuthContext = createContext(null);

// Chave usada no localStorage do navegador para manter o usuário logado
// mesmo depois de fechar a aba (evita pedir login toda hora).
const CHAVE_ARMAZENAMENTO = "controle-estoque:token";

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() =>
    localStorage.getItem(CHAVE_ARMAZENAMENTO)
  );

  const entrar = useCallback((novoToken) => {
    localStorage.setItem(CHAVE_ARMAZENAMENTO, novoToken);
    setToken(novoToken);
  }, []);

  const sair = useCallback(() => {
    localStorage.removeItem(CHAVE_ARMAZENAMENTO);
    setToken(null);
  }, []);

  // O nome do usuário já vem dentro do próprio token (claim "nome"), definido
  // no login pelo backend — não precisamos de uma chamada extra à API só para exibir isso.
  const nome = useMemo(() => {
    if (!token) return "";
    const payload = decodificarToken(token);
    return payload?.nome ?? "";
  }, [token]);

  const valor = {
    token,
    nome,
    estaLogado: Boolean(token),
    entrar,
    sair,
  };

  return <AuthContext.Provider value={valor}>{children}</AuthContext.Provider>;
}

// Hook de conveniência para acessar o contexto sem importar useContext + AuthContext
// em toda tela que precisa do token.
export function useAuth() {
  return useContext(AuthContext);
}
