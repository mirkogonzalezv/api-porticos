# ATLAS Seed - dockerfile
FROM golang:1.25.5-alpine


# Instalar dependencias necesarias
RUN apk update && apk add --no-cache ca-certificates git

#Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias primero
COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Compilación estática
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o rea/porticos ./cmd/api


# Exponer puerto
EXPOSE 4200

# Comando por defecto
CMD ["./api_porticos"]