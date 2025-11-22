ARG GO_VERSION=1.25.3
ARG HELM_VERSION=3.19.0

FROM golang:${GO_VERSION}-alpine AS build

ARG PLUGIN_VERSION=master

RUN apk add --no-cache git

WORKDIR /workspace/helm-s3

COPY . .

RUN CGO_ENABLED=0 \
    go build  \
    -trimpath \
    -mod=vendor \
    -ldflags "-X main.version=${PLUGIN_VERSION}" \
    -o bin/helm-s3 \
    ./cmd/helm-s3

# Correct the plugin manifest with docker-specific fixes:
# - remove hooks, because we are building everything locally from source
# - update version
RUN sed "/^hooks:/,+2 d" plugin.yaml > plugin.yaml.fixed \
    && sed -i "s/^version:.*$/version: ${PLUGIN_VERSION}/" plugin.yaml.fixed

FROM alpine/helm:${HELM_VERSION}

# Build-time metadata as defined at http://label-schema.org
ARG BUILD_DATE
ARG VCS_REF
ARG PLUGIN_VERSION
LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name="helm-s3" \
      org.label-schema.description="The Helm plugin that provides S3 protocol support and allows to use AWS S3 as a chart repository." \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/hypnoglow/helm-s3" \
      org.label-schema.version=$PLUGIN_VERSION \
      org.label-schema.schema-version="1.0"

COPY --from=build /workspace/helm-s3/plugin.yaml.fixed /root/.helm/cache/plugins/helm-s3/plugin.yaml
COPY --from=build /workspace/helm-s3/bin/helm-s3 /root/.helm/cache/plugins/helm-s3/bin/helm-s3

RUN mkdir -p /root/.helm/plugins \
    && helm plugin install /root/.helm/cache/plugins/helm-s3

ENTRYPOINT []
CMD []
