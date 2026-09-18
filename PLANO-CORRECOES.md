# Plano de correções e melhorias — Controle de Estoque

Documento vivo para orientar a correção dos riscos encontrados na análise de
17/09/2026. Cada alteração de código deve atualizar este documento com o
status, os arquivos afetados e as validações executadas.

## Regras de execução

- Corrigir em pequenas etapas, preservando as funcionalidades existentes.
- Validar autorização por usuário em toda operação autenticada.
- Fazer primeiro as correções de segurança e integridade de dados.
- Não considerar uma correção concluída sem teste ou evidência de validação.
- Não registrar segredos, senhas, tokens ou dados fiscais nos logs.
- Não fazer commit, merge, deploy ou publicação sem autorização explícita.

## Legenda

- **Pendente**: ainda não implementado.
- **Em andamento**: alteração iniciada, aguardando validação completa.
- **Concluído**: código alterado e critérios de aceite verificados.
- **Aceito como risco**: decisão consciente, com justificativa registrada.

## Fase 0 — Preparação e baseline

### P0. Inventário e baseline técnico

- **Status:** Concluído — baseline revalidado após as correções.
- **Escopo:** confirmar estrutura, documentação, estado do Git, build e testes.
- **Critérios de aceite:** manter `git diff --check` limpo; registrar falhas de
  ambiente separadamente de falhas do código; não incluir artefatos gerados.
- **Observação:** os testes Go foram executados com `GOCACHE` temporário; o build
  frontend foi validado fora da restrição local do esbuild e também no CI. Falhas
  de sandbox foram separadas de falhas do código.

## Fase 1 — Segurança crítica e autenticação

### P1. Endurecer validação de JWT

- **Status:** Concluído.
- **Arquivos previstos:** `Backend/middleware/auth.go`,
  `Backend/config/config.go`, possivelmente `Backend/main.go`.
- **Ações:**
  - [x] rejeitar inicialização se `JWT_SECRET` estiver vazio ou fraco;
  - [x] aceitar somente HS256;
  - [x] validar `usuario_id` com tipo numérico inteiro e positivo;
  - [x] usar o parser de claims do JWT, que valida a expiração;
  - [x] evitar assertions que poderiam causar panic;
  - [x] incluir e validar `iss`, `aud` e `iat` nos tokens emitidos;
  - [x] invalidar sessões após troca de senha por `token_versao`.
- **Critérios de aceite:** token sem assinatura válida, algoritmo diferente,
  segredo ausente, claim ausente ou claim com tipo inválido retorna 401 sem panic.
- **Evidência:** `go test ./...` e `go vet ./...` passaram em 17/09/2026,
  incluindo parser restrito a HS256, expiração obrigatória, emissor, audiência,
  `iat`, claims inválidos e invalidação por versão de sessão.

### P2. Corrigir recuperação de senha

- **Status:** Em andamento — invalidação e redução parcial de enumeração
  implementadas; a recuperação segura ainda depende de canal verificado.
- **Arquivos previstos:** `Backend/main.go`, `Backend/handlers/auth.go`,
  `Backend/middleware/ratelimit.go`, frontend de login.
- **Ações:**
  - [x] aplicar rate limit também à consulta da pergunta;
  - [x] uniformizar status e mensagem no endpoint de redefinição para usuário
    inexistente e resposta incorreta;
  - [ ] eliminar a exposição da pergunta e demais diferenças observáveis no
    fluxo completo;
  - [x] validar tamanho da nova senha no backend, respeitando o limite do bcrypt;
  - [ ] avaliar substituição da pergunta de segurança por código temporário/e-mail;
  - [x] invalidar sessões/tokens existentes após redefinição por `token_versao`;
- **Critérios de aceite:** não é possível enumerar usuários por respostas distintas;
  tentativas repetidas são limitadas; senha inválida é rejeitada pela API.
- **Evidência parcial:** a consulta da pergunta continua limitada por IP e a
  redefinição já não diferencia usuário inexistente de resposta incorreta. O
  cadastro/redefinição exigem senha de no mínimo 12 caracteres e até 72 bytes;
  login continua aceitando credenciais antigas. A tabela `usuarios` ganhou
  `token_versao`, incluída no JWT e conferida em cada requisição autenticada.
  Ainda falta substituir a pergunta por código temporário entregue por canal
  verificado; sem e-mail, SMS ou outro canal configurado não é seguro marcar
  essa vulnerabilidade como encerrada.

### P3. Tornar o rate limit confiável

- **Status:** Em andamento — proteção por IP e conta concluída; falta decidir
  armazenamento compartilhado e configurar o proxy em produção.
