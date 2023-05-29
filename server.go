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

	mongoclient "github.com/dvaumoron/puzzlemongoclient"
	widgetserver "github.com/dvaumoron/puzzlewidgetserver"
	pb "github.com/dvaumoron/puzzlewidgetservice"
)

//go:embed version.txt
var version string

func main() {
	s := widgetserver.Make("", version)
	clientOptions, databaseName := mongoclient.Create()
	g := s.CreateWidget("gallery")
	g.AddAction("view", pb.MethodKind_GET, "/", func(d widgetserver.Data) (string, string, []byte, error) {
		return "", "", nil, nil
	})

	s.Start()
}
