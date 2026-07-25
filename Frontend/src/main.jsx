import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App from "./App.jsx";
import { AuthProvider } from "./context/AuthContext.jsx";
import { ThemeProvider } from "./context/ThemeContext.jsx";
import "./index.css";

// ThemeProvider fica fora de tudo porque afeta a página inteira, incluindo a
// tela de login (que roda antes do AuthProvider ter um usuário autenticado).
createRoot(document.getElementById("root")).render(
  <StrictMode>
    <BrowserRouter>
      <ThemeProvider>
        <AuthProvider>
          <App />
        </AuthProvider>
      </ThemeProvider>
    </BrowserRouter>
  </StrictMode>
);

// Registra o service worker (public/sw.js) depois que a página termina de
// carregar, para não competir por recursos de rede com o carregamento inicial.
if ("serviceWorker" in navigator) {
  window.addEventListener("load", () => {
    navigator.serviceWorker.register("/sw.js").catch((erro) => {
      console.error("Falha ao registrar o service worker:", erro);
    });
  });
}
