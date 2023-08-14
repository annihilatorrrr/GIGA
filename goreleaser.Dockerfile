FROM golang:1.21
COPY giga changelog.json /app/
CMD ["/app/giga"]
