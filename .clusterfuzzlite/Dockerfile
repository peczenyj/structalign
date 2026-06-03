# ClusterFuzzLite build container: compiles the repo's native Go fuzz targets
# with the OSS-Fuzz toolchain. See https://google.github.io/clusterfuzzlite/
FROM gcr.io/oss-fuzz-base/base-builder-go:v1
COPY . $SRC/structalign
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/structalign
