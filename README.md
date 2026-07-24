# Controle de Estoque — Frontend

Interface em React + Vite para o app de controle de estoque de mantimentos de casa.

## Como rodar

```
npm install
npm run dev
```

O Vite sobe em `http://localhost:5173` por padrão. O backend em Go precisa estar
rodando em `http://localhost:8080` ao mesmo tempo (o CORS do backend já está
liberado especificamente para `http://localhost:5173`).

## Estrutura

```
src/
├── api/api.js                 → todas as chamadas à API centralizadas aqui
├── context/AuthContext.jsx    → guarda o token JWT (localStorage) e expõe entrar()/sair()
├── components/
│   ├── RotaProtegida.jsx      → redireciona para /login se não houver token
│   ├── CartoesResumo.jsx      → cards do dashboard (total, estoque baixo, locais)
│   ├── BarraFiltros.jsx       → busca por nome + filtro por localização
│   ├── TabelaItens.jsx        → lista os itens com ações de editar/excluir/retirar
│   ├── ModalFormularioItem.jsx→ modal único usado tanto para adicionar quanto editar
│   └── ModalRetirar.jsx       → modal para dar baixa em uma quantidade consumida
└── pages/
    ├── Login.jsx               → login e cadastro na mesma tela (alterna o modo)
    └── Dashboard.jsx            → tela principal, junta tudo acima
```

## Decisões de performance, acessibilidade e SEO

- **Sem fontes externas**: usa a pilha de fontes do sistema operacional
  (`-apple-system, "Segoe UI", Roboto...`), eliminando uma requisição de rede
  e evitando atraso/flash de texto — ajuda diretamente o LCP e o CLS.
- **Sem bibliotecas de UI pesadas**: CSS puro em `index.css`, com variáveis para
  cor/espaçamento. Menos JavaScript para baixar e interpretar.
- **`build.target: "es2020"`** no `vite.config.js`: gera menos código de
  compatibilidade para navegadores antigos, reduzindo o tamanho do bundle final.
- **Foco de teclado sempre visível** (`:focus-visible` no CSS) — nunca é removido.
- **Labels associados a todo input** (`htmlFor`/`id`), inclusive nos filtros
  (com `sr-only` quando o rótulo visível seria redundante com o placeholder).
- **`aria-live="polite"`** na área da tabela, para leitores de tela perceberem
  quando a lista termina de carregar.
- **`role="dialog"` + `aria-modal`** nos modais, e foco automático no primeiro
  campo ao abrir.
- **`prefers-reduced-motion`** respeitado no CSS global.
- **`meta name="robots" content="noindex"`** no `index.html`: é um app pessoal
  autenticado, não faz sentido para buscadores indexarem — mas o `title` e a
  `meta description` continuam corretos para quando o link for compartilhado.

## Pendências conhecidas

- Trocar a URL fixa do backend em `src/api/api.js` por uma variável de ambiente
  do Vite (`import.meta.env.VITE_API_URL`) antes de qualquer deploy em produção.
- Ajustar a origem liberada no CORS do backend (`corsMiddleware` em `main.go`)
  para o domínio real quando o frontend for publicado.
