.PHONY: all
all: client server

.PHONY: client
client:
	cd cmd/$@ && $(MAKE)

.PHONY: server
server:
	cd cmd/$@ && $(MAKE)
