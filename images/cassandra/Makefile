include ../image.mk

# ====================================================================================
# Image Build Options

CASSANDRA_OPERATOR_IMAGE = $(BUILD_REGISTRY)/cassandra-$(GOARCH)

TEMP := $(shell mktemp -d)

# ====================================================================================
# Build Rook Cassandra

do.build:
	@echo === docker build $(CASSANDRA_OPERATOR_IMAGE)
	@cp Dockerfile $(TEMP)
	@cp $(OUTPUT_DIR)/bin/linux_$(GOARCH)/rook $(TEMP)
	@docker build $(BUILD_ARGS) \
		--build-arg ARCH=$(GOARCH) \
    	--build-arg TINI_VERSION=$(TINI_VERSION) \
		-t $(CASSANDRA_OPERATOR_IMAGE) \
		$(TEMP)
	@rm -fr $(TEMP)
