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
	ws "github.com/dvaumoron/puzzlewidgetserver"
	pb "github.com/dvaumoron/puzzlewidgetservice"
)

const GalleryKey = "puzzleGallery"
const widgetName = "gallery"

const formKey = "formData"

func InitWidget(server ws.WidgetServer, service galleryservice.GalleryService) {
	logger := server.Logger()
	w := server.CreateWidget(widgetName)
	w.AddAction("list", pb.MethodKind_GET, "/", func(ctx context.Context, data ws.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)

		galleryId, err := ws.AsUint64(data["objectId"])
		if err != nil {
			return "", "", nil, err
		}

		// TODO paginate

		total, images, err := service.GetImages(ctxLogger, galleryId, 0, 0)
		if err != nil {
			return "", "", nil, err
		}

		newData := ws.Data{}
		data["Total"] = total
		newData["Images"] = images

		resData, err := json.Marshal(newData)
		if err != nil {
			return "", "", nil, err
		}
		return "", "gallery/view", resData, nil
	})
	w.AddAction("retrieve", pb.MethodKind_RAW, "/retrieve/:ImageId", func(ctx context.Context, data ws.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		imageId, err := ws.AsUint64(data["pathData/ImageId"])
		if err != nil {
			return "", "", nil, err
		}

		image, err := service.GetImageData(ctxLogger, imageId)
		if err != nil {
			return "", "", nil, err
		}
		return "", "", image, nil
	})
	w.AddAction("edit", pb.MethodKind_GET, "/edit/:ImageId", func(ctx context.Context, data ws.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		imageId, err := ws.AsUint64(data["pathData/ImageId"])
		if err != nil {
			return "", "", nil, err
		}

		image, err := service.GetImage(ctxLogger, imageId)
		if err != nil {
			return "", "", nil, err
		}

		newData := ws.Data{}
		newData["Image"] = image
		resData, err := json.Marshal(newData)
		if err != nil {
			return "", "", nil, err
		}
		return "", "gallery/edit", resData, nil
	})
	w.AddAction("save", pb.MethodKind_POST, "/save", func(ctx context.Context, data ws.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		galleryId, err := ws.AsUint64(data["objectId"])
		if err != nil {
			return "", "", nil, err
		}

		formData, err := ws.AsMap(data[formKey])
		if err != nil {
			return "", "", nil, err
		}

		imageId, err := ws.AsUint64(formData["ImageId"])
		if err != nil {
			return "", "", nil, err
		}

		userId, err := ws.AsUint64(data["Id"])
		if err != nil {
			return "", "", nil, err
		}

		title, err := ws.AsString(formData["Title"])
		if err != nil {
			return "", "", nil, err
		}

		desc, err := ws.AsString(formData["Desc"])
		if err != nil {
			return "", "", nil, err
		}

		imageInfo := galleryservice.GalleryImage{ImageId: imageId, UserId: userId, Title: title, Desc: desc}

		files, err := ws.GetFiles(data)
		if err != nil {
			return "", "", nil, err
		}

		if _, err = service.UpdateImage(ctxLogger, galleryId, imageInfo, files["image"]); err != nil {
			return "", "", nil, err
		}

		listUrl, err := ws.GetBaseUrl(1, data)
		if err != nil {
			return "", "", nil, err
		}
		return listUrl, "", nil, nil
	})
	w.AddAction("delete", pb.MethodKind_POST, "/delete/:ImageId", func(ctx context.Context, data ws.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		imageId, err := ws.AsUint64(data["pathData/ImageId"])
		if err != nil {
			return "", "", nil, err
		}

		if err = service.DeleteImage(ctxLogger, imageId); err != nil {
			return "", "", nil, err
		}

		listUrl, err := ws.GetBaseUrl(2, data)
		if err != nil {
			return "", "", nil, err
		}
		return listUrl, "", nil, nil
	})
}
