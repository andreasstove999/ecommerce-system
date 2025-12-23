FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src

# Copy csproj and restore as distinct layers
COPY ./src/PaymentService/PaymentService.csproj ./PaymentService/
RUN dotnet restore ./PaymentService/PaymentService.csproj

# Copy the remaining source
COPY ./src/PaymentService ./PaymentService
WORKDIR /src/PaymentService

RUN dotnet publish -c Release -o /out --no-restore

FROM mcr.microsoft.com/dotnet/aspnet:10.0
WORKDIR /app
COPY --from=build /out .

ENV ASPNETCORE_URLS=http://+:8080
EXPOSE 8080

ENTRYPOINT ["dotnet", "PaymentService.dll"]
