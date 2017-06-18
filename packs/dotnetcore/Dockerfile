FROM microsoft/aspnetcore-build:1.1
WORKDIR /app
EXPOSE 80
COPY . .
RUN dotnet restore
RUN dotnet publish --output /out/ --configuration Release /maxcpucount
ENTRYPOINT ["dotnet", "/out/WebApplication.dll"]
