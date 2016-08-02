CGO_ENABLED=0 go build -o app .
docker build -t "registry.aliyuncs.com/toy/news" .
docker push registry.aliyuncs.com/toy/news