- **Arquivos previstos:** `Backend/middleware/ratelimit.go`, configuração de
  proxy/deploy se necessário.
- **Ações:**
  - [x] aceitar `CF-Connecting-IP` e `X-Forwarded-For` somente quando a
    conexão vier de uma rede em `TRUSTED_PROXY_CIDRS`;
  - [x] validar o formato do IP e usar o peer da conexão como fallback seguro;
  - [x] usar chave composta por IP + identificador normalizado da conta;
  - [x] limpar registros expirados quando o mapa atingir grande volume;
  - [ ] avaliar armazenamento compartilhado em produção, caso existam múltiplas instâncias;
  - [ ] configurar `TRUSTED_PROXY_CIDRS` com as redes reais do proxy/túnel;
- **Critérios de aceite:** cabeçalhos arbitrários enviados diretamente ao backend
  não permitem contornar o limite; o comportamento atrás do Cloudflare continua correto.
- **Evidência:** `obterIP` agora só usa cabeçalhos quando o peer pertence às
  redes CIDR configuradas; as tentativas também são indexadas pela conta
  normalizada, IPs inválidos são ignorados e o mapa possui limpeza de registros
  expirados. `go test ./...`, `go vet ./...` e `git diff --check` passaram em
  17/09/2026. Antes do próximo deploy, configurar
  `TRUSTED_PROXY_CIDRS` com a rede real do proxy/túnel; sem essa variável o
  sistema fica seguro contra falsificação, mas pode agrupar usuários pelo peer.

### P4. Proteger tokens no frontend

- **Status:** Em andamento — migração implementada; validação funcional no
  container/produção permanece pendente.
- **Arquivos previstos:** `Frontend/src/context/AuthContext.jsx`,
  `Frontend/src/api/api.js`, backend e nginx.
- **Ações:**
  - [x] emitir JWT em cookie `HttpOnly`, `SameSite=Lax` e `Secure` em produção;
  - [x] remover persistência do JWT no `localStorage` do frontend;
  - [x] restaurar sessão por endpoint autenticado `/sessao`;
  - [x] enviar credenciais em chamadas da API e habilitar CORS credentials;
  - [x] limpar cookie e invalidar a versão da sessão no logout;
  - [x] proteger métodos mutáveis autenticados por cookie com validação da origem;
  - [ ] validar login, reload, logout, expiração e PWA no build/container.
- **Critérios de aceite:** decisão arquitetural registrada e fluxo de login,
  logout, expiração e renovação coberto por testes.
- **Evidência parcial:** backend emite/valida cookie de sessão e frontend não
  grava mais JWT; a sessão é restaurada por `/sessao`. `go test ./...`,
  `go vet ./...` e `git diff --check` passaram em 17/09/2026. O build frontend
  passou em 17/09/2026; a proteção CSRF foi adicionada no middleware global para
  operações mutáveis autenticadas por cookie. A validação funcional no container
  ainda está pendente.

## Fase 2 — Integridade do estoque e validação de entrada

### P5. Validar dados no backend

- **Status:** Concluído — validação server-side, limites de autenticação e corpo
  estruturado verificados nesta etapa.
- **Arquivos previstos:** `Backend/handlers/itens.go`,
  `Backend/handlers/notas_fiscais.go`, modelos e frontend.
- **Ações:**
  - [x] rejeitar quantidade, retirada e estoque mínimo negativos, não finitos ou acima
    de limite definido;
  - [x] exigir nome, unidade e local com tamanho máximo;
  - [x] validar locais permitidos no backend;
  - [x] aplicar as regras também às entradas confirmadas de importação;
  - [x] validar tamanho dos campos de autenticação e exigir 12 caracteres em
    novas senhas/redefinições;
  - [x] limitar corpos estruturados a 1 MiB no middleware global;
  - [x] repetir o mínimo de senha no frontend apenas para melhor experiência,
    mantendo o backend como defesa principal.
- **Critérios de aceite:** chamadas diretas à API não conseguem criar estoque
  negativo, desabilitar alertas com mínimo negativo ou gravar valores inválidos.
- **Evidência:** adicionada validação centralizada em
  `Backend/handlers/validacao.go`, aplicada aos endpoints de itens, retirada e
  confirmação de nota. Os campos de autenticação têm limites próprios e os
  corpos estruturados são limitados pelo middleware global. `go test ./...` e
  `go vet ./...` passaram em 17/09/2026.

### P6. Tornar retirada atômica

