# Build from a golang image.
FROM golang:latest as go_builder

RUN mkdir /polly

WORKDIR /polly

# Copy in the go-specific config, and JSON configs.
COPY go.* ./
COPY *.json ./
COPY portfolio-sheet-id.txt ./
RUN go mod download

# Copy in the go source files.
COPY ./backend ./backend

# Install Python3 and download/install the packages.
RUN apt-get update && \
    apt-get -y install python3 && \
    apt-get -y install python3-setuptools && \
    apt-get -y install python3-pip

# Download and install MongoDB.
# RUN apt-get install -y mongodb-org

# RUN mongod

# Build the go server.
RUN cd backend && \
    go build -v -o go-server.exe

# Update working directory to executable path.
WORKDIR /polly/backend
# Copy executable to top-level directory.
CMD ["./go-server.exe"]
