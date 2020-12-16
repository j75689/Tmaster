#!/usr/bin/env bash
. ./build/docker.env

VERSION=$(git describe --tags)
IMAGE_NAME="${PROJECT_NAME}/${SERVICE_NAME}"
docker build . -f ./build/Dockerfile -t "${IMAGE_NAME}:${VERSION}" -t "${IMAGE_NAME}:latest"
docker push "${IMAGE_NAME}:latest"
docker push "${IMAGE_NAME}:${VERSION}"