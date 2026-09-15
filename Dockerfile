FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /bin/battlesnake .

FROM gcr.io/distroless/static-debian12
COPY --from=build /bin/battlesnake /battlesnake
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/battlesnake"]
