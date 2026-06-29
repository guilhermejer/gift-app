# Guia de Integracao - Backend consumindo Gift LLM API

## 1. Objetivo
Este documento descreve como integrar o backend da aplicacao com a API Python do projeto Gift LLM.

Escopo deste guia:
- Como chamar os endpoints da API Python
- Ordem recomendada dos fluxos
- Contratos de request/response
- Tratamento de erros
- Recomendacoes para producao

---

## 2. URL base e configuracao
Base URL padrao da API Python:
- `http://localhost:8000`

Variaveis recomendadas no backend consumidor:
- `GIFT_LLM_API_BASE_URL` (ex: `http://localhost:8000`)
- `GIFT_LLM_API_TIMEOUT_SECONDS` (ex: `20`)

Headers padrao:
- `Content-Type: application/json`
- `Accept: application/json`
- `X-Request-Id: <uuid>` (recomendado para correlacao)

---

## 3. Endpoints disponiveis para o backend

### 3.1 Saude
`GET /health`

Uso:
- Health check de startup
- Readiness check do servico Python

Resposta esperada:
```json
{
  "status": "ok",
  "service": "Gift LLM API",
  "env": "dev",
  "model": "gpt-5.4-nano"
}
```

### 3.2 Profiles
Fluxo conversacional (novo modelo recomendado):
- `POST /profiles/agent/chat`
- `POST /profiles/agent/finalize`
- `DELETE /profiles/agent/session/{session_id}`

Fluxo direto (legado/suporte):
`POST /profiles`
`GET /profiles/{friend_id}`
`PATCH /profiles/{friend_id}`
`POST /profiles/{friend_id}/embedding`

Regra importante:
- `friend_id` deve ser sempre um UUID valido existente no backend principal.

### 3.3 Sugestoes
`POST /profiles/{friend_id}/suggestions`
`GET /profiles/{friend_id}/suggestions`

Internamente, essas operacoes sao mapeadas para gifts no backend de dados.

### 3.4 Eventos
`POST /profiles/{friend_id}/events`
`GET /profiles/{friend_id}/events?user_id=...`
`GET /events/upcoming?user_id=...&days=...`

Internamente, essas operacoes sao mapeadas para reminders no backend de dados.

---

## 4. Fluxo recomendado de uso

Fluxo padrao para cada amigo (`friend_id`):
1. Backend inicia uma sessao de conversa (`session_id`) para o `friend_id`
2. Backend envia cada mensagem do usuario para `POST /profiles/agent/chat`
3. API retorna a proxima pergunta/resposta do assistente
4. Quando a conversa estiver suficiente, backend chama `POST /profiles/agent/finalize`
5. A finalizacao persiste profile e embedding automaticamente (usando todo historico da conversa)
6. Opcionalmente, backend chama `DELETE /profiles/agent/session/{session_id}` para liberar cache
7. Depois disso, fluxo de sugestoes e reminders segue normal

---

## 5. Contratos de payload (copiar e usar)

### 5.1 Conversa com profile agent (novo modelo)
`POST /profiles/agent/chat`

```json
{
  "session_id": "session-001",
  "friend_id": "c096d057-a0cf-4c2d-857b-41beccb42de8",
  "message": "Meu amigo gosta de livros e cafe, e nao gosta de barulho."
}
```

Resposta esperada (exemplo):
```json
{
  "session": {
    "session_id": "session-001",
    "friend_id": "c096d057-a0cf-4c2d-857b-41beccb42de8",
    "exists": true,
    "message_count": 2,
    "updated_at": "2026-06-29T12:35:53.833962+00:00"
  },
  "assistant_message": "Entendi. Que tipo de livro ele prefere?"
}
```

### 5.2 Finalizar conversa e persistir profile + embedding
`POST /profiles/agent/finalize`

```json
{
  "session_id": "session-001"
}
```

Observacao:
- Esse endpoint cria/atualiza `likes/dislikes` e gera embedding com base no historico completo da conversa.

### 5.3 Limpar memoria da sessao
`DELETE /profiles/agent/session/{session_id}`

### 5.4 Criar/atualizar profile (fluxo direto opcional)
`POST /profiles`

```json
{
  "friend_id": "c096d057-a0cf-4c2d-857b-41beccb42de8",
  "likes": ["livros", "cafe", "tecnologia"],
  "dislikes": ["barulho", "filas"]
}
```

### 5.5 Gerar embedding (fluxo direto opcional)
`POST /profiles/{friend_id}/embedding`

Campos obrigatorios para compatibilidade atual:
- `friend_id`
- `likes`
- `dislikes`

```json
{
  "friend_id": "c096d057-a0cf-4c2d-857b-41beccb42de8",
  "likes": ["livros", "cafe", "tecnologia"],
  "dislikes": ["barulho", "filas"],
  "name": "Amigo2",
  "city": "Teste",
  "user_relation": "Amigo"
}
```

### 5.6 Criar sugestoes
`POST /profiles/{friend_id}/suggestions`

