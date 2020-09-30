# create the build container
FROM golang:1.15.2 AS build
LABEL stage=build
WORKDIR /app

# copy everything
COPY . /app/

# build static
ENV CGO_ENABLED 0
ENV GOOS linux
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o server .

# create the runtime container
FROM scratch
WORKDIR /app
COPY --from=build /app/server .
COPY --from=build /app/assets/ ./assets/
ENTRYPOINT ["./server"]