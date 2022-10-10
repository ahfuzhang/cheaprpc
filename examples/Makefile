
proto_path=/Users/ahfuzhang/code/

build: github.com/ahfuzhang/my_easy_service/pb/my_easy_service.pb.go github.com/ahfuzhang/my_easy_service/main.go
	@echo "ok"

gen_code: ../cmd/gen_code/gen_code
	go build -o gen_code ../cmd/gen_code/main.go

github.com/ahfuzhang/my_easy_service/pb/my_easy_service.pb.go:
	protoc -I=. -I=${proto_path} --gogofaster_out=plugins=grpc:. *.proto

github.com/ahfuzhang/my_easy_service/main.go: gen_code
	./gen_code -proto_path=${proto_path} \
	    -source_proto=./my_easy_service.proto \
	    -target_path=./ \
	    -template_path=../templates/
