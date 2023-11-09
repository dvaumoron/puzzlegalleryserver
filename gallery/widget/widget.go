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

const (
	GalleryKey  = "puzzleGallery"
	GalleryName = "gallery"

	objectIdKey    = "objectId"
	baseUrlName    = "BaseUrl"
	imageKey       = "Image"
	imageIdKey     = "ImageId"
	pathImageIdKey = "pathData/" + imageIdKey
)

func InitWidget(server ws.WidgetServer, widgetName string, service galleryservice.GalleryService, defaultPageSize uint64, args ...string) {
	logger := server.Logger()

	viewTmpl := "gallery/view"
	editTmpl := "gallery/edit"
	switch len(args) {
	default:
		logger.Info("InitWidget should be called with 0 to 2 optional arguments.")
		fallthrough
	case 2:
		if args[1] != "" {
			editTmpl = args[1]
		}
		fallthrough
	case 1:
		if args[0] != "" {
			viewTmpl = args[0]
		}
	case 0:
	}

	w := server.CreateWidget(widgetName)
	w.AddActionWithQuery("list", pb.MethodKind_GET, "/", ws.GetPaginationNames(), func(ctx context.Context, data ws.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)

		pageNumber, start, end, _ := ws.GetPagination(defaultPageSize, data)

		galleryId, err := ws.AsUint64(data[objectIdKey])
		if err != nil {
			return "", "", nil, err
		}

		total, images, err := service.GetImages(ctxLogger, galleryId, start, end)
		if err != nil {
			return "", "", nil, err
		}

		newData := ws.Data{}
		ws.InitPagination(newData, "", pageNumber, end, total)
		newData["Images"] = images
		resData, err := json.Marshal(newData)
		if err != nil {
			return "", "", nil, err
		}
		return "", viewTmpl, resData, nil
	})
	w.AddAction("retrieve", pb.MethodKind_RAW, "/retrieve/:ImageId", func(ctx context.Context, data ws.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		imageId, err := ws.AsUint64(data[pathImageIdKey])
		if err != nil {
			return "", "", nil, err
		}

		image, err := service.GetImageData(ctxLogger, imageId)
		if err != nil {
			return "", "", nil, err
		}
		return "", "", image, nil
	})
	w.AddAction("create", pb.MethodKind_GET, "/create", func(ctx context.Context, data ws.Data) (string, string, []byte, error) {
		baseUrl, err := ws.GetBaseUrl(1, data)
		if err != nil {
			return "", "", nil, err
		}

		newData := ws.Data{}
		newData[imageKey] = galleryservice.GalleryImage{Title: "new"}
		newData[baseUrlName] = baseUrl
		resData, err := json.Marshal(newData)
		if err != nil {
			return "", "", nil, err
		}
		return "", editTmpl, resData, nil
	})
	w.AddAction("edit", pb.MethodKind_GET, "/edit/:ImageId", func(ctx context.Context, data ws.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		imageId, err := ws.AsUint64(data[pathImageIdKey])
		if err != nil {
			return "", "", nil, err
		}

		image, err := service.GetImage(ctxLogger, imageId)
		if err != nil {
			return "", "", nil, err
		}

		baseUrl, err := ws.GetBaseUrl(2, data)
		if err != nil {
			return "", "", nil, err
		}

		newData := ws.Data{}
		newData[imageKey] = image
		newData[baseUrlName] = baseUrl
		resData, err := json.Marshal(newData)
		if err != nil {
			return "", "", nil, err
		}
		return "", editTmpl, resData, nil
	})
	w.AddAction("save", pb.MethodKind_POST, "/save", func(ctx context.Context, data ws.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		galleryId, err := ws.AsUint64(data[objectIdKey])
		if err != nil {
			return "", "", nil, err
		}

		listUrl, err := ws.GetBaseUrl(1, data)
		if err != nil {
			return "", "", nil, err
		}

		userId, err := ws.AsUint64(data["Id"])
		if err != nil {
			return "", "", nil, err
		}
		if userId == 0 {
			return listUrl + "?error=ErrorNotAuthorized", "", nil, nil
		}

		formData, err := ws.GetFormData(data)
		if err != nil {
			return "", "", nil, err
		}

		imageId, err := ws.AsUint64(formData[imageIdKey])
		if err != nil {
			return "", "", nil, err
		}

		title, err := ws.AsString(formData["Title"])
		if err != nil {
			return "", "", nil, err
		}

		if title == "new" || title == "" {
			return listUrl + "?error=ErrorBadImageTitle", "", nil, nil
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
		return listUrl, "", nil, nil
	})
	w.AddAction("delete", pb.MethodKind_POST, "/delete/:ImageId", func(ctx context.Context, data ws.Data) (string, string, []byte, error) {
		ctxLogger := logger.Ctx(ctx)
		imageId, err := ws.AsUint64(data[pathImageIdKey])
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
