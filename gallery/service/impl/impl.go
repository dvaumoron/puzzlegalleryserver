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
	"github.com/dvaumoron/puzzlegalleryserver/gallery/service"
	mongoclient "github.com/dvaumoron/puzzlemongoclient"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "images"

const galleryIdKey = "galleryId"
const imageIdKey = "imageId"
const userIdKey = "userId"
const titleKey = "title"
const descKey = "desc"
const imageKey = "imageData"

var optsExcludeImageField = options.Find().SetProjection(bson.D{{Key: imageKey, Value: false}})
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
	cursor, err := collection.Find(ctx, filter, optsExcludeImageField)
	if err != nil {
		return 0, nil, err
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return 0, nil, err
	}

	// TODO

	return 0, nil, err
}

func (i galleryImpl) GetImage(logger otelzap.LoggerWithCtx, imageId uint64) (service.GalleryImage, error) {
	ctx := logger.Context()
	client, err := mongo.Connect(ctx, i.clientOptions)
	if err != nil {
		return service.GalleryImage{}, err
	}
	defer mongoclient.Disconnect(client, logger)

	collection := client.Database(i.databaseName).Collection(collectionName)

	var result bson.D
	err = collection.FindOne(
		ctx, bson.D{{Key: imageIdKey, Value: imageId}}, optsOneExcludeImageField,
	).Decode(&result)
	if err != nil {
		return service.GalleryImage{}, err
	}

	//TODO
	return service.GalleryImage{}, nil
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
	image := mongoclient.ExtractBinary(result[1].Value)
	return image, nil
}

func (i galleryImpl) UpdateImage(logger otelzap.LoggerWithCtx, info service.GalleryImage, data []byte) (uint64, error) {
	// TODO
	return 0, nil
}

func (i galleryImpl) DeleteImage(logger otelzap.LoggerWithCtx, imageId uint64) error {
	// TODO
	return nil
}
