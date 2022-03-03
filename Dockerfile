FROM registry.access.redhat.com/ubi8/go-toolset:1.16.7 as builder

WORKDIR /workspace

COPY go.mod .
COPY go.sum .

USER 0

RUN go mod download

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o nanokube main.go

FROM registry.access.redhat.com/ubi8/go-toolset:1.16.7

USER 0

WORKDIR /workspace

RUN go get sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

RUN /opt/app-root/src/go/bin/setup-envtest use 1.21 --bin-dir /tmp -p path > kpath

RUN mkdir -p /opt/app-root/src/.cache/kubebuilder-envtest && chmod 775 /opt/app-root/src/.cache/kubebuilder-envtest

RUN mkdir -p static

COPY run.sh .

COPY static/appstudio.redhat.com_applications.yaml static
COPY static/appstudio.redhat.com_components.yaml static
COPY static/project.yaml static

COPY --from=builder /workspace/nanokube .

ENTRYPOINT [ "/bin/sh", "/workspace/run.sh" ]
