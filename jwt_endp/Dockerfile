FROM scratch
ADD ca-certificates.crt /etc/ssl/certs/
ADD jwt_endp /
ENV GIN_MODE=release
CMD ["/jwt_endp"]
