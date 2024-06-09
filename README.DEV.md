# BUILD

> sudo snap install go --classic

> sudo snap install protobuf --classic

> go version

> go mod vendor

> go get ./...

```shell
   go install \
      "google.golang.orggit /protobuf/cmd/protoc-gen-go" \
      "google.golang.org/grpc/cmd/protoc-gen-go-grpc" \
      "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway" \
      "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
```

> env PATH="$HOME/go/bin:$PATH" go generate -x ./...

> go build -trimpath -mod=vendor -o nakama

# DB

> docker compose -f docker-compose-postgres-dev.yml up -d

> ./nakama migrate up --database.address postgres:localdb@localhost:5432/nakama

# RUN

> ./nakama --name nakama1 --database.address postgres:localdb@localhost:5432/nakama --session.token_expiry_sec 7200 -cluster.etcd.endpoints localhost:2379

> ./nakama --name nakama2 --database.address postgres:localdb@localhost:5432/nakama --session.token_expiry_sec 7200 -cluster.etcd.endpoints localhost:2379 -socket.port 7360 -console.port 7361 -cluster.server.gossip_bindport 7362 

<!-- ./nakama --name nakama2 --database.address postgres:localdb@localhost:5432/nakama --logger.level DEBUG --session.token_expiry_sec 7200 -cluster.etcd.endpoints localhost:2379 -socket.port 7360 -console.port 7361 -cluster.server.gossip_bindport 7362 -logger.level debug  -->

# TEST

> curl -x POST http://localhost:7350/v2/console/authenticate 

> curl --header "Content-Type: application/json" --request POST --data '{"username":"admin","password":"password"}' http://localhost:7351/v2/console/authenticate

> curl --header "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInVzbiI6ImFkbWluIiwicm9sIjoxLCJleHAiOjE3MTYyMjI2MzIsImNraSI6Ijk5OGM4MzdjLTlkOTAtNDhiZi04ODgwLWM5MDFhODRmYjkxMiJ9.Ii2qpqbM7bNytuF1FL6Xi7O2pCxbffSc3BOXVgIPJyE" http://localhost:7351/v2/console/account

> curl --header "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInVzbiI6ImFkbWluIiwicm9sIjoxLCJleHAiOjE3MTYyMjI2MzIsImNraSI6Ijk5OGM4MzdjLTlkOTAtNDhiZi04ODgwLWM5MDFhODRmYjkxMiJ9.Ii2qpqbM7bNytuF1FL6Xi7O2pCxbffSc3BOXVgIPJyE" http://localhost:7351/v2/console/demodate

> curl --header "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInVzbiI6ImFkbWluIiwicm9sIjoxLCJleHAiOjE3MTYyMjI2MzIsImNraSI6Ijk5OGM4MzdjLTlkOTAtNDhiZi04ODgwLWM5MDFhODRmYjkxMiJ9.Ii2qpqbM7bNytuF1FL6Xi7O2pCxbffSc3BOXVgIPJyE" http://localhost:7351/v2/console/logdate

# Etcd

> docker compose -f docker-compose-postgres-dev.yml up -d

> ./nakama --name nakama10 --logger.level debug --database.address postgres:localdb@localhost:5432/nakama --session.token_expiry_sec 7200 -socket.port 7350  -console.port 7351 -cluster.server.gossip_bindport 7352 -cluster.etcd.endpoints localhost:2379

> ./nakama --name nakama20 --logger.level debug --database.address postgres:localdb@localhost:5432/nakama --session.token_expiry_sec 7200 -socket.port 7360  -console.port 7361 -cluster.server.gossip_bindport 7362 -cluster.etcd.endpoints localhost:2379

> ./nakama --name nakama10 --logger.level debug --database.address postgres:localdb@localhost:5432/nakama --session.token_expiry_sec 7200 -socket.port 7350  -console.port 7351 -cluster.server.gossip_bindport 7352 -cluster.etcd.endpoints localhost:2379 -runtime.path /home/alphalab/Downloads/nakama-godot-demo-master/nakama/modules -cluster.server.gossip_bindaddr 0.0.0.0

> ./nakama --name nakama20 --logger.level debug --database.address postgres:localdb@localhost:5432/nakama --session.token_expiry_sec 7200 -socket.port 7360  -console.port 7361 -cluster.server.gossip_bindport 7362 -cluster.etcd.endpoints localhost:2379 -runtime.path /home/alphalab/Downloads/nakama-godot-demo-master/nakama/modules -cluster.server.gossip_bindaddr 0.0.0.0
