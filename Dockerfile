### Description: Dockerfile for helm-diff-summary
FROM alpine:3.23

COPY helm-diff-summary /

# Starting
ENTRYPOINT [ "/helm-diff-summary" ]