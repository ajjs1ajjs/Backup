FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS base
WORKDIR /app
EXPOSE 8050
EXPOSE 8080

FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src
COPY ["src/server/Backup.Server/Backup.Server.csproj", "src/server/Backup.Server/"]
RUN dotnet restore "src/server/Backup.Server/Backup.Server.csproj"
COPY src/ .
RUN dotnet build "src/server/Backup.Server/Backup.Server.csproj" -c Release -o /app/build
FROM build AS publish
RUN dotnet publish "src/server/Backup.Server/Backup.Server.csproj" -c Release -o /app/publish /p:UseAppHost=false

FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .
ENTRYPOINT ["dotnet", "Backup.Server.dll"]
