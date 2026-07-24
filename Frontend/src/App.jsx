import { Routes, Route, Navigate } from "react-router-dom";
import Login from "./pages/Login.jsx";
import Dashboard from "./pages/Dashboard.jsx";
import RotaProtegida from "./components/RotaProtegida.jsx";
import Footer from "./components/Footer.jsx";

export default function App() {
  return (
    // app-shell empurra o rodapé para o fim da tela via flexbox: fica colado
    // embaixo em páginas curtas, e desce normalmente com o conteúdo em páginas
    // longas — evita sobrepor a tabela de itens, o que um footer "fixed" faria.
    <div className="app-shell">
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/"
          element={
            <RotaProtegida>
              <Dashboard />
            </RotaProtegida>
          }
        />
        {/* Qualquer rota desconhecida volta para a página inicial */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
      <Footer />
    </div>
  );
}
