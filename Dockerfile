FROM scratch
COPY nsq_to_dogstatsd /
ENTRYPOINT ["/nsq_to_dogstatsd"]