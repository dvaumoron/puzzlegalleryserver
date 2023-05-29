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
package widget

import (
	"context"
	"encoding/json"

	galleryservice "github.com/dvaumoron/puzzlegalleryserver/gallery/service"

	widgetserver "github.com/dvaumoron/puzzlewidgetserver"
	pb "github.com/dvaumoron/puzzlewidgetservice"
)

const GalleryKey = "puzzleGallery"
const widgetName = "gallery"

const formKey = "formData"

func InitWidget(server widgetserver.WidgetServer, service galleryservice.GalleryService) {
	logger := server.Logger()
	w := server.CreateWidget(widgetName)
	w.AddAction("view", pb.MethodKind_GET, "/", func(ctx context.Context, data widgetserver.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)

		galleryId, err := widgetserver.AsUint64(data["objectId"])
		if err != nil {
			return "", "", nil, err
		}

		total, images, err := service.GetImages(ctxLogger, galleryId, 0, 0)
		if err != nil {
			return "", "", nil, err
		}

		newData := widgetserver.Data{}
		data["Total"] = total
		newData["Images"] = images

		resData, err := json.Marshal(newData)
		if err != nil {
			return "", "", nil, err
		}
		return "", "gallery/view", resData, nil
	})
	w.AddAction("retrieve", pb.MethodKind_RAW, "/retrieve/:ImageId", func(ctx context.Context, data widgetserver.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		imageId, err := widgetserver.AsUint64(data["pathData/ImageId"])
		if err != nil {
			return "", "", nil, err
		}

		image, err := service.GetImageData(ctxLogger, imageId)
		if err != nil {
			return "", "", nil, err
		}
		return "", "", image, nil
	})
	w.AddAction("edit", pb.MethodKind_GET, "/edit/:ImageId", func(ctx context.Context, data widgetserver.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		imageId, err := widgetserver.AsUint64(data["pathData/ImageId"])
		if err != nil {
			return "", "", nil, err
		}

		image, err := service.GetImage(ctxLogger, imageId)
		if err != nil {
			return "", "", nil, err
		}

		newData := widgetserver.Data{}
		newData["Image"] = image
		resData, err := json.Marshal(newData)
		if err != nil {
			return "", "", nil, err
		}
		return "", "gallery/edit", resData, nil
	})
	w.AddAction("save", pb.MethodKind_POST, "/save", func(ctx context.Context, data widgetserver.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		galleryId, err := widgetserver.AsUint64(data["objectId"])
		if err != nil {
			return "", "", nil, err
		}

		service.UpdateImage(ctxLogger, galleryId, galleryservice.GalleryImage{}, nil)

		// TODO

		return "", "", nil, nil
	})
	w.AddAction("delete", pb.MethodKind_POST, "/delete/:ImageId", func(ctx context.Context, data widgetserver.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		imageId, err := widgetserver.AsUint64(data["pathData/ImageId"])
		if err != nil {
			return "", "", nil, err
		}

		service.DeleteImage(ctxLogger, imageId)

		// TODO

		return "", "", nil, nil
	})
}
