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
	Collection      *mongo.Collection
	auctionInterval time.Duration
	mutex           *sync.Mutex
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	repo := &AuctionRepository{
		Collection:      database.Collection("auctions"),
		auctionInterval: getAuctionInterval(),
		mutex:           &sync.Mutex{},
	}
	
	go repo.startAuctionClosingRoutine()
	
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

func (ar *AuctionRepository) UpdateAuctionStatus(
	ctx context.Context,
	auctionId string,
	status auction_entity.AuctionStatus) *internal_error.InternalError {
	
	ar.mutex.Lock()
	defer ar.mutex.Unlock()
	
	filter := bson.M{"_id": auctionId}
	update := bson.M{"$set": bson.M{"status": status}}
	
	_, err := ar.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Error trying to update auction status", err)
		return internal_error.NewInternalServerError("Error trying to update auction status")
	}
	
	return nil
}

func (ar *AuctionRepository) startAuctionClosingRoutine() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ar.checkAndCloseExpiredAuctions()
		}
	}
}

func (ar *AuctionRepository) checkAndCloseExpiredAuctions() {
	ctx := context.Background()
	
	filter := bson.M{
		"status": auction_entity.Active,
		"timestamp": bson.M{
			"$lt": time.Now().Add(-ar.auctionInterval).Unix(),
		},
	}
	
	cursor, err := ar.Collection.Find(ctx, filter)
	if err != nil {
		logger.Error("Error finding expired auctions", err)
		return
	}
	defer cursor.Close(ctx)
	
	var expiredAuctions []AuctionEntityMongo
	if err := cursor.All(ctx, &expiredAuctions); err != nil {
		logger.Error("Error decoding expired auctions", err)
		return
	}
	
	for _, auction := range expiredAuctions {
		if updateErr := ar.UpdateAuctionStatus(ctx, auction.Id, auction_entity.Completed); updateErr != nil {
			logger.Error("Error updating auction status to completed", updateErr)
		} else {
			logger.Info("Auction closed automatically: " + auction.Id)
		}
	}
}

func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5
	}
	
	return duration
}
