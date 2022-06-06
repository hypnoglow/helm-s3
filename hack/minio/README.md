This is a workaround for https://github.community/t/how-do-i-properly-override-a-service-entrypoint/17435

```shell
docker image build -t hypnoglow/minio:latest -f hack/minio/Dockerfile .
docker image push hypnoglow/minio:latest
```
