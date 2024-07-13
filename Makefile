compile_protos:
	protoc protos/notification.proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

make install_dependencies:
	cd ./server && go mod tidy
	cd ./client && go mod tidy