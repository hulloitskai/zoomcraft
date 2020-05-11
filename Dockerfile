# == BACKEND ==

FROM golang:1.14.2-alpine AS backend

WORKDIR /src
COPY ./backend/ ./
RUN mkdir /dist && go build -o /dist/backend .

# == CLIENT ==

FROM node:14.2.0-alpine AS client

WORKDIR /src
COPY ./client/ ./
RUN yarn install && yarn build && mv ./build/ /dist/

# == GATEWAY ==

FROM node:14.2.0-alpine AS gateway

WORKDIR /app

COPY --from=backend /dist/ ./backend/
COPY --from=client /dist/ ./client/
COPY ./gateway ./gateway
COPY ./scripts/entrypoint.sh entrypoint.sh

RUN apk add parallel
RUN cd ./gateway && yarn install --production

ENV GATEWAY_PORT=8080 BACKEND_PORT=9090 CLIENT_PATH=/app/client
EXPOSE 8080
ENTRYPOINT ["/app/entrypoint.sh"]
