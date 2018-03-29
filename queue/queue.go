package queue

import (
	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/gazoon/go-utils"
	"github.com/gazoon/go-utils/logging"
	"github.com/gazoon/go-utils/mongo"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/satori/go.uuid"
)

const (
	maxProcessingTime = 20000
)

type MongoWriter struct {
	client *mgo.Database
}

func NewMongoWriter(settings *utils.MongoDBSettings) (*MongoWriter, error) {

	client, err := mongo.Connect(settings)
	if err != nil {
		return nil, err
	}
	return &MongoWriter{client: client}, nil
}

func (self *MongoWriter) Put(ctx context.Context, queueName string, chatId int, message interface{}) error {
	collection := self.client.C(queueName)
	messageEnvelope := map[string]interface{}{
		"created_at": utils.TimestampMilliseconds(),
		"payload":    message,
	}
	_, err := collection.Upsert(
		bson.M{"chat_id": chatId},
		bson.M{
			"$set":  bson.M{"chat_id": chatId},
			"$push": bson.M{"msgs": messageEnvelope},
		})
	return err
}

type MongoReader struct {
	*logging.LoggerMixin
	client *mgo.Collection
}

func NewMongoReader(settings *utils.MongoDBSettings) (*MongoReader, error) {

	client, err := mongo.Connect(settings)
	if err != nil {
		return nil, err
	}
	logger := logging.NewLoggerMixin("mongo_queue_reader", nil)
	return &MongoReader{client: client.C(settings.Collection), LoggerMixin: logger}, nil
}

type Document struct {
	ChatID     int                      `bson:"chat_id"`
	Msgs       []map[string]interface{} `bson:"msgs"`
	Processing struct {
		StartedAt int    `bson:"started_at"`
		Id        string `bson:"id"`
	} `bson:"processing"`
}

func (self *MongoReader) GetNext() (interface{}, string, error) {
	var doc Document
	currentTime := utils.TimestampMilliseconds()
	processingID := uuid.NewV4().String()
	_, err := self.client.Find(
		bson.M{
			"$or": []bson.M{
				{"processing.started_at": bson.M{"$exists": false}},
				{"processing.started_at": bson.M{"$lt": currentTime - maxProcessingTime}},
			}}).Sort("msgs.0.created_at").Apply(
		mgo.Change{Update: bson.M{
			"$set": bson.M{"processing": bson.M{"started_at": currentTime, "id": processingID}},
			"$pop": bson.M{"msgs": -1},
		}},
		&doc)
	if err != nil {
		if err != mgo.ErrNotFound {
			return nil, "", err
		}
		return nil, "", nil
	}
	ctx := logging.NewContextBackground(log.WithField("chat_id", doc.ChatID))
	logger := self.GetLogger(ctx)
	if doc.Processing.StartedAt < currentTime-maxProcessingTime {
		logger.Errorf("Processing for chat took to long")
	}
	if len(doc.Msgs) == 0 {
		logger.Warn("Got document without messages, finish processing")
		self.FinishProcessing(ctx, processingID)
		return nil, "", nil
	}
	return doc.Msgs[0]["payload"], processingID, nil
}

func (self *MongoReader) FinishProcessing(ctx context.Context, processingID string) error {
	err := self.client.Remove(bson.M{"msgs": []interface{}{}, "processing.id": processingID})
	if err != nil {
		if err != mgo.ErrNotFound {
			return err
		}
	} else {
		return nil
	}

	err = self.client.Update(
		bson.M{"processing.id": processingID},
		bson.M{"$unset": bson.M{"processing": ""}},
	)
	if err == mgo.ErrNotFound {
		logger := self.GetLogger(ctx)
		logger.Warn("Message document with processing_id %s no longer exists", processingID)
		return nil
	}
	return err
}
