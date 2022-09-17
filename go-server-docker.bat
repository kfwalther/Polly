:: Build the go-server Docker image.
docker build -t go-server -f go-server.Dockerfile .

:: Run the go-server Docker container.
docker run -i -p 22 -p 5000 go-server 