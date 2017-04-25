# The following are targers that do not exist in the filesystem as real files and should be always executed by make
.PHONY: default deps base build dev shell start stop image test vet gogen build_release dep_install dep_update docs_build docs_shell serve_docs generate_docs

# Name of this service/application
SERVICE_NAME := ladder

# Docker image name for this project
IMAGE_NAME := themotion/$(SERVICE_NAME)

# Shell to use for running scripts
SHELL := $(shell which bash)

# Get docker path or an empty string
DOCKER := $(shell command -v docker)

# Get docker-compose path or an empty string
DOCKER_COMPOSE := $(shell command -v docker-compose)

# Get the main unix group for the user running make (to be used by docker-compose later)
GID := $(shell id -g)

# Get the unix user id for the user running make (to be used by docker-compose later)
UID := $(shell id -u)

# Get the username of theuser running make. On the devbox, we give priority to /etc/username
USERNAME ?= $(shell ( [ -f /etc/username ] && cat /etc/username  ) || whoami)

# Bash history file for container shell
HISTORY_FILE := ~/.bash_history.$(SERVICE_NAME)
DOCS_HISTORY_FILE := ~/.bash_history.$(SERVICE_NAME)-docs

# Try to detect current branch if not provided from environment
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

# Commit hash from git
COMMIT=$(shell git rev-parse --short HEAD)

# The default action of this Makefile is to build the development docker image
default: build

# Test if the dependencies we need to run this Makefile are installed
deps:
ifndef DOCKER
	@echo "Docker is not available. Please install docker"
	@exit 1
endif
ifndef DOCKER_COMPOSE
	@echo "docker-compose is not available. Please install docker-compose"
	@exit 1
endif


# Build the base docker image which is shared between the development and production images
base: deps
	docker build -t $(IMAGE_NAME)_base:latest .

# Build the development docker image
build: base
	cd environment/dev && docker-compose build

# Run the development environment in non-daemonized mode (foreground)
dev: build
	cd environment/dev && \
	( docker-compose up; \
		docker-compose stop; \
		docker-compose rm -f; )

# Run a shell into the development docker image
shell: build
	-touch $(HISTORY_FILE)
	cd environment/dev && docker-compose run --service-ports --rm $(SERVICE_NAME) /bin/bash

# Run the development environment in the background
start: build
	cd environment/dev && \
		docker-compose up -d

# Stop the development environment (background and/or foreground)
stop:
	cd environment/dev && ( \
		docker-compose stop; \
		docker-compose rm -f; \
		)

# Build release, target on /bin
build_release:build
		cd environment/dev && docker-compose run --rm $(SERVICE_NAME) /bin/bash -c "./build.sh"

# Update project dependencies to tle latest version
dep_update:build
	cd environment/dev && docker-compose run --rm $(SERVICE_NAME) /bin/bash -c 'glide up --strip-vcs --update-vendored'
# Install new dependency make dep_install args="github.com/Sirupsen/logrus"
dep_install:build
		cd environment/dev && docker-compose run --rm $(SERVICE_NAME) /bin/bash -c 'glide get --strip-vcs $(args)'

# Pass the golang vet check
vet: build
	cd environment/dev && docker-compose run --rm $(SERVICE_NAME) /bin/bash -c 'go vet `glide nv`'

# Execute unit tests
test:build
	cd environment/dev && docker-compose run --rm $(SERVICE_NAME) /bin/bash -c 'go test `glide nv` --tags="integration" -v'

# Generate required code (mocks...)
gogen: build
	cd environment/dev && docker-compose run --rm $(SERVICE_NAME) /bin/bash -c 'go generate `glide nv`'

# Build the production themotion image based on the public one
image: base
	docker build \
	-t $(SERVICE_NAME) \
	-t $(SERVICE_NAME):latest \
	-t $(SERVICE_NAME):$(COMMIT) \
	-t $(SERVICE_NAME):$(BRANCH) \
	-f environment/prod/Dockerfile \
	environment/prod

# Docs command
docs_build:
	cd environment/docs && docker-compose build

docs_shell: docs_build
	-touch $(DOCS_HISTORY_FILE)
	cd environment/docs && docker-compose run --service-ports --rm hugo /bin/bash

serve_docs: docs_build
	cd environment/docs && docker-compose run --service-ports --rm hugo /bin/bash -c 'hugo serve --bind=0.0.0.0'

generate_docs: docs_build
	cd environment/docs && docker-compose run --rm hugo /bin/bash -c "rm -rf ./public && mkdir ./public && hugo -s ./docs/ -d ../public"