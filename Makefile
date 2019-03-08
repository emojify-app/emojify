build_protos:
	protoc -I protos/ protos/emojify.proto --go_out=plugins=grpc:protos/emojify
