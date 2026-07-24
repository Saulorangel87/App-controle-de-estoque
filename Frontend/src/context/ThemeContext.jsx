import { createContext, useContext, useState, useCallback, useEffect } from "react";

const ThemeContext = createContext(null);
const CHAVE_ARMAZENAMENTO = "controle-estoque:tema";

export function ThemeProvider({ children }) {
  // Ordem de prioridade: preferência salva pelo usuário > preferência do sistema
  // operacional (prefers-color-scheme) > claro como último recurso.
  const [tema, setTema] = useState(() => {
    const salvo = localStorage.getItem(CHAVE_ARMAZENAMENTO);
    if (salvo) return salvo;
    const prefereEscuro = window.matchMedia(
      "(prefers-color-scheme: dark)"
    ).matches;
    return prefereEscuro ? "escuro" : "claro";
  });

  // O atributo data-theme no <html> é o que o CSS usa para aplicar as
  // variáveis de cor corretas (ver seleção [data-theme="escuro"] no index.css).
  useEffect(() => {
    document.documentElement.setAttribute("data-theme", tema);
    localStorage.setItem(CHAVE_ARMAZENAMENTO, tema);
  }, [tema]);

  const alternarTema = useCallback(() => {
    setTema((atual) => (atual === "claro" ? "escuro" : "claro"));
  }, []);

  return (
    <ThemeContext.Provider value={{ tema, alternarTema }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  return useContext(ThemeContext);
}
