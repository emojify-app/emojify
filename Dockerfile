FROM scratch

COPY emojify /

ENTRYPOINT ["/emojify"]
