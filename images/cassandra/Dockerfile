FROM alpine:3.8

ADD rook /usr/local/bin/

ADD http://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.6.0/jolokia-jvm-1.6.0-agent.jar /jolokia.jar

RUN chmod a+r /jolokia.jar && touch /jolokia.jar

ENTRYPOINT ["/usr/local/bin/rook"]
CMD [""]
