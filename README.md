# Controle de Estoque — Frontend

Interface em React + Vite para o app de controle de estoque de mantimentos de casa.

Em produção: **https://estoque.devsaulo.com.br**

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
│   ├── ModalConfirmacao.jsx     → modal genérico de confirmação (usado para excluir item)
│   ├── ModalImportarNota.jsx    → upload, conferência e confirmação da importação de nota fiscal
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

## Importação de nota fiscal

Em vez de cadastrar item por item manualmente, dá para importar uma compra
inteira de uma vez pelo botão **"Importar nota"** do dashboard. O fluxo é o
mesmo pra qualquer uma das formas de entrada abaixo:

1. Envia a nota (arquivo XML, print de tela ou foto do cupom, dependendo da
   opção escolhida)
2. O backend lê os itens e compara com o estoque atual por nome (comparação
   exata + uma heurística de substring, pra pegar nomes parecidos como
   "Peito de Frango Resfriado" batendo com "Peito de Frango")
3. Uma tela de conferência mostra cada item como **Encontrado** (vai somar
   quantidade) ou **Novo item** (vai cadastrar) — nada é salvo ainda nesse
   ponto, e dá pra corrigir manualmente qualquer associação que a
   comparação automática não acertou
4. Só depois de confirmar é que o estoque é atualizado de fato

Três formas de entrada disponíveis hoje:

- **Arquivo XML da NF-e** — o caminho mais confiável, sem depender de OCR.
  Sempre a primeira opção se você tiver o XML salvo.
- **Print da tela de confirmação da SEFAZ** — depois de escanear o QR Code
  da nota no navegador, um print dessa tela é lido via OCR local (Tesseract,
  de graça, sem dependência externa) — funciona bem, já que é texto digital
  nítido.
- **Foto do cupom de papel físico** — lida via OCR numa API de nuvem
  (OCR.space), motor mais robusto pra lidar com ângulo, iluminação e papel
  térmico desbotado do que o Tesseract local (testado e confirmado: foto
  real de cupom é bem mais difícil que print de tela pra qualquer OCR).
  Precisa de uma chave de API configurada no backend (grátis, sem cartão) —
  detalhes em `NOTAS-DEPLOY.md`. Precisão ainda inconsistente em papel real,
  a tela de conferência é quem garante a correção final.

**Leitura por QR Code direto (sem print manual) está desativada de
propósito**: a consulta pública da SEFAZ-RJ exige reCAPTCHA v3 e proteção
anti-bot corporativa, que não é seguro nem confiável de automatizar via
scraping. Detalhes completos (e o porquê de cada decisão de OCR) em
`NOTAS-DEPLOY.md`.

## Licença

Este projeto está sob a licença MIT — veja o arquivo `LICENSE` na raiz do
repositório.

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

## 📄 Licença

Este projeto está licenciado sob a licença MIT.

Veja o arquivo [LICENSE](LICENSE) para mais detalhes.