- **Status:** Concluído.
- **Arquivos previstos:** `Backend/handlers/itens.go`.
- **Ações:**
  - [x] substituir o fluxo leitura-depois-atualização por operação atômica;
  - [x] manter o estoque em zero;
  - [x] verificar linhas afetadas e retornar 404 para item inexistente;
  - [x] consultar e devolver a quantidade atualizada após a operação.
- **Critérios de aceite:** retiradas concorrentes não perdem movimentações e nunca
  deixam o estoque abaixo de zero.
- **Evidência:** `RetirarItem` agora executa a subtração e o limite inferior na
  mesma instrução `UPDATE`. `go test ./...`, `go vet ./...` e `git diff --check`
  passaram em 17/09/2026.

### P7. Tornar confirmação de nota transacional e idempotente

- **Status:** Concluído — transação, idempotência e build frontend validados.
- **Arquivos previstos:** `Backend/handlers/notas_fiscais.go`, banco/modelos e frontend.
- **Ações:**
  - [x] validar toda a lista recebida, com limite de 500 itens;
  - [x] usar transação para atualizar/criar tudo ou nada;
  - [x] retornar erro e rollback em vez de ignorar falhas silenciosamente;
  - [x] usar chave única por usuário para impedir reprocessamento;
  - [x] manter isolamento por usuário em todas as operações;
  - [x] atualizar o frontend para gerar e reenviar a chave durante a confirmação.
- **Critérios de aceite:** falha em qualquer item gera rollback e resposta clara;
  reenvio da mesma confirmação não duplica entrada.
- **Evidência:** criada a tabela `confirmacoes_importacao`, a API passou a
  receber `{ chave, entradas }` e o modal gera uma chave por prévia. `go test
  ./...`, `go vet ./...` e `npm run build` passaram; o build foi repetido fora da
  restrição local do esbuild e já havia sido confirmado no CI.

## Fase 3 — Uploads, OCR e integrações externas

### P8. Limitar efetivamente uploads e processamento

- **Status:** Em andamento — limites de concorrência, corpo e dimensão concluídos;
  timeout e teto de processamento adicionados nesta etapa; imagens reais ainda
  precisam de validação funcional.
- **Arquivos previstos:** `Backend/main.go`,
  `Backend/handlers/notas_fiscais.go`, `Backend/services/ocr_nota.go` e
  `Backend/services/ocr_cloud.go`.
- **Ações:**
  - [x] aplicar `http.MaxBytesReader` no corpo das requisições multipart;
  - [x] validar o formato real e aceitar somente JPEG/PNG nos fluxos de imagem;
  - [x] limitar dimensões a 6000 px por lado e 12 milhões de pixels;
  - [x] limitar a duas leituras OCR simultâneas, rejeitando espera acima de 5 segundos;
  - [x] limitar a resposta do OCR.space a 2 MB;
  - [x] manter timeouts e mensagens sem expor detalhes internos.
  - [x] cancelar o Tesseract quando a requisição termina ou excede 40 segundos;
  - [x] reduzir a ampliação de imagens grandes para manter o processamento em
    no máximo 36 milhões de pixels.
- **Critérios de aceite:** corpo acima do limite é rejeitado cedo; imagem com
  dimensões abusivas não causa consumo excessivo; concorrência é controlada.
- **Evidência:** os três endpoints multipart usam `MaxBytesReader`; os fluxos
  de imagem passam por `ValidarImagem`; Tesseract e OCR.space compartilham um
  limite de concorrência; a resposta externa é limitada por `io.LimitReader`.
  O processamento local respeita cancelamento da requisição e limite total de
  40 segundos, com ampliação adaptativa para imagens grandes. `go test ./...`,
  `go vet ./...` e `git diff --check` passaram em 17/09/2026.

### P9. Remover debug sensível de produção

- **Status:** Em andamento — gravação está protegida por flag; diretório,
  retenção e limpeza do servidor ainda precisam de tratamento operacional.
- **Arquivos previstos:** `Backend/services/ocr_nota.go`,
  `Backend/services/nfce_scraper.go`, configuração e `.gitignore`.
- **Ações:**
  - [x] remover gravações automáticas por padrão;
  - [x] proteger a gravação por `DEBUG_OCR=true` explícito;
  - [x] usar permissão `0600` para os arquivos quando o debug for habilitado;
  - [ ] usar diretório temporário seguro e retenção controlada;
  - [ ] revisar/remover arquivos existentes no servidor.
- **Critérios de aceite:** produção não grava texto OCR, HTML de nota ou conteúdo
  fiscal sem configuração explícita e auditável.
- **Evidência:** `DebugOCRAtivo()` retorna falso por padrão e protege os dois
  pontos de gravação (`debug_ocr_ultimo_texto.txt` e
  `debug_nfce_ultima_consulta.html`). `go test ./...`, `go vet ./...` e
  `git diff --check` passaram em 17/09/2026.

