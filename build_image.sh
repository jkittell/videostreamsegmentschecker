docker login -u jpkitt 

docker buildx create

docker buildx build --push \
--provenance false \
--platform linux/arm64/v8,linux/amd64 \
--tag jpkitt/videostreamsegmentschecker:latest .
