#!/bin/bash

# Exemplos de uso da API do Sistema de Leilões
# Execute este script após iniciar o projeto com docker-compose up

BASE_URL="http://localhost:8080"

echo "=== Sistema de Leilões - Exemplos de API ==="
echo ""

# 1. Criar um leilão
echo "1. Criando um leilão..."
AUCTION_RESPONSE=$(curl -s -X POST $BASE_URL/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15 Pro Max",
    "category": "Electronics",
    "description": "Novo iPhone 15 Pro Max 256GB Titanium",
    "condition": 1
  }')

echo "Resposta: $AUCTION_RESPONSE"
AUCTION_ID=$(echo $AUCTION_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "ID do Leilão: $AUCTION_ID"
echo ""

# 2. Listar todos os leilões
echo "2. Listando todos os leilões..."
curl -s -X GET $BASE_URL/auction | jq '.'
echo ""

# 3. Buscar leilão específico
echo "3. Buscando leilão específico..."
curl -s -X GET $BASE_URL/auction/$AUCTION_ID | jq '.'
echo ""

# 4. Criar lances
echo "4. Criando lances..."
curl -s -X POST $BASE_URL/bid \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user1\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 5000.00
  }"

curl -s -X POST $BASE_URL/bid \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user2\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 5500.00
  }"

curl -s -X POST $BASE_URL/bid \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user3\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 6000.00
  }"
echo ""

# 5. Listar lances do leilão
echo "5. Listando lances do leilão..."
curl -s -X GET $BASE_URL/bid/$AUCTION_ID | jq '.'
echo ""

# 6. Buscar lance vencedor
echo "6. Buscando lance vencedor..."
curl -s -X GET $BASE_URL/auction/winner/$AUCTION_ID | jq '.'
echo ""

# 7. Aguardar fechamento automático (se AUCTION_INTERVAL=20s)
echo "7. Aguardando fechamento automático do leilão..."
echo "Aguardando 25 segundos para o leilão ser fechado automaticamente..."
sleep 25

# 8. Verificar status do leilão após fechamento
echo "8. Verificando status do leilão após fechamento..."
curl -s -X GET $BASE_URL/auction/$AUCTION_ID | jq '.'
echo ""

# 9. Buscar lance vencedor após fechamento
echo "9. Buscando lance vencedor após fechamento..."
curl -s -X GET $BASE_URL/auction/winner/$AUCTION_ID | jq '.'
echo ""

echo "=== Teste concluído ==="
echo "Para ver os logs do fechamento automático, execute:"
echo "docker-compose logs -f app" 