PKG := github.com/xbit-gg/xln

GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v
GOTEST := GO111MODULE=on go test

PROTO_GEN := xlnrpc/xln_grpc.pb.go \
						 xlnrpc/xln.pb.gw.go \
						 xlnrpc/xlnadmin_grpc.pb.go \
						 xlnrpc/xlnadmin.pb.gw.go \
						 xlnrpc/lnurl_grpc.pb.go \
						 xlnrpc/lnurl.pb.gw.go

# BUILD
build: ${PROTO_GEN}
	$(GOBUILD) -o xln $(PKG)/cmd/xln
	@$(call print, "Building xln.")

install: build
	@$(call print, "Installing xbit.")
	$(GOINSTALL) $(PKG)/cmd/xln

# TEST
test:
	TEST_ENV=UNIT $(GOTEST) ./...

integration_test:
	TEST_ENV=PostgresIntegration $(GOTEST) ./...
	TEST_ENV=SqliteIntegration $(GOTEST) ./...

e2e_test:
	TEST_ENV=PostgresE2E $(GOTEST) ./...

# PROTOS
%_grpc.pb.go: %.proto
	protoc --proto_path xlnrpc/ --go_out=paths=source_relative:xlnrpc/ --go-grpc_out=paths=source_relative:xlnrpc/ $<

%.pb.gw.go: %.proto %.yaml
	protoc -I/usr/local/include -I. --proto_path xlnrpc/ \
       --grpc-gateway_out=logtostderr=true,paths=source_relative,grpc_api_configuration=$(word 2,$^):. $<

# CLEANUP
clean:
	rm -f vln
	find . -name "*.pb.go" -type f -delete
	find . -name "*.pb.gw.go" -type f -delete

.PHONY: clean \
		install \
		build \
		build_protos
