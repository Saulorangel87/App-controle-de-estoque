import { createContext, useContext, useState, useCallback, useEffect } from "react";
import { encerrarSessao, sessaoAtual } from "../api/api.js";

const AuthContext = createContext(null);

// Chave usada no localStorage do navegador para manter o usuário logado
// mesmo depois de fechar a aba (evita pedir login toda hora).
const CHAVE_ARMAZENAMENTO = "controle-estoque:token";

export function AuthProvider({ children }) {
  const [token] = useState(null);
  const [nome, setNome] = useState("");
  const [estaLogado, setEstaLogado] = useState(false);
  const [carregandoSessao, setCarregandoSessao] = useState(true);

  useEffect(() => {
    localStorage.removeItem(CHAVE_ARMAZENAMENTO);
    sessaoAtual()
      .then((sessao) => {
        setNome(sessao.nome ?? "");
        setEstaLogado(true);
      })
      .catch(() => setEstaLogado(false))
      .finally(() => setCarregandoSessao(false));
  }, []);

  const entrar = useCallback((nomeUsuario) => {
    setNome(nomeUsuario ?? "");
    setEstaLogado(true);
  }, []);

  const sair = useCallback(() => {
    encerrarSessao().catch(() => {}).finally(() => {
      setNome("");
      setEstaLogado(false);
    });
  }, []);

  const valor = {
    token,
    nome,
    estaLogado,
    carregandoSessao,
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
