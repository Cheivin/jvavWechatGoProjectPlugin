.DEFAULT_GOAL := build
.PHONY : build

IMAGE_NAME:=wechat-hub-plugin

build:
	docker build -f build/Dockerfile -t ${IMAGE_NAME} .