### P10. Manter consulta QR Code segura antes de reativar

- **Status:** Concluído no código; cobertura automatizada específica ainda é
  parcial e a funcionalidade continua desativada.
- **Arquivos previstos:** `Backend/services/nfce_scraper.go` e handlers.
- **Ações:**
  - [x] exigir HTTPS e rejeitar userinfo, fragmentos e portas não permitidas;
  - [x] manter allowlist exata por hostname;
  - [x] limitar resposta HTML a 2 MB;
  - [x] bloquear redirecionamentos para hosts/esquemas não permitidos;
  - [x] não reativar scraping sem nova validação do estado real da SEFAZ;
  - [ ] ampliar os testes automatizados para redirects, IP privado e resposta
    acima do limite.
- **Critérios de aceite:** testes de SSRF cobrem host, esquema, redirecionamento,
  IP privado e respostas grandes.
- **Evidência:** validação agora aceita somente HTTPS, allowlist exata e porta
  443; redirecionamentos passam pela mesma validação e o corpo HTML é limitado.
  `go test ./...`, `go vet ./...` e `git diff --check` passaram em 17/09/2026.
  Ainda faltam testes automatizados específicos de SSRF, previstos na P13.

## Fase 4 — Headers, deploy e qualidade operacional

### P14. Exigir confirmação explícita para excluir produto

- **Status:** Concluído.
- **Arquivos previstos:** `Frontend/src/pages/Dashboard.jsx`,
  `Frontend/src/components/TabelaItens.jsx`,
  `Frontend/src/components/ModalConfirmacao.jsx` e estilos relacionados.
- **Ações:**
  - [x] abrir uma confirmação visível ao clicar em excluir;
  - [x] deixar claro o nome do produto e que a ação é permanente;
  - [x] exigir uma ação positiva separada, como botão “Excluir”;
  - [x] manter “Cancelar” como opção segura e não executar a exclusão ao fechar o modal;
  - [x] impedir cliques repetidos enquanto a exclusão estiver sendo processada;
  - [x] garantir foco inicial, foco de teclado, `Esc` e leitura adequada por leitores de tela.
- **Critérios de aceite:** clicar no ícone de exclusão não remove o produto
  imediatamente; somente a confirmação explícita chama o endpoint `DELETE`;
  cancelar, clicar fora ou pressionar `Esc` não exclui o produto; o fluxo funciona
  corretamente em telas pequenas e por teclado.
- **Evidência:** o fluxo já utilizava `ModalConfirmacao`; o componente foi
  reforçado para tratar `Escape` como cancelamento e desabilitar os dois botões
  durante a operação. O build frontend passou nesta etapa e no CI; a validação
  funcional no navegador/produção continua recomendada.

### P11. Adicionar headers de segurança no nginx

- **Status:** Em andamento — configuração e validação de build concluídas;
  validação de headers e fluxos no container/produção pendente.
- **Arquivos previstos:** `Frontend/nginx.conf`.
- **Ações:**
  - [x] configurar CSP compatível com React/Vite e PWA;
  - [x] configurar HSTS para o domínio HTTPS;
  - [x] configurar `X-Content-Type-Options`;
  - [x] configurar `Referrer-Policy`;
  - [x] configurar `Permissions-Policy` e `frame-ancestors`/`X-Frame-Options`;
  - [ ] validar headers e ausência de regressões no container em build/ambiente de produção.
- **Critérios de aceite:** headers aparecem em produção sem quebrar login, PWA,
  câmera ou chamadas à API.
- **Evidência parcial:** adicionados headers de segurança no bloco principal e
  na rota `/assets/` do nginx, evitando perda por herança de `add_header`.
  `git diff --check` passou em 17/09/2026. O `npm run build` do frontend passou;
  o nginx/Docker não está disponível localmente, portanto a sintaxe e a
  validação visual/funcional ainda devem ser confirmadas pelo container/CI.

### P12. Reforçar pipeline e reprodutibilidade

- **Status:** Em andamento — CI/deploy validado no workflow #17; a cobertura foi
  ampliada nesta etapa e precisa ser confirmada na próxima execução.
