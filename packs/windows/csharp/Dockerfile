# escape=`

# WCF Server 1803
FROM microsoft/dotnet-framework:4.7.2-sdk-windowsservercore-1803 AS builder
WORKDIR /app

# Build and package dotnet
COPY . .
RUN nuget restore
RUN msbuild /p:Configuration=Release /p:DeployOnBuild=true /p:PublishProfile="FolderProfile.pubxml"

# Multi stage build
FROM microsoft/dotnet-framework:4.7.2-runtime-windowsservercore-1803 as runtime
WORKDIR /inetpub/wwwroot
COPY --from=builder /app/*/Release/Publish ./

ENV PORT 80
EXPOSE 80
