FROM microsoft/aspnetcore-build:1.1
WORKDIR /app
ENV PORT 80
EXPOSE 80
COPY . .
RUN dotnet restore
RUN dotnet publish --output /out/ --configuration Release /maxcpucount
RUN sed -n 's:.*<AssemblyName>\(.*\)</AssemblyName>.*:\1:p' *.csproj > __assemblyname
RUN if [ ! -s __assemblyname ]; then filename=$(ls *.csproj); echo ${filename%.*} > __assemblyname; fi
WORKDIR /out
ENTRYPOINT dotnet /out/$(cat /app/__assemblyname).dll
