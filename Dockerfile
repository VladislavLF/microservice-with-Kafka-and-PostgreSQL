FROM golang:1.24.5 as build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/ordersvc ./cmd/ordersvc

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /bin/ordersvc /app/ordersvc
COPY configs/ /app/configs/
COPY web/ /app/web/
COPY migrations/ /app/migrations/
EXPOSE 8081
USER 65532:65532
ENTRYPOINT ["/app/ordersvc"]
