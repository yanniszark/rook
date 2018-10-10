include ../image.mk

# ====================================================================================
# Image Build Options

CASSANDRA_OPERATOR_IMAGE = $(BUILD_REGISTRY)/cassandra-$(GOARCH)

TEMP := $(shell mktemp -d)

# ====================================================================================
# Build Rook CockroachDB

do.build:
	@echo === docker build $(CASSANDRA_OPERATOR_IMAGE)
	@cp Dockerfile $(TEMP)
	@cp $(OUTPUT_DIR)/bin/linux_$(GOARCH)/rook $(TEMP)
	@docker build $(BUILD_ARGS) \
		-t $(CASSANDRA_OPERATOR_IMAGE) \
		$(TEMP)
	@rm -fr $(TEMP)
