:: Build the react-client Docker image.
docker build -t react-client -f react-client.Dockerfile .

:: Run the react-client Docker container.
docker run -it -p 3001:3000 react-client
