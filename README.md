# Controle de Estoque — Frontend

Interface em React + Vite para o app de controle de estoque de mantimentos de casa.

Em produção: **https://estoque.devsaulo.com.br**

<img width="1366" height="857" alt="Captura de tela 2026-07-24 212800" src="https://github.com/user-attachments/assets/a3c207d4-949a-491a-95e2-1b6f9580e3f1" />

<img width="1897" height="847" alt="Captura de tela 2026-07-24 212728" src="https://github.com/user-attachments/assets/7ce601a5-748d-494b-a2fc-2e849aa572de" />

## Instalável como app (PWA)

O app pode ser instalado no celular ou no computador como um aplicativo de
verdade, sem barra de endereço — não é só um atalho.

**Como instalar (Android/Chrome):** acesse o site, abra o menu (⋮) do
navegador e toque em **"Instalar app"**. No desktop, o Chrome mostra um ícone
de instalação (⊕) do lado direito da barra de endereço.

**Como instalar (iOS/Safari):** toque no ícone de compartilhar e escolha
**"Adicionar à Tela de Início"**.

Isso funciona por causa de três peças em `public/`:

- **`manifest.json`** — nome, cores e ícones do app; é o que faz o navegador
  reconhecer a página como instalável de verdade (não só um atalho)
- **`icon-192.png` e `icon-512.png`** — ícones em duas resoluções, exigidos
  pelo manifesto
- **`sw.js`** — service worker mínimo, registrado em `src/main.jsx`; cacheia o
  "shell" do app para abrir mais rápido numa segunda visita (nunca cacheia
  dados da API, só os arquivos estáticos do próprio frontend)

## Como rodar localmente

```
npm install
npm run dev
```

O Vite sobe em `http://localhost:5173` por padrão. O backend em Go precisa estar
rodando em `http://localhost:8080` ao mesmo tempo (o CORS do backend já está
liberado especificamente para `http://localhost:5173`).

> Service worker só funciona em `localhost` ou em produção (HTTPS) — em outro
> IP da rede local sem HTTPS, o navegador bloqueia o registro por segurança.

## Estrutura

```
src/
├── api/api.js                  → todas as chamadas à API centralizadas aqui
├── context/
│   ├── AuthContext.jsx          → guarda o token JWT (localStorage), expõe nome/entrar()/sair()
│   └── ThemeContext.jsx         → tema claro/escuro, persistido em localStorage
├── utils/
│   ├── jwt.js                    → decodifica o payload do JWT no navegador
│   └── perguntasSeguranca.js     → lista de perguntas de segurança do cadastro
├── components/
│   ├── RotaProtegida.jsx        → redireciona para /login se não houver token
│   ├── CartoesResumo.jsx        → cards do dashboard (total, estoque baixo, locais)
│   ├── BarraFiltros.jsx         → busca por nome, filtro por localização e por estoque baixo
│   ├── TabelaItens.jsx          → lista os itens com ações de editar/excluir/retirar
│   ├── ModalFormularioItem.jsx  → modal único usado tanto para adicionar quanto editar
│   ├── ModalRetirar.jsx         → modal para dar baixa em uma quantidade consumida
│   └── Footer.jsx               → rodapé fixo com créditos, versão e links
└── pages/
    ├── Login.jsx                 → login, cadastro e recuperação de senha (pergunta de segurança)
    └── Dashboard.jsx             → tela principal, junta tudo acima

public/
├── manifest.json, sw.js, icon-192.png, icon-512.png  → PWA (ver seção acima)
├── robots.txt                   → libera rastreamento (app é público desde a v1.1.0)
└── llms.txt                     → resumo do site para agentes de IA (spec llmstxt.org)
```

## Decisões de performance, acessibilidade e SEO

- **Sem fontes externas**: usa a pilha de fontes do sistema operacional
  (`-apple-system, "Segoe UI", Roboto...`), eliminando uma requisição de rede
  e evitando atraso/flash de texto — ajuda diretamente o LCP e o CLS.
- **Sem bibliotecas de UI pesadas**: CSS puro em `index.css`, com variáveis para
  cor/espaçamento (inclusive tema claro/escuro). Menos JavaScript para baixar.
- **`build.target: "es2020"`** no `vite.config.js`: gera menos código de
  compatibilidade para navegadores antigos, reduzindo o tamanho do bundle final.
- **Foco de teclado sempre visível** (`:focus-visible` no CSS) — nunca é removido.
- **Labels associados a todo input** (`htmlFor`/`id`), inclusive nos filtros
  (com `sr-only` quando o rótulo visível seria redundante com o placeholder).
- **`aria-live="polite"`** na área da tabela, **`role="dialog"` + `aria-modal`**
  nos modais com foco automático no primeiro campo, **`prefers-reduced-motion`**
  respeitado no CSS global.
- **SEO público**: `meta robots` em `index, follow` (o app é público desde a
  v1.1.0), `link rel="canonical"` apontando para o domínio real, `robots.txt`
  liberando rastreamento, e `llms.txt` seguindo a especificação llmstxt.org
  para a auditoria experimental de Navegação Agêntica do Lighthouse.

Última medição no PageSpeed Insights: Desempenho 98, Acessibilidade 95,
Práticas recomendadas 100, SEO 100, Navegação agêntica 3/3.

## Deploy

Produção roda via Docker Compose no Oracle Cloud (Backend na porta 8090,
Frontend com nginx na 8091), atrás de Cloudflare Tunnel, nos domínios
`estoque.devsaulo.com.br` (frontend) e `estoque-api.devsaulo.com.br` (backend).
Detalhes completos em `NOTAS-DEPLOY.md` na raiz do projeto.

## Pendências conhecidas

- Ícones de editar/excluir na tabela de itens ficam com rolagem lateral em
  telas de celular muito estreitas — funcional, mas vale revisar o layout
  responsivo dessa linha da tabela em algum momento.
- Planejado: publicação na Play Store (provavelmente via TWA, aproveitando o
  manifest.json e o HTTPS que já existem).
