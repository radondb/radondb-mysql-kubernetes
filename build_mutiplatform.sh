
#
TAG=v3.0.0
IMGPREFIX=radondb/
builder_exists=$(docker buildx ls | awk '{if ($1=="multi-platform") print $1}')
	# docker buildx rm multi-platform
	# docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
if [ "$builder_exists" ]; then
	echo "exist the multiarch"
else
	docker buildx create --use --name multi-platform --platform=linux/amd64,linux/arm64 > /dev/null
fi


IMGAMD=${IMGPREFIX}mysql-operator-amd64:${TAG}
IMGARM=${IMGPREFIX}mysql-operator-arm64:${TAG}
GO_PROXY=on
DOCKER_BUILDKIT=1 docker build --build-arg GO_PROXY=${GO_PROXY} -t ${IMGAMD} .

#docker buildx create --use --name multi-platform --driver docker-container --platform=linux/amd64,linux/arm64 --config /root/radondb-mysql-kubernetes/buildkitd.toml > /dev/null
docker buildx build --build-arg GO_PROXY=on   --platform linux/arm64  -t $IMGARM -o type=docker .
docker push ${IMGAMD}
docker push ${IMGARM}
docker manifest create --amend ${IMGPREFIX}mysql-operator:${TAG}  ${IMGAMD} ${IMGARM}
docker manifest  push --purge  ${IMGPREFIX}mysql-operator:${TAG}


XENON_IMGAMD=${IMGPREFIX}xenon-amd64:${TAG}
XENON_IMGARM=${IMGPREFIX}xenon-arm64:${TAG}
GO_PROXY=on
DOCKER_BUILDKIT=1 docker build -f build/xenon/Dockerfile --build-arg GO_PROXY=${GO_PROXY} -t ${XENON_IMGAMD} .
#docker buildx create --use --name multi-platform --driver docker-container --platform=linux/amd64,linux/arm64 --config /root/radondb-mysql-kubernetes/buildkitd.toml > /dev/null
docker buildx build --build-arg GO_PROXY=on  --platform linux/arm64 -f build/xenon/Dockerfile.arm64 --build-arg GO_PROXY=${GO_PROXY} -t ${XENON_IMGARM} -o type=docker   .
docker push $XENON_IMGAMD
docker push $XENON_IMGARM
docker manifest create --amend ${IMGPREFIX}xenon:${TAG}  ${XENON_IMGAMD} ${XENON_IMGARM}
docker manifest  push --purge  ${IMGPREFIX}xenon:${TAG}





SIDECAR57_IMGAMD=${IMGPREFIX}mysql57-sidecar-amd64:${TAG}
SIDECAR80_IMGAMD=${IMGPREFIX}mysql80-sidecar-amd64:${TAG}
SIDECAR57_IMGARM=${IMGPREFIX}mysql57-sidecar-arm64:${TAG}
SIDECAR80_IMGARM=${IMGPREFIX}mysql80-sidecar-arm64:${TAG}
GO_PROXY=on
DOCKER_BUILDKIT=1 docker build --build-arg IMAGE_FROM=radondb/mysql80-sidecar:v2.4.0  --build-arg GO_PROXY=${GO_PROXY} -f  Dockerfile80.sidecar -t ${SIDECAR80_IMGAMD} .
DOCKER_BUILDKIT=1 docker build -f Dockerfile57.sidecar --build-arg GO_PROXY=${GO_PROXY} -t ${SIDECAR57_IMGAMD} .
docker buildx build --build-arg GO_PROXY=on   --platform linux/arm64  --build-arg GO_PROXY=${GO_PROXY} -f  Dockerfile80.sidecar-arm64  -t ${SIDECAR80_IMGARM}  -o type=docker  .
docker buildx build --build-arg GO_PROXY=on   --platform linux/arm64  -f Dockerfile57.sidecar-arm64 --build-arg GO_PROXY=${GO_PROXY} -t ${SIDECAR57_IMGARM}  -o type=docker .
docker push ${SIDECAR57_IMGAMD}
docker push ${SIDECAR80_IMGAMD}
docker push ${SIDECAR57_IMGARM}
docker push ${SIDECAR80_IMGARM}
docker manifest create --amend ${IMGPREFIX}mysql57-sidecar:${TAG}  ${SIDECAR57_IMGAMD} ${SIDECAR57_IMGARM}
docker manifest  push --purge  ${IMGPREFIX}mysql57-sidecar:${TAG} 
docker manifest create --amend ${IMGPREFIX}mysql80-sidecar:${TAG}  ${SIDECAR80_IMGAMD} ${SIDECAR80_IMGARM}
docker manifest  push --purge  ${IMGPREFIX}mysql80-sidecar:${TAG} 



