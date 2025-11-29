FROM eclipse-temurin:17-jdk-alpine AS build
WORKDIR /app

# Copy maven wrapper and pom
COPY pom.xml .
COPY .mvn .mvn
COPY mvnw .

# Download dependencies
RUN chmod +x mvnw && ./mvnw dependency:go-offline -B

# Copy source code
COPY src src

# Build the application
RUN ./mvnw package -DskipTests -B

FROM eclipse-temurin:17-jre-alpine
WORKDIR /app

# Copy the jar file
COPY --from=build /app/target/*.jar app.jar

# Expose the service port
EXPOSE 8086

# Run the application
ENTRYPOINT ["java", "-jar", "app.jar"]
