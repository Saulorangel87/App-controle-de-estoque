# Notas — Controle de Estoque

**Versão atual: v1.1.0** · Em produção: https://estoque.devsaulo.com.br

## Stack
- Backend: Go (net/http puro), SQLite (modernc.org/sqlite), bcrypt + JWT
- Frontend: React 19 + Vite + react-router-dom, CSS puro (sem libs de UI), PWA instalável
- Deploy: Docker Compose (backend + frontend com nginx) no Oracle Cloud, via Cloudflare Tunnel

## Estrutura
```
App controle de estoque/
├── docker-compose.yml
├── Backend.Dockerfile
├── Frontend.Dockerfile
├── nginx.conf
├── .gitignore
├── Backend/
│   ├── main.go
│   ├── config/       → chave JWT via variável de ambiente
│   ├── database/     → conexão SQLite + migração de colunas (DB_PATH configurável)
│   ├── models/
│   ├── handlers/      → auth (cadastro/login/recuperação de senha) + itens
│   └── middleware/    → autenticação JWT
└── Frontend/
    ├── public/         → manifest.json, sw.js, ícones, robots.txt, llms.txt
    ├── src/api/         → cliente HTTP centralizado (URL via VITE_API_URL)
    ├── src/context/     → Auth + Theme
    ├── src/components/  → Footer, modais, tabela, filtros, cards
    ├── src/pages/       → Login, Dashboard
    └── src/utils/       → decodificação de JWT, perguntas de segurança
```

## Funcionalidades prontas
- Cadastro (usuário/senha + pergunta de segurança) e login com JWT
- Recuperação de senha via pergunta de segurança (fluxo de 2 passos)
- CRUD completo de itens: adicionar, listar, editar, excluir, retirar (baixa de estoque)
- Alerta de estoque baixo (quantidade ≤ estoque mínimo) com filtro dedicado
- Dashboard com cards de resumo (total, estoque baixo, localizações em uso)
- Busca por nome + filtro por localização (Despensa, Geladeira, Freezer, Armário)
- Tema claro/escuro com persistência em localStorage
- Nome da conta (2 primeiros nomes) exibido no cabeçalho, extraído do JWT
- Rodapé fixo com copyright, versão e links (LinkedIn, GitHub, email)
- **PWA instalável**: manifest.json + ícones + service worker — funciona como
  app de verdade no celular/desktop, não só atalho
- Código comentado, com cuidados de acessibilidade (labels, foco visível,
  aria-live, aria-modal) e performance (sem fontes externas, sem libs de UI pesadas)

## SEO e status público (desde v1.1.0)
- App deixou de ser só uso pessoal — agora é público, com planos de publicação na Play Store
- `meta robots` mudou de `noindex` para `index, follow`
- `robots.txt` liberando rastreamento (antes bloqueava por engano — bug corrigido)
- `link rel="canonical"` apontando para o domínio de produção
- `llms.txt` criado seguindo a spec llmstxt.org (auditoria de Navegação Agêntica)
- PageSpeed Insights: Desempenho 98, Acessibilidade 95, Práticas recomendadas 100,
  SEO 100, Navegação agêntica 3/3 — todos os critérios no verde

## Deploy — concluído
- Domínios: `estoque.devsaulo.com.br` (frontend, porta 8091) e
  `estoque-api.devsaulo.com.br` (backend, porta 8090)
- CORS, caminho do banco (`DB_PATH`) e URL da API (`VITE_API_URL`) configuráveis
  via variável de ambiente/build arg
- Banco SQLite persistido via volume Docker (`./data`), sobrevive a rebuilds
- Chave JWT de produção gerada separadamente da chave de desenvolvimento

## Pendências conhecidas
- Ícones de editar/excluir na tabela ficam com rolagem lateral em telas de
  celular muito estreitas — funcional, mas vale revisar o layout responsivo
- Publicação na Play Store (provavelmente via TWA — Trusted Web Activity —
  reaproveitando o manifest.json e o HTTPS já configurados)
- Revisar rate limiting no login/cadastro antes de divulgar amplamente (evitar
  força bruta), já que o app agora aceita público externo
