# Multi-stage build
FROM golang:1.23.5-bookworm AS builder
RUN apt update && apt upgrade -y && apt install curl git make -y
# Install latest helm cli and plugins
RUN curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
RUN helm plugin install https://github.com/databus23/helm-diff
RUN helm plugin install https://github.com/aslafy-z/helm-git
# Install latest helmfile cli
RUN curl -Lo /usr/local/bin/helmfile https://github.com/helmfile/helmfile/releases/latest/download/helmfile_linux_amd64
RUN chmod +x /usr/local/bin/helmfile
# Install latest kustomize cli
RUN curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
RUN mv kustomize /usr/local/bin/
# Build manifestus cli
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build

# Final image
FROM debian:bookworm-slim
RUN apt update && apt upgrade -y
COPY --from=builder /usr/local/bin/ /usr/local/bin/
COPY --from=builder /app/bin/manifestus /usr/local/bin/manifestus
RUN useradd -u 1001 -ms /bin/bash app
USER app
WORKDIR /home/app
USER 1001:1001
ENTRYPOINT ["/usr/local/bin/manifestus"]
