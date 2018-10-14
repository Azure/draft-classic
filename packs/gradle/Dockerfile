FROM gradle:jdk11 as BUILD

COPY --chown=gradle:gradle . /project
RUN gradle -i -s -b /project/build.gradle clean installDist && \
    rm -rf /project/build/install/*/bin/*.bat

FROM openjdk:11-jre-slim
ENV PORT 4567
EXPOSE 4567
COPY --from=BUILD /project/build/install/* /opt/
WORKDIR /opt/bin
CMD ["/bin/bash", "-c", "find -type f -name '*' | xargs bash"]
