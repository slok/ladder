FROM busybox

# Prepare
WORKDIR /
RUN mkdir -p /bin
RUN mkdir -p /etc/ladder

# put ladder
COPY ./bin/ladder /bin/ladder
RUN chmod 755 /bin/ladder

# When building image copy config if present
ONBUILD COPY ladder.yml /etc/ladder/ladder.yml
ONBUILD COPY cfg-autoscalers /etc/ladder/cfg-autoscalers

EXPOSE 9094

ENTRYPOINT [ "/bin/ladder" ]
CMD        [ "-config.file=/etc/ladder/ladder.yml -json.log"]