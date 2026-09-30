FROM --platform=${TARGETPLATFORM:-linux/amd64} gcr.io/distroless/static

ENV TZ Europe/Brussels

COPY build/webapp .
      
ENTRYPOINT ["./webapp"]