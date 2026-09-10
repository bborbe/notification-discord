include Makefile.variables
include Makefile.precommit
include Makefile.docker

SERVICE = notification-discord

.PHONY: run
run:
	@go run -mod=mod . -listen="localhost:8081" -datadir="$(DATADIR)" -kafka-brokers="$(KAFKA_BROKERS)" -branch="$(BRANCH)" -v=2
