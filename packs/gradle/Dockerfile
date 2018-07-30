FROM gradle:jdk10 as BUILD

COPY . /project
RUN gradle -b /project/build.gradle clean build

FROM openjdk:10-jre-slim
ENV PORT 4567
EXPOSE 4567
COPY --from=BUILD /project/target/app.jar /opt/app.jar
WORKDIR /opt
CMD ["java", "-jar", "app.jar"]