- **Arquivos previstos:** `.github/workflows/deploy.yml`, Dockerfiles.
- **Ações:**
  - [x] executar `go test ./...` e `go vet ./...` no CI;
  - [x] executar `npm audit --omit=dev --audit-level=high` no CI;
  - [x] executar lint do frontend no CI;
  - [x] atualizar `react-router-dom` para `7.18.2` e repetir a auditoria;
  - [x] usar `npm ci` no Dockerfile e no CI;
  - [x] fixar actions por SHA após revisão das versões;
  - [x] fixar o runner em `ubuntu-24.04` para evitar migração automática do
    `ubuntu-latest`;
  - [x] avaliar `appleboy/ssh-action` e identificar incompatibilidade persistente
    de fingerprint;
  - [x] substituir a action por SSH nativo com `ssh-keyscan`, comparação de
    fingerprint e `StrictHostKeyChecking=yes`;
  - [x] tornar pull previsível com `git pull --ff-only origin main`;
  - [x] impedir deploys concorrentes com `concurrency`;
  - [x] adicionar smoke tests HTTP para backend e frontend;
  - [x] repetir qualquer falha de conexão/resposta durante a inicialização,
    mantendo limite de tentativas e falha definitiva ao esgotá-lo;
  - [x] exigir início manual do workflow e ambiente `production` antes do deploy;
- **Critérios de aceite:** CI detecta regressões; build usa lockfile; deploy falho
  não é apresentado como concluído.
- **Evidência parcial:** além das proteções já registradas, o workflow agora
  executa testes Go, auditoria de dependências e lint antes do deploy. A versão
  vulnerável do React Router foi atualizada para `7.18.2`; `npm audit
  --omit=dev` retornou zero vulnerabilidades. O próximo workflow deve confirmar
  a nova sequência integral.

#### P12-A. Acesso privado da VPS via Tailscale

- **Status:** Concluído com pendência operacional — OAuth, acesso privado e
  deploy foram validados; a credencial legada ainda deve ser removida/revogada.
- **Constatação:** o IP informado (`100.67.151.30`) é um endereço Tailscale.
  O workflow anterior usava runner GitHub hospedado e SSH direto, sem conectar
  o runner ao tailnet; por isso não funcionaria com a porta 22 pública fechada.
- **Ações:**
  - [x] conectar o runner temporariamente à Tailscale com
    `tailscale/github-action@v4`;
  - [x] usar `VPS_HOST` como host Tailscale, sem gravar o IP no código;
  - [x] usar secrets separados para `VPS_USER` e `VPS_SSH_KEY`;
  - [x] validar fingerprint SHA-256 com `VPS_HOST_FINGERPRINT`;
  - [x] manter SSH na porta 22 privada;
  - [x] criar/configurar `TAILSCALE_AUTHKEY`, `VPS_HOST`, `VPS_USER`,
    `VPS_SSH_KEY` e `VPS_HOST_FINGERPRINT` no ambiente protegido;
  - [x] criar a tag `tag:github-actions` com `autogroup:admin` como proprietário;
  - [x] criar credencial OAuth exclusiva para o deploy, com apenas
    `auth_keys: Write` e tag `tag:github-actions`;
  - [x] cadastrar `TAILSCALE_OAUTH_CLIENT_ID` e `TAILSCALE_OAUTH_SECRET` no
    ambiente `production`;
  - [x] migrar o workflow de `authkey` para OAuth client com a tag exclusiva do
    runner, mantendo o acesso privado à VPS pela porta 22;
  - [x] executar o workflow migrado e confirmar Tailscale, SSH privado,
    fingerprint e smoke tests;
  - [x] usar auth key reutilizável e efêmera, adequada a runners descartáveis do GitHub Actions;
  - [x] executar workflow e confirmar smoke tests sem abrir a porta 22 pública;
  - [x] trocar o disparo automático por `workflow_dispatch`;
  - [x] usar o ambiente `production` para permitir regras de aprovação;
  - [x] configurar aprovação obrigatória no ambiente do GitHub;
  - [ ] remover o secret legado `TAILSCALE_AUTHKEY` depois de confirmar que não
    há workflow ativo que ainda o utiliza;
  - [ ] revogar a auth key legada no console Tailscale.
- **Critérios de aceite:** o workflow conecta ao tailnet, acessa a VPS pelo
  endereço Tailscale, valida a host key e executa o smoke test; uma execução
  fora do tailnet não alcança a VPS.
- **Atenção de segurança:** foi localizada uma anotação local fora deste
  projeto contendo credenciais junto das instruções de SSH. Os valores não
  foram reproduzidos nem adicionados ao repositório. Rotacionar os segredos
  expostos e remover cópias inseguras após confirmar que não são necessárias.

### P13. Criar suíte de testes

