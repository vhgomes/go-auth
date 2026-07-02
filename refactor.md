# Go Auth — Checklist de Refatoração

> **Observação:** o objetivo declarado dessa refatoração não é só corrigir bugs — é deixar o projeto genérico o suficiente pra funcionar como **boilerplate de autenticação**: copiar a pasta pra um projeto novo e só trocar o que for específico do domínio (nome do módulo, campos extras de usuário, se necessário). Por isso incluí uma seção extra (🔧) só com os pontos que travam esse objetivo — eles são tão importantes quanto os bugs de segurança, porque um boilerplate com acoplamento forte não é reaproveitável, é só um projeto a mais pra manter.

Ordem sugerida: `[CRÍTICO]` primeiro (bugs de auth ativos), depois 🔧 (generalização), depois o resto.

---

## 🔴 CRÍTICO — Bugs ativos / segurança

- [x] **Erro do `HashPassword` ignorado** (`internal/repository/user_repository.go`)
  `hashedPassword, _ := utils.HashPassword(password)` — se o hash falhar (ex: senha > 72 bytes, limite do bcrypt), uma string vazia é inserida como senha do usuário.
  **Fix:** tratar o erro e abortar o registro se o hash falhar.

- [x] **`bcrypt.MinCost` em produção** (`pkg/utils/utils.go`)
  Custo 4 — mínimo absoluto da lib, pensado pra testes, oferece pouquíssima resistência a brute-force se o hash vazar.
  **Fix:** `bcrypt.DefaultCost` (10) como piso.

- [x] **Sessão no Redis nunca expira de fato** (`internal/repository/user_repository.go`)
  `Expire(ctx, sessionToken, ...)` seta TTL numa key de topo que não existe — o token está armazenado como *field* dentro do hash `"user_sessions"`, não como key própria. O `Expire` não tem efeito nenhum.
  **Fix:** trocar o hash monolítico por uma chave por sessão (`session:<token>` → `userID`) com `SET` + TTL nativo (`SETEX`), ou usar `HEXPIRE` se a versão do Redis suportar.

- [x] **`GetTokenByUserId` varre todas as sessões do sistema a cada login**
  `HScan` sobre o hash inteiro pra achar o token de um único usuário — O(n) no total de sessões ativas, não O(1) por usuário.
  **Fix:** resolvido junto com o ponto acima — chave por sessão elimina o scan.

- [x] **Falhas de inicialização em `main.go` não interrompem o boot**
  Erros de `InitPostgres` e `InitRedis` só são impressos com `fmt.Printf`, nunca abortam. Isso leva a nil pointer dereference em `InitDB(db)` quando a conexão falha.
  **Fix:** todo erro nessa cadeia vira `log.Fatalf` imediato.

---

## 🔧 Generalização — pré-requisito pra virar boilerplate reaproveitável

- [x] **Nenhum middleware de autenticação existe**
  O projeto só tem os três endpoints de auth (`register`, `login`, `logout`) — não há nada que valide o `session_token` em rotas protegidas. Todo projeto novo que reutilizar isso vai precisar reescrever esse middleware do zero, que é justamente a parte mais repetitiva entre projetos.
  **Fix:** criar `middleware/auth.go` com uma função `RequireAuth(userRepo)` que lê o cookie, valida a sessão no Redis, injeta `userID` no contexto do Gin e retorna 401 se inválido/expirado. Esse é o componente que mais justifica um boilerplate existir.

- [x] **`UserRepository` é um struct concreto, não uma interface**
  `UserService` depende diretamente de `*repository.UserRepository`. Em um projeto novo com requisitos diferentes de storage (ex: outro banco, ou Redis Cluster), não dá pra trocar a implementação sem editar o service.
  **Fix:** extrair uma interface (`UserRepository` com os métodos `RegisterUser`, `LoginUser`, `LogoutUser`) no pacote `service`, e fazer `repository.UserRepository` satisfazer essa interface. `NewUserService` passa a receber a interface, não o struct concreto.

---

## 🟠 IMPORTANTE — Robustez, segurança secundária

- [x] **Race condition em `RegisterUser`** (check-then-insert)
  `SELECT COUNT(*)` seguido de `INSERT`, sem atomicidade — a tabela já tem `UNIQUE` em `username`.
  **Fix:** inserir direto e tratar erro de constraint violation (`23505` no `lib/pq`) como "usuário já existe", removendo o check prévio.

- [x] **User enumeration + vazamento de erro interno pro cliente**
  `LoginUser` retorna `"username does not exist"` vs `"invalid password"` como erros distintos, e handlers fazem `c.JSON(..., gin.H{"error": err.Error()})` — devolvendo erro interno (inclusive de banco) direto pro cliente.
  **Fix:** mensagem genérica ("credenciais inválidas") pro cliente; log detalhado só no servidor.

- [x] **`context.Background()` em vez do contexto da requisição**
  `LoginUser`/`LogoutUser` criam contexto do zero no repository; `UserService` nem recebe `ctx` como parâmetro.
  **Fix:** `ctx := c.Request.Context()` no handler, propagado por parâmetro até as chamadas `...Context(ctx, ...)`.

- [x] **Verificação de método HTTP redundante nos handlers**
  `if c.Request.Method != http.MethodPost` é inalcançável — a rota já é registrada com método fixo no `main.go`. Também há uma função inteira comentada (`GetTokenBySession`) sobrando no repository.
  **Fix:** remover os dois.

- [x] **Cookie de sessão sem `Secure` e sem `SameSite`**
  `Secure: false` hardcoded em `SetCookie` — cookie aceito em HTTP puro, sem `SameSite` explícito.
  **Fix:** `Secure` configurável via env (true em produção); definir `SameSite` explicitamente.

---

## 🟡 MELHORIA — Qualidade, manutenibilidade

- [x] **`LoginUser` responde `201 Created`** — deveria ser `200 OK` (não cria recurso).
- [x] **`panic` dentro de `InitRedis`** (`internal/config/init_db.go`) — fora de `main()`; trocar por retorno de erro tratado com `log.Fatalf` no chamador.
- [x] **Sem rate limiting em `/register` e `/login`** — nenhuma proteção contra brute-force.
- [x] **Imagens `:latest` no `docker-compose.yml`** — pinar versões (`postgres:16-alpine`, `redis:7-alpine`).

---

## Ordem de ataque sugerida

1. CRÍTICOs de auth — hash ignorado e sessão que nunca expira são os que mais pesam, porque comprometem a garantia básica que um serviço de auth precisa dar.
2. Boot do `main.go` (falha silenciosa → crash confuso).
3. Seção 🔧 inteira — sem isso, mesmo com os bugs corrigidos, o projeto continua sendo "um app de auth" e não "um boilerplate de auth". O middleware de autenticação e a interface de repository são os dois itens que mais desbloqueiam reuso.
4. IMPORTANTE de higiene (context, enumeration, cookie).
5. MELHORIA conforme tempo disponível.
