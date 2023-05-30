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
package main

import (
	_ "embed"
	"os"
	"strconv"

	galleryimpl "github.com/dvaumoron/puzzlegalleryserver/gallery/service/impl"
	gallerywidget "github.com/dvaumoron/puzzlegalleryserver/gallery/widget"
	mongoclient "github.com/dvaumoron/puzzlemongoclient"
	widgetserver "github.com/dvaumoron/puzzlewidgetserver"
)

//go:embed version.txt
var version string

func main() {
	s := widgetserver.Make(gallerywidget.GalleryKey, version)

	defaultPageSize, _ := strconv.ParseUint(os.Getenv("PAGE_SIZE"), 10, 64)
	if defaultPageSize == 0 {
		defaultPageSize = 20
	}

	clientOptions, databaseName := mongoclient.Create()

	gallerywidget.InitWidget(s, defaultPageSize, galleryimpl.New(clientOptions, databaseName))
	s.Start()
}
