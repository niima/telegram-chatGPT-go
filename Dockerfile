FROM golang:1.19-alpine

# Set the Current Working Directory inside the container
WORKDIR /opt/

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go mod download

# Install the package
RUN go build -o chatgptgo

# Run the executable
CMD ["/opt/chatgptgo"]