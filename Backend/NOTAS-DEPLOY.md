# Notas — Controle de Estoque (Backend)

## Stack
- Go (net/http puro, sem framework)
- SQLite via `modernc.org/sqlite` (driver puro Go)
- Autenticação: usuário/senha (bcrypt) + JWT (golang-jwt/jwt/v5)

## Estrutura do projeto
```
controle-estoque/
├── main.go
├── config/
│   └── config.go       (ChaveSecreta do JWT — isolada para evitar import cycle)
├── database/
│   └── db.go            (conexão SQLite + criação de tabelas)
├── models/
│   ├── usuario.go
│   └── item.go
├── handlers/
│   ├── auth.go           (cadastro e login)
│   └── itens.go          (CRUD de itens)
└── middleware/
    └── auth.go            (validação de JWT via header Authorization)
```

## O que foi feito hoje
- Estrutura de pastas do backend criada e módulo Go inicializado
- Conexão com SQLite funcionando, tabelas `usuarios` e `itens` criadas automaticamente na primeira execução
- Cadastro de usuário (nome + senha, sem email) com hash bcrypt
- Login retornando token JWT (validade de 72h)
- Middleware de autenticação protegendo as rotas de itens via header `Authorization: Bearer <token>`
- Resolvido import cycle entre `handlers` e `middleware`, criando o pacote `config` para isolar a chave secreta do JWT
- CRUD de itens:
  - `POST /itens` — adicionar item (nome, quantidade, unidade, local, estoque mínimo)
  - `GET /itens` — listar itens do usuário logado
  - `POST /itens/{id}/retirar` — dar baixa em uma quantidade (trava em 0, nunca fica negativo)
  - `GET /itens/estoque-baixo` — lista itens onde quantidade ≤ estoque mínimo
- Todas as rotas testadas via curl, ponta a ponta, com resultado esperado confirmado

## Decisões tomadas
- Login simples (usuário/senha), sem verificação de email — diferente do Controle de Despesas
- Cada usuário só enxerga seus próprios itens (filtro por `usuario_id` em todas as queries)
- Quando o item chega a 0, continua na lista (não some) — aparece como "acabou"
- Localização física do item (Despensa, Geladeira, Freezer, Armário) faz parte do modelo do item
- Leitor de código de barras e alerta de validade foram descartados do escopo inicial (avaliar depois)

## Pendente
- Rota de editar/excluir item
- Dashboard (resumo: total de itens, itens com estoque baixo, localizações)
- Frontend em React + Vite, seguindo referência visual aprovada (lista de itens estilo dashboard de produtos, com busca e filtro por localização, cards de resumo no topo)
- Trocar a `ChaveSecreta` fixa em `config/config.go` por variável de ambiente antes de qualquer deploy público
