# Build stage
FROM eclipse-temurin:21.0.2_13-jdk-jammy AS builder

RUN mkdir /opt/app
COPY . /opt/app

WORKDIR /opt/app
RUN /opt/app/gradlew clean build

# Build the actual runtime image
FROM eclipse-temurin:21.0.2_13-jdk-jammy

# Run updates to ensure latest OS security patches are applied
# and clean up after installing updates to reduce image size
RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# STEP 1: Add the executable jar file in this container
# Follow standards from readme on https://hub.docker.com/_/eclipse-temurin
RUN mkdir /opt/app
COPY --from=builder /opt/app/build/libs/helloworld.jar /opt/app/helloworld.jar
# STEP 2: Run the application when the container starts
# Use entrypoint as it allows adding arguments when running the container.

# Create a non-root user to run the application
RUN useradd -ms /bin/bash helloworlduser
USER helloworlduser

# Open port 8080
EXPOSE 8080

# Healthcheck to verify the application is running
HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:8080/ || exit 1

# Use entrypoint to run the jar so arguments can be passed at runtime
ENTRYPOINT ["java", "-jar", "/opt/app/helloworld.jar"]
