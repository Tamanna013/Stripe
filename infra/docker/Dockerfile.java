# Multi-stage build for Java services
FROM maven:3.9.6-eclipse-temurin-21-alpine AS builder

ARG SERVICE_NAME

WORKDIR /src
COPY services/${SERVICE_NAME}/pom.xml .
RUN mvn dependency:go-offline -B

COPY services/${SERVICE_NAME}/src ./src
RUN mvn clean package -DskipTests

# Runtime stage
FROM eclipse-temurin:21-jre-alpine AS runtime

ARG SERVICE_NAME

WORKDIR /app
COPY --from=builder /src/target/*.jar app.jar

# Run as non-root
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
USER nonroot:nonroot

ENTRYPOINT ["java", "-jar", "app.jar"]
