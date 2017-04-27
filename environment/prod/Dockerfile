FROM themotion/ladder_base:latest

USER root

# Prepare
RUN mkdir -p /bin
RUN mkdir -p /etc/ladder

# Build
RUN ./build.sh /bin/ladder
RUN chmod 755 /bin/ladder

# Clean up
WORKDIR /
RUN rm -rf /go/src/*

# When building image copy config if present
ONBUILD COPY ladder.yml /etc/ladder/ladder.yml
ONBUILD COPY cfg-autoscalers /etc/ladder/cfg-autoscalers
RUN chown -R ladder:ladder /etc/ladder/

USER ladder

ENTRYPOINT [ "/bin/ladder" ]
CMD        [ "-config.file=/etc/ladder/ladder.yml -json.log"]