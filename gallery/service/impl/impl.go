/*
 *
 * Copyright 2023 puzzlegalleryserver authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package impl

import (
	"context"

	"github.com/dvaumoron/puzzlegalleryserver/gallery/service"
	mongoclient "github.com/dvaumoron/puzzlemongoclient"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "images"

const setOperator = "$set"

const galleryIdKey = "galleryId"
const imageIdKey = "imageId"
const userIdKey = "userId"
const titleKey = "title"
const descKey = "desc"
const imageKey = "imageData"

var optsCreateUnexisting = options.Update().SetUpsert(true)
var optsMaxImageId = options.FindOne().SetSort(bson.D{{Key: imageIdKey, Value: -1}}).SetProjection(bson.D{{Key: imageIdKey, Value: true}})
var optsOnlyImageField = options.FindOne().SetProjection(bson.D{{Key: imageKey, Value: true}})
var optsOneExcludeImageField = options.FindOne().SetProjection(bson.D{{Key: imageKey, Value: false}})

type galleryImpl struct {
	clientOptions *options.ClientOptions
	databaseName  string
}

func New() service.GalleryService {
	clientOptions, databaseName := mongoclient.Create()
	return galleryImpl{clientOptions: clientOptions, databaseName: databaseName}
}

func (i galleryImpl) GetImages(logger otelzap.LoggerWithCtx, galleryId uint64, start uint64, end uint64) (uint64, []service.GalleryImage, error) {
	ctx := logger.Context()
	client, err := mongo.Connect(logger.Context(), i.clientOptions)
	if err != nil {
		return 0, nil, err
	}
	defer mongoclient.Disconnect(client, logger)

	collection := client.Database(i.databaseName).Collection(collectionName)
	filter := bson.D{{Key: galleryIdKey, Value: galleryId}}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, nil, err
	}

	cursor, err := collection.Find(ctx, filter, initPaginationOpts(start, end))
	if err != nil {
		return 0, nil, err
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return 0, nil, err
	}
	return uint64(total), mongoclient.ConvertSlice(results, convertToImage), nil
}

func (i galleryImpl) GetImage(logger otelzap.LoggerWithCtx, imageId uint64) (service.GalleryImage, error) {
	ctx := logger.Context()
	client, err := mongo.Connect(ctx, i.clientOptions)
	if err != nil {
		return service.GalleryImage{}, err
	}
	defer mongoclient.Disconnect(client, logger)

	collection := client.Database(i.databaseName).Collection(collectionName)

	var result bson.M
	err = collection.FindOne(
		ctx, bson.D{{Key: imageIdKey, Value: imageId}}, optsOneExcludeImageField,
	).Decode(&result)
	if err != nil {
		return service.GalleryImage{}, err
	}
	return convertToImage(result), nil
}

func (i galleryImpl) GetImageData(logger otelzap.LoggerWithCtx, imageId uint64) ([]byte, error) {
	ctx := logger.Context()
	client, err := mongo.Connect(ctx, i.clientOptions)
	if err != nil {
		return nil, err
	}
	defer mongoclient.Disconnect(client, logger)

	collection := client.Database(i.databaseName).Collection(collectionName)

	var result bson.D
	err = collection.FindOne(
		ctx, bson.D{{Key: imageIdKey, Value: imageId}}, optsOnlyImageField,
	).Decode(&result)
	if err != nil {
		return nil, err
	}

	// call [1] to get image because result has only the id and one field
	return mongoclient.ExtractBinary(result[1].Value), nil
}

func (i galleryImpl) UpdateImage(logger otelzap.LoggerWithCtx, galleryId uint64, info service.GalleryImage, data []byte) (uint64, error) {
	ctx := logger.Context()
	client, err := mongo.Connect(ctx, i.clientOptions)
	if err != nil {
		return 0, err
	}
	defer mongoclient.Disconnect(client, logger)

	collection := client.Database(i.databaseName).Collection(collectionName)

	imageId := info.ImageId
	image := bson.M{galleryIdKey: galleryId, imageIdKey: imageId, userIdKey: info.UserId, titleKey: info.Title, descKey: info.Desc}
	if len(data) != 0 {
		image[imageKey] = data
	}

	if imageId == 0 {
		return createImage(collection, ctx, image)
	}

	return imageId, updateImage(collection, ctx, image)
}

func (i galleryImpl) DeleteImage(logger otelzap.LoggerWithCtx, imageId uint64) error {
	ctx := logger.Context()
	client, err := mongo.Connect(ctx, i.clientOptions)
	if err != nil {
		return err
	}
	defer mongoclient.Disconnect(client, logger)

	collection := client.Database(i.databaseName).Collection(collectionName)

	_, err = collection.DeleteMany(
		ctx, bson.D{{Key: imageIdKey, Value: imageId}},
	)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	return nil
}

func createImage(collection *mongo.Collection, ctx context.Context, image bson.M) (uint64, error) {
	// rely on the mongo server to ensure there will be no duplicate
	imageId := uint64(1)

	var err error
	var result bson.D
GenerateImageIdStep:
	err = collection.FindOne(ctx, bson.D{{Key: galleryIdKey, Value: image[galleryIdKey]}}, optsMaxImageId).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			goto CreateImageStep
		}

		return 0, err
	}

	// call [1] to get imageId because result has only the id and one field
	imageId = mongoclient.ExtractUint64(result[1].Value) + 1

CreateImageStep:
	image[imageIdKey] = imageId
	_, err = collection.InsertOne(ctx, image)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			// retry
			goto GenerateImageIdStep
		}

		return 0, err
	}
	return imageId, nil
}

func updateImage(collection *mongo.Collection, ctx context.Context, image bson.M) error {
	request := bson.D{{Key: setOperator, Value: image}}
	_, err := collection.UpdateOne(
		ctx, bson.D{{Key: imageIdKey, Value: image[imageIdKey]}}, request, optsCreateUnexisting,
	)
	return err
}

func initPaginationOpts(start uint64, end uint64) *options.FindOptions {
	opts := options.Find().SetSort(bson.D{{Key: imageIdKey, Value: -1}}).SetProjection(bson.D{{Key: imageKey, Value: false}})
	castedStart := int64(start)
	return opts.SetSkip(castedStart).SetLimit(int64(end) - castedStart)
}

func convertToImage(image bson.M) service.GalleryImage {
	title, _ := image[titleKey].(string)
	desc, _ := image[descKey].(string)
	return service.GalleryImage{
		ImageId: mongoclient.ExtractUint64(image[imageIdKey]),
		UserId:  mongoclient.ExtractUint64(image[userIdKey]),
		Title:   title, Desc: desc,
	}
}