- **Status:** Em andamento — cobertura crítica inicial criada.
- **Escopo mínimo:**
  - [x] autenticação e claims JWT;
  - [x] isolamento entre usuários;
  - [x] validação de quantidades e locais;
  - [x] retirada concorrente;
  - [x] transação/idempotência de nota;
  - [x] limite/validação de URLs SSRF;
  - [x] limite de dimensão de imagens OCR;
  - [x] limite de uploads HTTP;
  - [x] build do frontend;
  - [x] lint do frontend com script `npm run lint`.
- **Critérios de aceite:** testes rodam localmente e no CI, com casos de falha
  reproduzindo os riscos listados neste documento.
- **Evidência parcial:** criados testes em `Backend/config`,
  `Backend/handlers`, `Backend/middleware` e `Backend/services` cobrindo segredo
  JWT, claims inválidos, algoritmo não permitido, invalidação de token após
  troca de senha, dados de estoque, importação transacional/idempotente,
  isolamento entre usuários e allowlist HTTPS. `go test ./...`,
  `go vet ./...` e `git diff --check` passaram em 17/09/2026. O frontend agora
  possui lint executado localmente e no CI; ainda não há testes de
  componentes/fluxos no navegador. O build frontend passou em 17/09/2026, com
  alerta de bundle JavaScript acima de 500 kB.

### P15. Endurecer container e superfície de rede

- **Status:** Pendente — requer validação do ambiente de produção antes de
  alterar permissões, volumes ou exposição de portas.
- **Constatações:** o backend ainda é iniciado como root no `Backend.Dockerfile`
  e o `docker-compose.yml` publica `8090:8080` em todas as interfaces. Como a
  publicação atual é consumida por túnel/proxy e usa bind mount em `./data`,
  trocar o usuário ou a interface sem conferir a VPS pode interromper banco e
  acesso externo.
- **Ações:**
  - [ ] executar o backend com usuário não privilegiado, validando a posse/permissão
    do volume persistente;
  - [ ] restringir as portas publicadas a loopback ou comprovar firewall/túnel
    equivalente na VPS;
  - [ ] registrar a configuração efetiva de firewall, Cloudflare Tunnel e
    `TRUSTED_PROXY_CIDRS`.

## Ordem para encerramento

1. Executar o workflow atualizado e confirmar `go test`, auditoria, lint, build,
   deploy privado e smoke tests.
2. Validar no container/produção os headers, login, reload, logout, PWA e OCR
   com imagens reais.
3. Configurar `TRUSTED_PROXY_CIDRS` e decidir rate limit por conta/armazenamento
   compartilhado se houver mais de uma instância.
4. Revisar arquivos de debug na VPS e removê-los de forma segura, mantendo
   `DEBUG_OCR` desativado em produção.
5. Validar usuário não privilegiado e exposição de portas do Docker com o
   Cloudflare Tunnel/firewall efetivos.
6. Substituir a recuperação por pergunta por código temporário entregue por
   canal verificado.
7. Depois de confirmar que nenhum workflow usa a credencial antiga, remover o
   secret `TAILSCALE_AUTHKEY` e revogar a auth key legada no Tailscale.

## Pendências que impedem declarar o projeto encerrado

- recuperação de senha ainda usa pergunta de segurança e não possui canal
  verificado para entrega de código de uso único;
- rate limit ainda é local ao processo, embora já combine IP e conta;
- configuração efetiva de `TRUSTED_PROXY_CIDRS`, firewall e túnel não está
  comprovada neste checkout;
- usuário não-root do backend e bind de portas precisam ser validados com os
  volumes e o túnel reais;
- arquivos de debug e fluxos de headers/PWA/OCR precisam de verificação na VPS;
- limpeza da credencial Tailscale legada ainda requer confirmação operacional;
- não há testes automatizados de componentes ou navegador no frontend.

## Registro de progresso

