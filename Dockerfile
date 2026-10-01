FROM scratch
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/nsq_to_dogstatsd /
ENTRYPOINT ["/nsq_to_dogstatsd"]
