#!/usr/bin/env bash

./build/build.sh

buildah from --name puzzlegalleryserver-working-container scratch
buildah copy puzzlegalleryserver-working-container $HOME/go/bin/puzzlegalleryserver /bin/puzzlegalleryserver
buildah config --env SERVICE_PORT=50051 puzzlegalleryserver-working-container
buildah config --port 50051 puzzlegalleryserver-working-container
buildah config --entrypoint '["/bin/puzzlegalleryserver"]' puzzlegalleryserver-working-container
buildah commit puzzlegalleryserver-working-container puzzlegalleryserver
buildah rm puzzlegalleryserver-working-container

buildah push puzzlegalleryserver docker-daemon:puzzlegalleryserver:latest
