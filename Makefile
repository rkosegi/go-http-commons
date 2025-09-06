# Copyright 2025 Richard Kosegi
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

VERSION 		:= $(shell cat VERSION)
VER_PARTS   	:= $(subst ., ,$(VERSION))
VER_MAJOR		:= $(word 1,$(VER_PARTS))
VER_MINOR   	:= $(word 2,$(VER_PARTS))
VER_PATCH   	:= $(word 3,$(VER_PARTS))
VER_NEXT_PATCH  := $(VER_MAJOR).$(VER_MINOR).$(shell echo $$(($(VER_PATCH)+1)))

.DEFAULT_GOAL := build-local

update-go-deps:
	@for m in $$(go list -mod=readonly -m -f '{{ if and (not .Indirect) (not .Main)}}{{.Path}}{{end}}' all); do \
		go get $$m; \
	done
	@go mod tidy


build-local:
	go mod tidy
	go fmt ./...
	CGO_ENABLED=0 go build -v ./...

lint:
	pre-commit run --all-files

test:
	go test $$(go list ./...) -coverprofile=coverage.out

bump-patch-version:
	@echo Current: $(VERSION)
	@echo Next: $(VER_NEXT_PATCH)
	@echo "$(VER_NEXT_PATCH)" > VERSION
	git add -- VERSION
	git commit -sm "Bump version to $(VER_NEXT_PATCH)"

git-tag:
	git tag -am "Release v$(VERSION)" v$(VERSION)

git-push-tag:
	git push --tags

new-release:
	$(MAKE) bump-patch-version
	$(MAKE) git-tag
