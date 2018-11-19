FROM alpine:3.8

ARG ARCH
ARG TINI_VERSION

ADD rook /usr/local/bin/

# Add files for the sidecar
RUN mkdir -p /sidecar
RUN mkdir -p /sidecar/plugins

ADD rook /sidecar/
# Jolokia plugin for JMX<->HTTP
ADD "http://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.6.0/jolokia-jvm-1.6.0-agent.jar" /sidecar/plugins/jolokia.jar

# Run tini as PID 1 and avoid signal handling issues
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-static-${ARCH} /sidecar/tini
RUN chmod +x /sidecar/tini && chmod +x /usr/local/bin/rook


ENTRYPOINT ["/sidecar/tini", "--", "/usr/local/bin/rook"]
CMD [""]
