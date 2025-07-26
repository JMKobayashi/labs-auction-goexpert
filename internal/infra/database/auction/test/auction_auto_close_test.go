package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/infra/database/auction"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionAutoClose(t *testing.T) {
	// Configura variável de ambiente para teste
	os.Setenv("AUCTION_INTERVAL", "2s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// Conecta ao MongoDB de teste
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	database := client.Database("test_auctions")
	defer database.Drop(ctx)

	// Cria o repositório
	repo := auction.NewAuctionRepository(database)
	defer repo.StopAuctionMonitor()

	// Cria um leilão que expira em 1 segundo
	auctionEntity, err := auction_entity.CreateAuction(
		"Test Product",
		"Electronics",
		"Test Description",
		auction_entity.New,
	)
	require.NoError(t, err)

	// Modifica o timestamp para simular um leilão criado há 3 segundos
	auctionEntity.Timestamp = time.Now().Add(-3 * time.Second)

	// Insere o leilão no banco
	err = repo.CreateAuction(ctx, auctionEntity)
	require.NoError(t, err)

	// Verifica se o leilão está ativo inicialmente
	foundAuction, err := repo.FindAuctionById(ctx, auctionEntity.Id)
	require.NoError(t, err)
	assert.Equal(t, auction_entity.Active, foundAuction.Status)

	// Aguarda o tempo suficiente para o leilão ser fechado automaticamente
	time.Sleep(3 * time.Second)

	// Verifica se o leilão foi fechado automaticamente
	foundAuction, err = repo.FindAuctionById(ctx, auctionEntity.Id)
	require.NoError(t, err)
	assert.Equal(t, auction_entity.Completed, foundAuction.Status)
}

func TestCalculateAuctionDuration(t *testing.T) {
	// Testa com intervalo válido
	os.Setenv("AUCTION_INTERVAL", "5m")
	duration := auction.CalculateAuctionDuration()
	assert.Equal(t, 5*time.Minute, duration)

	// Testa com intervalo inválido (deve usar default)
	os.Setenv("AUCTION_INTERVAL", "invalid")
	duration = auction.CalculateAuctionDuration()
	assert.Equal(t, 5*time.Minute, duration)

	// Testa com intervalo em segundos
	os.Setenv("AUCTION_INTERVAL", "30s")
	duration = auction.CalculateAuctionDuration()
	assert.Equal(t, 30*time.Second, duration)
}

func TestCloseAuction(t *testing.T) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	database := client.Database("test_auctions_close")
	defer database.Drop(ctx)

	repo := auction.NewAuctionRepository(database)
	defer repo.StopAuctionMonitor()

	// Cria um leilão
	auctionEntity, err := auction_entity.CreateAuction(
		"Test Product",
		"Electronics",
		"Test Description",
		auction_entity.New,
	)
	require.NoError(t, err)

	// Insere o leilão
	err = repo.CreateAuction(ctx, auctionEntity)
	require.NoError(t, err)

	// Verifica se está ativo
	foundAuction, err := repo.FindAuctionById(ctx, auctionEntity.Id)
	require.NoError(t, err)
	assert.Equal(t, auction_entity.Active, foundAuction.Status)

	// Fecha o leilão manualmente
	err = repo.CloseAuction(ctx, auctionEntity.Id)
	require.NoError(t, err)

	// Verifica se foi fechado
	foundAuction, err = repo.FindAuctionById(ctx, auctionEntity.Id)
	require.NoError(t, err)
	assert.Equal(t, auction_entity.Completed, foundAuction.Status)
}
