FROM hub.oneitfarm.com/library/node:12 AS node-builder
WORKDIR /app
COPY react .
RUN yarn config set registry 'https://registry.npm.taobao.org'
RUN yarn && yarn run build

FROM harbor.oneitfarm.com/zhirenyun/go:1.15.6 AS go-builder
WORKDIR /app
ENV GOPROXY=https://goproxy.oneitfarm.com,https://goproxy.cn,direct
ADD go.* ./
ADD *.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w" -a -o serve main.go

FROM hub.oneitfarm.com/library/alpine
WORKDIR /workspace
COPY --from=node-builder /app/build build
COPY --from=go-builder /app/serve .
ENTRYPOINT ["/workspace/serve"]