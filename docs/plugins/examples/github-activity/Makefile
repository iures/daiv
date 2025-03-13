PLUGIN_NAME=daiv-github-activity

.PHONY: build install clean tidy

install: build
	cp ./out/$(PLUGIN_NAME).so ~/.daiv/plugins/

build: tidy
	go build -o ./out/$(PLUGIN_NAME).so -buildmode=plugin main.go

tidy: clean
	go mod tidy

clean:
	rm -f ./out/$(PLUGIN_NAME).so
	rm -f ~/.daiv/plugins/$(PLUGIN_NAME).so 
