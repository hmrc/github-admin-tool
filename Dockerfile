FROM golang:1.16-alpine3.13 as base
RUN apk add --no-cache \
    shadow~=4.8 \
    bash~=5.1
# UID of current user who runs the build
ARG user_id
# GID of current user who runs the build
ARG group_id
# HOME of current user who runs the build
ARG home
# change GID for dialout group which collides with MacOS staff GID (20) and
# create group and user to match permmisions of current who runs the build
ARG workdir
WORKDIR ${workdir}
RUN groupmod -g 64 dialout \
    && addgroup -S \
    -g "${group_id}" \
    union \
    && groupmod -g 2999 ping \
    && mkdir -p "${home}" \
    && adduser -S \
    -u "${user_id}" \
    -h "${home}" \
    -s "/bin/bash" \
    -G union \
    builder \
    && chown -R builder:union "${workdir}"

FROM base AS gofmt
ENTRYPOINT [ "/usr/local/go/bin/gofmt" ]

FROM base AS go
ENV CGO_ENABLED=0
COPY go.mod go.sum ${workdir}/
RUN go mod download
ENTRYPOINT [ "/usr/local/go/bin/go" ]

FROM go AS golangci-lint
ENV GOLANGCI_LINT_VERSION=1.40.0
SHELL [ "/bin/bash", "-euo", "pipefail", "-c" ]
RUN wget -qO- "https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_LINT_VERSION}/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-amd64.tar.gz" \
    | tar -xzv -C /bin --strip-components=1 "golangci-lint-${GOLANGCI_LINT_VERSION}-linux-amd64/golangci-lint"
USER builder
ENTRYPOINT [ "/bin/golangci-lint" ]
