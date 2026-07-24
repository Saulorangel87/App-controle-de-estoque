# Notas — Controle de Estoque

## Stack
- Backend: Go (net/http puro), SQLite (modernc.org/sqlite), bcrypt + JWT
- Frontend: React 19 + Vite + react-router-dom, CSS puro (sem libs de UI)
- Deploy: Docker Compose (backend + frontend com nginx) no Oracle Cloud, via Cloudflare Tunnel

## Estrutura

App controle de estoque/
├── docker-compose.yml
├── Backend.Dockerfile
├── Frontend.Dockerfile
├── nginx.conf
├── .gitignore
├── Backend/
│ ├── main.go
│ ├── config/ (chave JWT via variável de ambiente)
│ ├── database/ (conexão SQLite + migração de colunas)
│ ├── models/
│ ├── handlers/ (auth + itens)
│ └── middleware/ (autenticação JWT)
└── Frontend/
├── src/api/ (cliente HTTP centralizado)
├── src/context/ (Auth + Theme)
├── src/components/ (Footer, modais, tabela, filtros, cards)
├── src/pages/ (Login, Dashboard)
└── src/utils/ (JWT decode, perguntas de segurança)


## Funcionalidades prontas
- Cadastro (usuário/senha + pergunta de segurança) e login com JWT
- Recuperação de senha via pergunta de segurança (2 passos)
- CRUD completo de itens: adicionar, listar, editar, excluir, retirar (baixa de estoque)
- Alerta de estoque baixo (quantidade ≤ estoque mínimo) com filtro dedicado
- Dashboard com cards de resumo (total, estoque baixo, localizações em uso)
- Busca por nome + filtro por localização (Despensa, Geladeira, Freezer, Armário)
- Tema claro/escuro com persistência em localStorage
- Nome da conta (2 primeiros nomes) exibido no cabeçalho, extraído do JWT
- Rodapé com copyright, versão (v1.0.0) e links (LinkedIn, GitHub, email), fixo na base da tela
- Código comentado, com cuidados de acessibilidade (labels, foco visível, aria-live, aria-modal) e performance (sem fontes externas, sem libs de UI pesadas)

## Deploy — em andamento
- Dockerfiles e docker-compose.yml prontos (backend na porta 8090, frontend na 8091)
- CORS, caminho do banco (DB_PATH) e URL da API (VITE_API_URL) configuráveis via variável de ambiente/build arg
- Domínios planejados: estoque.devsaulo.com.br (frontend) e estoque-api.devsaulo.com.br (backend)
- Falta: gerar chave JWT de produção, subir para o GitHub, `git pull` + `docker compose up -d --build` na VM, apontar os subdomínios no Nginx Proxy Manager/Cloudflare

## Pendências conhecidas
- Ícones de editar/excluir na tabela precisam de ajuste mobile (ficar pequenos ao lado do botão "Retirar" em vez de sumir/cortar)
- Trocar a `ChaveSecreta` do `.env` local se ela algum dia vazar (nunca reaproveitar a de produção)