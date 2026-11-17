FROM golang:1.27.4-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/app .

FROM gcr.io/distroless/static-debian12:nonroot
ARG VERSION=dev
ENV APP_VERSION=${VERSION}
COPY --from=build /out/app /app
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app"]
