import { Navigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext.jsx";

// Envolve páginas que exigem login. Se não houver token, redireciona para /login
// em vez de deixar a tela tentar chamar a API e falhar com 401.
export default function RotaProtegida({ children }) {
  const { estaLogado, carregandoSessao } = useAuth();

  if (carregandoSessao) {
    return <p className="estado-vazio">Verificando sessão...</p>;
  }

  if (!estaLogado) {
    return <Navigate to="/login" replace />;
  }

  return children;
}
