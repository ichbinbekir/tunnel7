FROM golang:latest

WORKDIR /root
COPY . .

RUN make build


FROM scratch

WORKDIR /root
COPY --from=0 /root/bin/tunnel7 .

ENTRYPOINT [ "/root/tunnel7" ]
CMD [ "-t", "/data/tunnels.json" ]
