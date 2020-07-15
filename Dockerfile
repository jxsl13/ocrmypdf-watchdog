FROM golang:latest as build

LABEL maintainer "jxsl13@gmail.com"
WORKDIR /build
COPY *.go ./
COPY go.mod .
COPY go.sum .
RUN go get -d && \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -extldflags "-static"' -o main .


FROM jbarlow83/ocrmypdf:latest as ocrmypdf
ENV PUID "1000"
ENV PGID "100"
ENV OCRMYPDF_ARGS --pdf-renderer sandwich --tesseract-timeout 1800 --rotate-pages -l eng+fra+deu --deskew --clean --skip-text
WORKDIR /app
COPY --from=build /build/main .
VOLUME ["/in", "/out"]
ENTRYPOINT ["/app/main"]