```json
{
  "suggestions": [
    {
      "type": "gift",
      "title": "Livro de ficcao",
      "reason": "Combina com os interesses informados",
      "price_range": "R$ 40 - R$ 90",
      "confidence": 0.82
    }
  ]
}
```

### 5.7 Criar evento (reminder)
`POST /profiles/{friend_id}/events`

```json
{
  "user_id": "f76d78fd-7473-4611-a6b3-330e1b5e3bae",
  "event_type": "custom",
  "event_name": "Lembrete de presente",
  "event_date": "2026-12-01"
}
```

### 5.8 Listar eventos do friend
`GET /profiles/{friend_id}/events?user_id=<user_id>`

Exemplo:
`GET /profiles/c096d057-a0cf-4c2d-857b-41beccb42de8/events?user_id=f76d78fd-7473-4611-a6b3-330e1b5e3bae`

### 5.9 Listar eventos proximos
`GET /events/upcoming?user_id=<user_id>&days=120`

---

## 6. Exemplo de sequencia de integracao no backend

Pseudo fluxo:
1. Backend gera `session_id` por conversa de profile
2. A cada mensagem do usuario, backend chama `POST /profiles/agent/chat`
3. Exibe `assistant_message` e continua o loop de conversa
4. Ao concluir coleta de dados, backend chama `POST /profiles/agent/finalize`
5. Opcionalmente limpa cache com `DELETE /profiles/agent/session/{session_id}`
6. Quando precisar de recomendacoes, chama `POST /profiles/{friend_id}/suggestions`
7. Para agenda, chama `POST /profiles/{friend_id}/events`
8. Para proximos lembretes, chama `GET /events/upcoming`

---

## 7. Tratamento de erros no backend consumidor
Padrao esperado da API Python:
```json
{
  "error": {
    "message": "Backend de dados retornou erro",
    "details": "..."
  }
}
```

Recomendacoes:
- Repassar `status code` original para a camada de aplicacao
- Logar `message` e `details` com `X-Request-Id`
- Aplicar retry somente para `502` e `504`
- Nao aplicar retry cego para `400` e `422`

---

## 8. Requisitos de idempotencia e concorrencia
- `POST /profiles/agent/chat` pode ser chamado repetidamente para o mesmo `session_id`
- `POST /profiles/agent/finalize` deve ser chamado uma vez por conversa concluida
- Em caso de chamadas paralelas, serializar por `session_id` (chat) e por `friend_id` (finalize)
- Fluxo direto (`POST /profiles` e `POST /profiles/{friend_id}/embedding`) continua suportado para fallback

---

## 9. Recomendacoes de observabilidade
- Sempre enviar `X-Request-Id`
- Medir latencia por endpoint
- Medir taxa de erro por endpoint
- Criar alerta para falhas consecutivas em `POST /profiles/{friend_id}/embedding`
- Criar alerta para falhas consecutivas em `POST /profiles/agent/finalize`

---

## 10. Checklist para entrar em producao
- [ ] Health check `GET /health` integrado
- [ ] Timeout configurado no consumidor (20s recomendado)
- [ ] Retry configurado apenas para erros transitivos
- [ ] Logs com correlacao de request
- [ ] Fluxo conversacional (`chat -> finalize`) validado com IDs reais
- [ ] Fluxo events/reminders validado com `user_id` real

---

## 11. Exemplo rapido com curl (novo fluxo)

Conversar com o agent:
```bash
curl -X POST "http://localhost:8000/profiles/agent/chat" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"session-001","friend_id":"c096d057-a0cf-4c2d-857b-41beccb42de8","message":"Meu amigo gosta de livros e cafe."}'
```

Finalizar e persistir profile + embedding:
```bash
curl -X POST "http://localhost:8000/profiles/agent/finalize" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"session-001"}'
```

Limpar sessao:
```bash
curl -X DELETE "http://localhost:8000/profiles/agent/session/session-001"
```

## 12. Exemplo rapido com curl (fluxo direto opcional)

Criar profile:
```bash
curl -X POST "http://localhost:8000/profiles" \
  -H "Content-Type: application/json" \
  -d '{"friend_id":"c096d057-a0cf-4c2d-857b-41beccb42de8","likes":["livros"],"dislikes":["barulho"]}'
```

Gerar embedding:
```bash
curl -X POST "http://localhost:8000/profiles/c096d057-a0cf-4c2d-857b-41beccb42de8/embedding" \
  -H "Content-Type: application/json" \
  -d '{"friend_id":"c096d057-a0cf-4c2d-857b-41beccb42de8","likes":["livros"],"dislikes":["barulho"],"name":"Amigo2","city":"Teste","user_relation":"Amigo"}'
```

Criar evento:
```bash
curl -X POST "http://localhost:8000/profiles/c096d057-a0cf-4c2d-857b-41beccb42de8/events" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"f76d78fd-7473-4611-a6b3-330e1b5e3bae","event_type":"custom","event_name":"Lembrete de presente","event_date":"2026-12-01"}'
```
