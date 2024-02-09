FROM alpine:3.19

COPY fml /usr/local/bin/

VOLUME /data
ENV "FML_LISTEN_ADDRESS" ":5000"
ENV "FML_DATABASE_URI" "sqlite:///data/fasttrackml.db"
ENTRYPOINT ["fml"]
CMD ["server"]
