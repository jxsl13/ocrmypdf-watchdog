FROM golang:latest as build

LABEL maintainer "jxsl13@gmail.com"
WORKDIR /build
COPY . .
RUN go get -d && \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -extldflags "-static"' -o watcher .


FROM jbarlow83/ocrmypdf:latest as ocrmypdf

ENV OCRMYPDF_ARGS --pdf-renderer sandwich --tesseract-timeout 1800 --rotate-pages -l eng+fra+deu --deskew --clean --skip-text

WORKDIR /app

COPY --from=build /build/watcher .

VOLUME ["/in", "/out"]

ENTRYPOINT ["/app/watcher"]


