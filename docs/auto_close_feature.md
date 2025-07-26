# Funcionalidade de Fechamento Automático de Leilões

## Visão Geral

Esta funcionalidade implementa o fechamento automático de leilões baseado em tempo configurável através de variáveis de ambiente. O sistema utiliza goroutines para monitorar continuamente os leilões ativos e fechá-los automaticamente quando o tempo de duração é atingido.

## Arquitetura

### Componentes Principais

1. **CalculateAuctionDuration()**: Função que calcula a duração do leilão baseada na variável de ambiente `AUCTION_INTERVAL`
2. **startAuctionMonitor()**: Goroutine que inicia o monitoramento de leilões
3. **checkAndCloseExpiredAuctions()**: Função que verifica e fecha leilões vencidos
4. **CloseAuction()**: Função que fecha um leilão específico

### Fluxo de Funcionamento

```
1. Aplicação inicia
   ↓
2. NewAuctionRepository() é chamado
   ↓
3. startAuctionMonitor() inicia goroutine
   ↓
4. A cada minuto, checkAndCloseExpiredAuctions() executa
   ↓
5. Busca todos os leilões ativos
   ↓
6. Para cada leilão, calcula tempo de expiração
   ↓
7. Se expirado, chama CloseAuction()
   ↓
8. Status do leilão é alterado para Completed
```

## Implementação Técnica

### Cálculo de Duração

```go
func CalculateAuctionDuration() time.Duration {
    auctionInterval := os.Getenv("AUCTION_INTERVAL")
    duration, err := time.ParseDuration(auctionInterval)
    if err != nil {
        logger.Error("Error parsing AUCTION_INTERVAL, using default 5 minutes", err)
        return time.Minute * 5
    }
    return duration
}
```

### Monitoramento com Goroutine

```go
func (ar *AuctionRepository) startAuctionMonitor() {
    ar.wg.Add(1)
    go func() {
        defer ar.wg.Done()
        
        ticker := time.NewTicker(time.Minute)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                ar.checkAndCloseExpiredAuctions()
            case <-ar.stopChan:
                logger.Info("Auction monitor stopped")
                return
            }
        }
    }()
}
```

### Verificação e Fechamento

```go
func (ar *AuctionRepository) checkAndCloseExpiredAuctions() {
    ctx := context.Background()
    
    // Busca todos os leilões ativos
    filter := bson.M{"status": auction_entity.Active}
    cursor, err := ar.Collection.Find(ctx, filter)
    // ... processamento
    
    auctionDuration := CalculateAuctionDuration()
    now := time.Now()
    
    for _, auctionMongo := range auctionsMongo {
        auctionTime := time.Unix(auctionMongo.Timestamp, 0)
        expirationTime := auctionTime.Add(auctionDuration)
        
        if now.After(expirationTime) {
            // Fecha o leilão
            ar.CloseAuction(ctx, auctionMongo.Id)
        }
    }
}
```

## Configuração

### Variáveis de Ambiente

- `AUCTION_INTERVAL`: Define a duração do leilão
  - Exemplos: "20s", "5m", "1h", "24h"
  - Padrão: "5m" (se inválido)

### Exemplo de Configuração

```bash
# Leilão de 30 segundos (para testes)
AUCTION_INTERVAL=30s

# Leilão de 5 minutos
AUCTION_INTERVAL=5m

# Leilão de 1 hora
AUCTION_INTERVAL=1h

# Leilão de 24 horas
AUCTION_INTERVAL=24h
```

## Concorrência e Thread Safety

### Mecanismos de Sincronização

1. **WaitGroup**: Garante que a goroutine seja finalizada adequadamente
2. **Channel (stopChan)**: Permite parar o monitoramento de forma controlada
3. **Context**: Usado para operações de banco de dados

### Tratamento de Concorrência

- O monitoramento é executado em uma goroutine separada
- Operações de banco de dados são thread-safe
- Logs são thread-safe através do Zap logger

## Logs e Monitoramento

### Logs Gerados

```json
{
  "level": "info",
  "message": "Closing expired auction",
  "auction_id": "uuid-do-leilao",
  "expired_at": "2024-01-15T10:30:00Z"
}

{
  "level": "info", 
  "message": "Auction closed successfully",
  "auction_id": "uuid-do-leilao"
}

{
  "level": "info",
  "message": "Auction monitor stopped"
}
```

### Monitoramento

Para monitorar o funcionamento:

```bash
# Ver logs em tempo real
docker-compose logs -f app

# Filtrar logs de fechamento
docker-compose logs app | grep "Closing expired auction"
```

## Testes

### Testes Implementados

1. **TestAuctionAutoClose**: Testa o fechamento automático
2. **TestCalculateAuctionDuration**: Testa o cálculo de duração
3. **TestCloseAuction**: Testa o fechamento manual

### Executando Testes

```bash
# Executar todos os testes
go test ./...

# Executar testes específicos
go test ./internal/infra/database/auction/test/...

# Executar com verbose
go test -v ./internal/infra/database/auction/test/...
```

## Exemplo de Uso

### 1. Configurar Duração

```bash
export AUCTION_INTERVAL="30s"
```

### 2. Criar Leilão

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15",
    "category": "Electronics", 
    "description": "Novo iPhone 15 Pro Max",
    "condition": 1
  }'
```

### 3. Aguardar Fechamento

O leilão será fechado automaticamente após 30 segundos.

### 4. Verificar Status

```bash
curl http://localhost:8080/auction/{auction-id}
```

## Considerações de Performance

### Otimizações Implementadas

1. **Verificação Periódica**: Monitoramento a cada minuto (não em tempo real)
2. **Query Otimizada**: Busca apenas leilões ativos
3. **Logs Estruturados**: Uso do Zap para performance
4. **Graceful Shutdown**: Parada controlada do monitoramento

### Limitações

1. **Latência**: Máximo de 1 minuto para detectar expiração
2. **Escalabilidade**: Para muitos leilões, considerar otimizações adicionais
3. **Persistência**: Status é persistido no MongoDB

## Troubleshooting

### Problemas Comuns

1. **Leilão não fecha**: Verificar `AUCTION_INTERVAL` e logs
2. **Monitor não inicia**: Verificar logs de inicialização
3. **Performance lenta**: Verificar quantidade de leilões ativos

### Debug

```bash
# Verificar variáveis de ambiente
docker-compose exec app env | grep AUCTION

# Verificar logs detalhados
docker-compose logs app | grep -E "(auction|monitor|close)"

# Verificar status dos leilões
curl http://localhost:8080/auction
``` 