# --- Build stage ---
FROM golang:1.25.1-alpine3.22 AS build

WORKDIR /src

COPY data_interface_source/ ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o data_interface ./main.go


# --- Runtime stage ---
FROM almalinux:9.7-minimal

WORKDIR /interface

COPY --from=build /src/data_interface .

EXPOSE 9876

ENTRYPOINT ["./data_interface"]