| Data | Item | Alteração/resultado | Evidência |
|---|---|---|---|
| 17/09/2026 | Análise inicial | Riscos e melhorias catalogados; nenhum código alterado | `git status`, leitura do projeto |
| 17/09/2026 | Documento | Plano criado no projeto | Este arquivo |
| 17/09/2026 | P1 | JWT endurecido; segredo vazio/curto bloqueia inicialização e claims inválidos retornam 401 | `go test ./...`, `go vet ./...` |
| 17/09/2026 | P5 | Validação server-side de números, limites, textos, locais e importação | `go test ./...`, `go vet ./...` |
| 17/09/2026 | P6 | Retirada de estoque convertida para atualização atômica | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P7 | Confirmação de nota transacional, com rollback e idempotência por usuário | `go test ./...`, `go vet ./...`; build frontend bloqueado pelo ambiente |
| 17/09/2026 | P2 | Rate limit aplicado à pergunta e validação de senha adicionada no backend | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P3 | Rate limit deixou de confiar em cabeçalhos de IP não verificados | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P8 | Uploads e processamento OCR passaram a ter limites de corpo, imagem, concorrência e resposta | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P9 | Debug de OCR/NFC-e desativado por padrão e protegido por flag | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P10 | QR Code mantido desativado e scraper protegido contra esquemas, redirects e respostas grandes | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P11 | Headers de segurança adicionados ao nginx principal e aos assets | `git diff --check`; validação do container pendente |
| 17/09/2026 | P12 | CI/deploy usa lockfile, pull previsível, concorrência controlada e smoke tests | `git diff --check`; execução do workflow pendente |
| 17/09/2026 | P12-A | Workflow passou a conectar runner à Tailscale e validar fingerprint antes do SSH | `.github/workflows/deploy.yml`; secrets/execução pendentes |
| 17/09/2026 | P13 | Testes unitários iniciais para JWT, validação de estoque e SSRF/allowlist | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P2 | Tokens antigos passam a ser invalidados após redefinição por `token_versao` | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P13 | Teste de integração confirma rejeição de token após incremento da versão de sessão | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P14 | Confirmação de exclusão reforçada com cancelamento por Esc e bloqueio contra duplo clique | Revisão do fluxo; build frontend pendente |
| 17/09/2026 | P13 | Testes de importação cobrem idempotência, rollback e isolamento entre usuários | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P13 | Testes cobrem limite de dimensão para imagens OCR | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P13 | Teste concorrente cobre duas retiradas sem perda ou estoque negativo | `go test ./... -count=5`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P13 | Teste confirma rejeição de multipart acima de 10 MB | `go test ./... -count=3`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P4 | Sessão migrada para cookie HttpOnly e restauração por `/sessao` | `go test ./...`, `go vet ./...`, `git diff --check`; frontend pendente |
| 17/09/2026 | P4 | Proteção de origem adicionada para operações mutáveis com cookie de sessão | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P4/P11/P13 | Build de produção do frontend passou; Docker/nginx permanece pendente por daemon indisponível | `npm run build` |
| 17/09/2026 | P12 | Deploy alterado para acionamento manual com ambiente `production` | `.github/workflows/deploy.yml`; execução pendente |
| 17/09/2026 | P12-A | Execução #5: build passou e Tailscale conectou, mas SSH recusou o host por divergência de fingerprint | GitHub Actions #5; `ssh: handshake failed: ssh: host key fingerprint mismatch` |
| 17/09/2026 | P12-A | Nova execução: build passou, mas Tailscale recusou a credencial antes do SSH | GitHub Actions; `backend error: invalid key` |
| 17/09/2026 | P12-A | Causa provável identificada: auth key não reutilizável foi invalidada após o uso pelo runner descartável | Tela do Tailscale indica auth key recentemente invalidada |
| 17/09/2026 | P12-A | Execução seguinte: autenticação Tailscale passou, mas o SSH ainda recusou o fingerprint cadastrado | GitHub Actions; `ssh: handshake failed: ssh: host key fingerprint mismatch` |
| 17/09/2026 | P12-A | Verificação local confirmou a chave ED25519 apresentada pelo IP Tailscale | SSH local sem autenticação; fingerprint `SHA256:H70y+CElKTME1FJYmGNxXQGILhSBDLycqGbGqvpZZbk` |
| 17/09/2026 | P12 | Revisão da execução #9 confirmou ambiente, secrets e Tailscale; falha permanece na validação do fingerprint pelo `appleboy/ssh-action@v1.0.3` | GitHub Actions #9; issue oficial #275 da action relata o mesmo erro |
| 17/09/2026 | P12 | Action SSH atualizada para v1.2.5 e fixada no commit `0ff4204d59e8e51228ff73bce53f80d53301dee2` | Release oficial v1.2.5; `git diff --check` passou |
| 17/09/2026 | P12 | Execução automática após o commit revelou erro de indentação YAML na linha 63; deploy não chegou a iniciar | GitHub Actions #11; workflow inválido |
| 17/09/2026 | P12 | Execução #12: YAML e build passaram, mas a action v1.2.5 ainda recusou o fingerprint; `script_stop` também foi identificado como input não suportado | GitHub Actions #12; `host key fingerprint mismatch` e aviso de `script_stop` inesperado |
| 17/09/2026 | P12 | Deploy migrado para SSH nativo; fingerprint ED25519 é comparado antes da autenticação e o `known_hosts` temporário é usado com verificação estrita | `.github/workflows/deploy.yml`; execução #13 confirmou SSH, Tailscale e rebuild |
| 17/09/2026 | P12 | Execução #13 confirmou a conexão privada e a recriação dos containers, mas o primeiro probe recebeu `curl (52) Empty reply from server` durante a inicialização | GitHub Actions #13; ajuste de prontidão pendente |
| 17/09/2026 | P12 | Smoke tests ajustados para repetir falhas de conexão/resposta por até 20 tentativas antes de falhar | `.github/workflows/deploy.yml`; `git diff --check` pendente |
| 17/09/2026 | P12-A | Auth key reutilizável e efêmera configurada; aprovação obrigatória do ambiente `production` habilitada | Configuração do Tailscale e GitHub Environment |
| 17/09/2026 | P12 | Execução #14 passou com build, Tailscale, SSH nativo, fingerprint, rebuild Docker e smoke tests HTTP | GitHub Actions #14; status `Success` em 1m37s |
| 17/09/2026 | P12 | O retry explícito dos smoke tests absorveu a janela de inicialização sem mascarar falha definitiva | GitHub Actions #14; job `deploy` concluído em 12s |
| 17/09/2026 | P12 | Permanecem dois avisos não bloqueantes no CI: actions forçadas para Node.js 24 e cache Go sem `go.sum` na raiz | GitHub Actions #14; melhoria futura recomendada |
| 17/09/2026 | P12 | Actions atualizadas e fixadas por SHA: checkout v7.0.1, setup-go v7.0.0, setup-node v7.0.0 e Tailscale v4.1.3; cache Go apontado para `Backend/go.sum` | `.github/workflows/deploy.yml`; validação no próximo CI |
| 17/09/2026 | P12 | Execução #15 confirmou build e deploy após as actions fixadas; avisos de Node.js e cache Go não reapareceram | GitHub Actions #15; `build-and-test` 58s e `deploy` 17s |
| 17/09/2026 | P12 | Runner fixado em `ubuntu-24.04`; o aviso de migração automática do `ubuntu-latest` será eliminado na próxima execução | `.github/workflows/deploy.yml`; validação no próximo CI |
| 17/09/2026 | P12-A | Execução #15 confirmou Tailscale, SSH privado, fingerprint e smoke tests; permanece o aviso de depreciação do parâmetro `authkey` | GitHub Actions #15; status `Success` em 1m55s |
| 17/09/2026 | P12 | Execução #16 confirmou o runner `ubuntu-24.04`; o aviso de migração do `ubuntu-latest` não reapareceu | GitHub Actions #16; status `Success` em 1m19s |
| 17/09/2026 | P12-A | Execução #16 confirmou novamente Tailscale, SSH privado, fingerprint e smoke tests; permanece somente o aviso de depreciação do `authkey` | GitHub Actions #16; job `deploy` concluído em 15s |
| 17/09/2026 | P12-A | Tag `tag:github-actions` criada; credencial OAuth perdida revogada e nova credencial gerada com escopo mínimo `auth_keys: Write`; segredos não registrados no projeto | Console Tailscale; cadastro dos secrets e validação do workflow pendentes |
| 17/09/2026 | P12-A | Secrets `TAILSCALE_OAUTH_CLIENT_ID` e `TAILSCALE_OAUTH_SECRET` cadastrados no ambiente `production`; workflow migrado para OAuth com a tag exclusiva | GitHub Environment e `.github/workflows/deploy.yml`; execução de validação pendente |
| 17/09/2026 | P12-A | Execução #17 passou após a migração para OAuth; `build-and-test` em 28s, `deploy` em 21s e duração total de 1m44s | [GitHub Actions #17](https://github.com/Saulorangel87/App-controle-de-estoque/actions/runs/35286690984); status `Success` |
| 17/09/2026 | P1 | JWT passou a exigir emissor, audiência, `iat` e `exp` válidos; login de contas antigas foi preservado | `go test ./...`, `go vet ./...` |
| 17/09/2026 | P5 | Limites de nome/pergunta/resposta, senha mínima de 12 caracteres para novas credenciais e corpo estruturado de 1 MiB adicionados | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P8 | Tesseract passou a respeitar cancelamento da requisição e teto de 40s; ampliação de imagens grandes foi reduzida | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P12/P13 | React Router atualizado para `7.18.2`; lint, auditoria de produção e testes Go adicionados ao workflow | `npm run lint`, `npm run build`, `npm audit --omit=dev`, `git diff --check` |
| 17/09/2026 | P2/P12-A/P15 | Documento corrigido com pendências reais: canal verificado de recuperação, limpeza da credencial Tailscale legada e hardening de container/rede | Revisão do código, workflow e configuração versionada |

## Registro de decisões e riscos aceitos

Nenhum risco foi formalmente aceito até o momento.
