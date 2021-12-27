ARG BUILDPLATFORM=linux/amd64
ARG BUILD_BASE_IMAGE

FROM --platform=$BUILDPLATFORM $BUILD_BASE_IMAGE AS build
WORKDIR /Sesame

ARG BUILD_GOPROXY
ENV GOPROXY=${BUILD_GOPROXY}
COPY go.mod go.sum /Sesame/
RUN go mod download

COPY cmd cmd
COPY internal internal
COPY pkg pkg
COPY apis apis
COPY Makefile Makefile

ARG BUILD_BRANCH
ARG BUILD_SHA
ARG BUILD_VERSION
ARG BUILD_CGO_ENABLED
ARG BUILD_EXTRA_GO_LDFLAGS
ARG TARGETOS
ARG TARGETARCH

RUN make build \
	    CGO_ENABLED=${BUILD_CGO_ENABLED} \
		EXTRA_GO_LDFLAGS="${BUILD_EXTRA_GO_LDFLAGS}" \
		GOOS=${TARGETOS} \
		GOARCH=${TARGETARCH} \
	    BUILD_VERSION=${BUILD_VERSION} \
	    BUILD_SHA=${BUILD_SHA} \
	    BUILD_BRANCH=${BUILD_BRANCH}

# Ensure we produced a static binary.
RUN ldd Sesame 2>&1 | grep 'not a dynamic executable'

FROM scratch AS final
COPY --from=build /Sesame/Sesame /bin/Sesame
