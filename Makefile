BUILD_DIR=build
BINARY=video-downloader-bot
IMAGE=video-downloader-bot

.PHONY: install-deps build run docker-build docker-run

install-deps:
	@(command -v brew > /dev/null && brew install ffmpeg) || \
	 (command -v apt-get > /dev/null && sudo apt-get install -y ffmpeg) || \
	 (echo "No supported package manager found, install ffmpeg manually: https://ffmpeg.org/download.html" && exit 1)
	@command -v pip3 > /dev/null || (echo "pip3 not found, install Python 3 to download yt-dlp with bundled curl-cffi: https://www.python.org/downloads/" && exit 1)
	pip3 install "yt-dlp[default,curl-cffi]"

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/main.go

run: build
	./$(BUILD_DIR)/$(BINARY)

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run -d --name $(IMAGE) \
		-v $(PWD)/config.yaml:/app/config.yaml:ro \
		-v $(PWD)/files:/app/files \
		-v $(PWD)/logs:/app/logs \
		$(IMAGE)
