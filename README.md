# Sistema de Leilões em Go

Este projeto implementa um sistema de leilões com funcionalidade de fechamento automático baseado em tempo.

## Funcionalidades

- Criação de leilões
- Sistema de lances (bids)
- Fechamento automático de leilões baseado em tempo configurável
- Monitoramento em tempo real com goroutines
- API REST para interação com o sistema

## Variáveis de Ambiente

O projeto utiliza as seguintes variáveis de ambiente (definidas no arquivo `cmd/auction/.env`):

- `AUCTION_INTERVAL`: Duração do leilão (ex: "20s", "5m", "1h")
- `BATCH_INSERT_INTERVAL`: Intervalo para inserção em lote
- `MAX_BATCH_SIZE`: Tamanho máximo do lote
- `MONGODB_URL`: URL de conexão com MongoDB
- `MONGODB_DB`: Nome do banco de dados

## Como Executar

### Pré-requisitos

- Docker e Docker Compose instalados
- Go 1.20 ou superior

### Executando com Docker

1. Clone o repositório:
```bash
git clone <repository-url>
cd labs-auction-goexpert
```

2. Execute o projeto com Docker Compose:
```bash
docker-compose up --build
```

O projeto estará disponível em `http://localhost:8080`

### Executando Localmente

1. Instale as dependências:
```bash
go mod download
```

2. Configure as variáveis de ambiente no arquivo `cmd/auction/.env`

3. Execute o projeto:
```bash
go run cmd/auction/main.go
```

## API Endpoints

- `GET /auction` - Lista todos os leilões
- `GET /auction/:auctionId` - Busca leilão por ID
- `POST /auction` - Cria novo leilão
- `GET /auction/winner/:auctionId` - Busca lance vencedor
- `POST /bid` - Cria novo lance
- `GET /bid/:auctionId` - Lista lances de um leilão
- `GET /user/:userId` - Busca usuário por ID

## Funcionalidade de Fechamento Automático

### Como Funciona

O sistema implementa uma goroutine que monitora continuamente os leilões ativos e os fecha automaticamente quando o tempo configurado em `AUCTION_INTERVAL` é atingido.

### Implementação

1. **Cálculo de Duração**: A função `CalculateAuctionDuration()` lê a variável de ambiente `AUCTION_INTERVAL` e retorna a duração do leilão.

2. **Monitoramento**: Uma goroutine (`startAuctionMonitor()`) executa a cada minuto para verificar leilões vencidos.

3. **Fechamento Automático**: A função `checkAndCloseExpiredAuctions()` busca todos os leilões ativos e verifica se o tempo de expiração foi atingido.

4. **Atualização de Status**: Quando um leilão expira, seu status é alterado de `Active` para `Completed`.

### Exemplo de Uso

```bash
# Configurar duração do leilão para 30 segundos
export AUCTION_INTERVAL="30s"

# Criar um leilão
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15",
    "category": "Electronics",
    "description": "Novo iPhone 15 Pro Max",
    "condition": 1
  }'

# O leilão será fechado automaticamente após 30 segundos
```

## Executando Testes

Para executar os testes de fechamento automático:

```bash
# Executar todos os testes
go test ./...

# Executar testes específicos
go test ./internal/infra/database/auction/test/...

# Executar com verbose
go test -v ./internal/infra/database/auction/test/...
```

### Testes Implementados

1. **TestAuctionAutoClose**: Testa o fechamento automático de leilões
2. **TestCalculateAuctionDuration**: Testa o cálculo de duração baseado em variáveis de ambiente
3. **TestCloseAuction**: Testa o fechamento manual de leilões

## Estrutura do Projeto

```
labs-auction-goexpert/
├── cmd/auction/           # Ponto de entrada da aplicação
├── configuration/          # Configurações (logger, database)
├── internal/
│   ├── entity/           # Entidades do domínio
│   ├── infra/            # Infraestrutura (database, api)
│   ├── internal_error/   # Tratamento de erros
│   └── usecase/          # Casos de uso
├── docker-compose.yml    # Configuração Docker
├── Dockerfile           # Dockerfile da aplicação
└── go.mod              # Dependências Go
```

## Concorrência

O projeto utiliza goroutines para:

1. **Monitoramento de Leilões**: Uma goroutine verifica periodicamente leilões vencidos
2. **Processamento de Lances**: Múltiplas goroutines processam lances simultaneamente
3. **Sincronização**: Uso de mutexes para garantir thread-safety

## Logs

O sistema utiliza o Zap logger para registrar:
- Criação de leilões
- Fechamento automático de leilões
- Erros de processamento
- Informações de debug

## Monitoramento

Para monitorar o funcionamento do fechamento automático, observe os logs:

```bash
docker-compose logs -f app
```

Você verá mensagens como:
- "Closing expired auction" - quando um leilão é fechado automaticamente
- "Auction closed successfully" - confirmação de fechamento
- "Auction monitor stopped" - quando o monitor é parado 