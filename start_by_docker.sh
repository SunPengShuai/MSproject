
echo "etcd cluster building"
docker-compose -f  ./etcd-cluster/docker-compose.yml up -d
echo "kong gateway building"
docker-compose -f  ./gateway-kong/docker-compose.yml up -d
echo "test service starting"
go run ./src/test-service/main.go -d

