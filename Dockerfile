# CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway .
FROM scratch
ADD gateway /
CMD ["/gateway"]
