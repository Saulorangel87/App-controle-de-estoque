import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Configuração padrão do Vite para React.
// build.target moderno reduz o tamanho do bundle final (menos polyfill),
// o que ajuda diretamente na métrica de performance do PageSpeed Insights.
export default defineConfig({
  plugins: [react()],
  build: {
    target: "es2020",
  },
});
