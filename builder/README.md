# builder

```bash
GIT_IMPORT=github.com/thinkgos/proc-extra/builder
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_FULL_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell git describe --abbrev=0 --tags --always --match "v*")
BUILD_DATE=$(shell date "+%F %T %z")

BUILDER_LDFLAGS=-X '${GIT_IMPORT}.Version=${VERSION}' \
				-X '${GIT_IMPORT}.GitBranch=${GIT_BRANCH}' \
				-X '${GIT_IMPORT}.GitCommit=${GIT_COMMIT}' \
				-X '${GIT_IMPORT}.GitFullCommit=${GIT_FULL_COMMIT}' \
				-X '${GIT_IMPORT}.GitTag=${GIT_TAG}' \
				-X '${GIT_IMPORT}.BuildDate=${BUILD_DATE}'
```