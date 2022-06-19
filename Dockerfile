FROM scratch
COPY hfs /
ENTRYPOINT ["/hfs", "serve"]
