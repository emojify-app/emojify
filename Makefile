build_protos:
	protoc -I protos/ protos/emojify.proto --go_out=plugins=grpc:protos/emojify

goconvey:
	goconvey -excludedDirs protos

# Spin up environment
run_functional:
	REDIS_ADDRESS=localhost:6379 REDIS_PASSWORD=password FACEBOX_ADDRESS=http://localhost:9070 CACHE_ADDRESS=localhost:9080 LOG_LEVEL=trace go run main.go

run_testapp:
	cd functional && go run main.go
