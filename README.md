```markdown
# Go Auth Boilerplate

Um boilerplate de autenticação genérico, robusto e performático desenvolvido em Go. Este projeto foi estruturado para servir como base reutilizável em qualquer novo ecossistema que necessite de controle de acesso seguro.

---

## 🚀 Funcionalidades

* **Autenticação Baseada em Sessão**: Armazenamento de sessões ativas no Redis vinculando o token ao ID do usuário com TTL (Time-To-Live) automático de 1 hora.
* **Segurança Avançada**:
  * Criptografia de senhas utilizando `bcrypt` com custo padrão (`DefaultCost`).
  * Tratamento estrito contra *User Enumeration* (mensagens de erro genéricas como "invalid credentials" para o cliente).
  * Cookies de sessão configurados com `HttpOnly`, `SameSite=Lax` e `Secure` dinâmico (via variável de ambiente).
* **Proteção contra Abuso (Rate Limiting)**: Middleware integrado por IP e rota nos endpoints públicos utilizando Redis (configurado para 5 tentativas por minuto).
* **Arquitetura Desacoplada**: Totalmente baseado em Interfaces (`UserRepository`), facilitando a substituição da camada de dados ou testes automatizados.
* **Gerenciamento de Contexto**: Propagação nativa de `context.Context` vindo da requisição HTTP até as camadas de serviço e banco de dados.

---

## 🛠️ Stack Tecnológica

* **Linguagem**: Go 1.23.4
* **Web Framework**: Gin Gonic
* **Banco de Dados**: PostgreSQL 16 (Persistência de Usuários)
* **Cache/Sessões**: Redis 7 (Tokens, Controle de Sessão e Rate Limiting)
* **Ambiente**: Docker & Docker Compose

---

## 📁 Estrutura do Projeto

O projeto segue uma estrutura organizada e limpa:

```text
├── cmd/
│   └── api/
│       └── main.go               # Ponto de entrada da aplicação e Injeção de Dependências
├── internal/
│   ├── config/
│   │   ├── config.go             # Gerenciamento e parsing de variáveis de ambiente (.env)
│   │   └── init_db.go            # Inicialização e pool de conexões (Postgres/Redis)
│   ├── handler/
│   │   └── user_handler.go       # Camada de transporte (HTTP), validação de formulários e cookies
│   ├── middleware/
│   │   ├── auth.go               # Middleware de proteção de rotas (RequireAuth)
│   │   └── rate_limiter.go       # Middleware de limitação de taxa baseado em Redis
│   ├── repository/
│   │   ├── user_repository_interface.go # Interface/Contrato da camada de dados
│   │   └── user_repository.go    # Implementação concreta usando Postgres + Redis (TxPipeline)
│   └── service/
│       └── user_service.go       # Regras de negócio da aplicação
├── pkg/
│   └── utils/
│       └── utils.go              # Utilitários de hash (bcrypt) e geração de tokens criptográficos
├── docker-compose.yml            # Infraestrutura local (Postgres e Redis)
└── go.mod                        # Dependências do projeto

```

---

## ⚙️ Configuração e Instalação

### Pré-requisitos

* Docker e Docker Compose instalados.
* Go 1.23+ instalado localmente.

### 1. Clonar o repositório

```bash
git clone [https://github.com/seu-usuario/go-auth.git](https://github.com/seu-usuario/go-auth.git)
cd go-auth

```

### 2. Configurar as Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto com base nas chaves lidas pelo sistema:

```env
PORT=:8080
COOKIE_SECURE=false

# PostgreSQL
DB_ADDR=postgres://admin:admin@localhost:5432/auth-go?sslmode=disable
DB_MAX_OPEN_CONNS=30
DB_MAX_IDLE_CONNS=30
DB_MAX_CONN_LIFETIME=15m

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### 3. Subir a Infraestrutura (Postgres & Redis)

```bash
docker compose up -d
```

### 4. Executar a Aplicação

```bash
go run cmd/api/main.go
```

---

## 🔗 Endpoints da API

A API responde sob o prefixo `/api/v1`.

| Método | Endpoint | Protegido? | Descrição | Payload (Form Data) |
| --- | --- | --- | --- | --- |
| **POST** | `/api/v1/register` | Não | Registra um novo usuário (mínimo 8 caracteres) | `username`, `password`<br> |
| **POST** | `/api/v1/login` | Não | Autentica e define o cookie `session_token` | `username`, `password`<br> |
| **GET** | `/api/v1/logout` | **Sim** | Invalida a sessão no Redis e expira o cookie | Envia o Cookie `session_token`<br> |

### Exemplo de Resposta de Rate Limit (429)

Quando o limite de requisições por IP e rota é excedido, a API retorna o header `Retry-After` com o tempo restante em segundos:

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 45
Content-Type: application/json

{
  "error": "too many requests, try again later"
}

```
---
