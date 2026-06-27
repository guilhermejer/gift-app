# Copilot Instructions — gift-app

## Visão Geral do Projeto

Aplicação para sugestões de presentes a amigos e familiares. O usuário, com auxílio de uma LLM, constrói perfis detalhados de pessoas próximas. A LLM questiona o usuário sobre gostos, idade, localização, cultura, religião, histórico de conversas, etc. O app envia lembretes de aniversário com sugestões de presentes personalizadas.

### Stack

| Camada       | Tecnologia                                      |
|--------------|-------------------------------------------------|
| Back-end API | Go (REST)                                       |
| LLM API      | Python (LangChain / LangGraph) — serviço separado |
| Banco        | PostgreSQL + extensão vetorial (pgvector) para RAG |
| Front-end    | React (Web)                                     |

---

## Diretrizes de Desenvolvimento

### Geral

- Mantenha a solução **simples e funcional**; evite over-engineering.
- Não adicione abstrações, helpers ou features além do que foi explicitamente pedido.
- Prefira clareza de código à esperteza.
- Não adicione comentários óbvios; comente apenas lógica não evidente.

### Linguagem — Go (back-end REST)

- Use Go idiomático: siga as convenções de nomenclatura, tratamento de erro e organização de pacotes da linguagem.
- Não use frameworks pesados; prefira `net/http` da stdlib ou roteadores leves (ex: `chi`).
- Retorne erros explicitamente; nunca ignore `err`.
- Use `context.Context` para propagação de cancelamento e prazos.

### Arquitetura Hexagonal

Organize o código em três camadas:

```
internal/
  domain/      # Entidades, regras de negócio, interfaces (ports)
  port/        # Interfaces de entrada (driven) e saída (driving)
  adapter/     # Implementações concretas: HTTP handlers, repositórios, clientes externos
```

- **Domain**: sem dependências de infraestrutura. Contém entidades (`Friend`, `User`, `Profile`, `Gift`, `Reminder`, etc.) e interfaces de repositório/serviço.
- **Port**: define contratos (interfaces Go) que o domínio expõe ou consome.
- **Adapter**: implementa os ports — handlers HTTP, repositórios Postgres, cliente da LLM API Python.

### Testes

- Escreva testes unitários **apenas para funcionalidades core** do domínio (lógica de negócio).
- Use a stdlib `testing`; mocks manuais ou `testify/mock` são aceitáveis.
- Não crie testes para adapters sem uma razão clara (prefira testes de integração quando necessário).
- Nomeie os testes seguindo o padrão `TestNomeDaFunção_Cenário`.

### Banco de Dados — PostgreSQL + pgvector

- Use `pgvector` para armazenar embeddings de perfis (RAG).
- Migrations versionadas (ex: com `golang-migrate`).
- Queries em SQL puro ou `sqlx`; evite ORMs pesados.

### Integração com LLM (Python API)

- O serviço Go consome a LLM API Python via HTTP/REST.
- Defina um port (interface) para o cliente LLM no domínio; o adapter concreto faz as chamadas HTTP.
- A LLM usa **tools** para buscar tendências e refinar preferências.

### Segurança

- Valide e sanitize todas as entradas na borda (handlers HTTP).
- Não exponha detalhes internos de erro ao cliente; logue server-side.
- Use variáveis de ambiente para segredos (nunca hardcode).
- Aplique autenticação nas rotas que exigem identidade do usuário.
