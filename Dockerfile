### Description: Dockerfile for helm-diff-summary
FROM alpine:3.23

ARG TARGETOS
ARG TARGETARCH

COPY ${TARGETOS}/${TARGETARCH}/helm-diff-summary /

# Starting
ENTRYPOINT [ "/helm-diff-summary" ]