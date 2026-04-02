FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS base
WORKDIR /app
EXPOSE 8050

FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src
COPY ["Backup.Server/Backup.Server.csproj", "Backup.Server/"]
RUN dotnet restore "Backup.Server/Backup.Server.csproj"
COPY Backup.Server/ ./Backup.Server/
COPY ../protos/ ./protos/
RUN dotnet build "Backup.Server/Backup.Server.csproj" -c Release -o /app/build
FROM build AS publish
RUN dotnet publish "Backup.Server/Backup.Server.csproj" -c Release -o /app/publish /p:UseAppHost=false

FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .
ENV ASPNETCORE_URLS=http://+:8050
ENTRYPOINT ["dotnet", "Backup.Server.dll"]
