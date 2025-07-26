package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}

type AuctionRepository struct {
	Collection *mongo.Collection
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	repo := &AuctionRepository{
		Collection: database.Collection("auctions"),
		stopChan:   make(chan struct{}),
	}

	// Inicia a goroutine para monitorar leilões vencidos
	repo.startAuctionMonitor()

	return repo
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	return nil
}

// CalculateAuctionDuration calcula a duração do leilão baseado nas variáveis de ambiente
func CalculateAuctionDuration() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		logger.Error("Error parsing AUCTION_INTERVAL, using default 5 minutes", err)
		return time.Minute * 5
	}
	return duration
}

// CloseAuction fecha um leilão específico
func (ar *AuctionRepository) CloseAuction(ctx context.Context, auctionId string) *internal_error.InternalError {
	filter := bson.M{"_id": auctionId}
	update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

	_, err := ar.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Error trying to close auction", err)
		return internal_error.NewInternalServerError("Error trying to close auction")
	}

	logger.Info("Auction closed successfully", zap.String("auction_id", auctionId))

	return nil
}

// startAuctionMonitor inicia a goroutine para monitorar leilões vencidos
func (ar *AuctionRepository) startAuctionMonitor() {
	ar.wg.Add(1)
	go func() {
		defer ar.wg.Done()

		ticker := time.NewTicker(time.Minute) // Verifica a cada minuto
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

// checkAndCloseExpiredAuctions verifica e fecha leilões vencidos
func (ar *AuctionRepository) checkAndCloseExpiredAuctions() {
	ctx := context.Background()

	// Busca todos os leilões ativos
	filter := bson.M{"status": auction_entity.Active}
	cursor, err := ar.Collection.Find(ctx, filter)
	if err != nil {
		logger.Error("Error finding active auctions", err)
		return
	}
	defer cursor.Close(ctx)

	var auctionsMongo []AuctionEntityMongo
	if err := cursor.All(ctx, &auctionsMongo); err != nil {
		logger.Error("Error decoding auctions", err)
		return
	}

	auctionDuration := CalculateAuctionDuration()
	now := time.Now()

	for _, auctionMongo := range auctionsMongo {
		auctionTime := time.Unix(auctionMongo.Timestamp, 0)
		expirationTime := auctionTime.Add(auctionDuration)

		if now.After(expirationTime) {
			logger.Info("Closing expired auction", zap.String("auction_id", auctionMongo.Id), zap.Time("expired_at", expirationTime))

			if err := ar.CloseAuction(ctx, auctionMongo.Id); err != nil {
				logger.Error("Error closing expired auction", err)
			}
		}
	}
}

// StopAuctionMonitor para a goroutine de monitoramento
func (ar *AuctionRepository) StopAuctionMonitor() {
	close(ar.stopChan)
	ar.wg.Wait()
}
