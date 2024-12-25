echo "etcd cluster building"
docker-compose -f  ./etcd-cluster/docker-compose.yml up -d
echo "kong gateway building"
docker-compose -f  ./gateway-kong/docker-compose.yml up -d
echo "test service starting"
go build -o ./src/test-service/test-service ./src/test-service/main.go
./src/test-service/test-service

