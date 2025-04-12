# Azure Go Application

This is a simple Go application that outputs "Hello, World!" when run. It is designed to be deployed to Azure Web Applications.

## Prerequisites

- Go installed on your machine
- Azure account
- Azure CLI installed

## Building the Application

1. Clone the repository:
   ```
   git clone <repository-url>
   cd azure-go-app
   ```

2. Build the application:
   ```
   go build -o hello-world ./cmd
   ```

## Running the Application Locally

To run the application locally, execute the following command:
```
./hello-world
```
You should see the output:
```
Hello, World!
```

## Deploying to Azure

1. Log in to your Azure account:
   ```
   az login
   ```

2. Create a new Azure Web App:
   ```
   az webapp create --resource-group <resource-group-name> --plan <app-service-plan-name> --name <app-name> --runtime "GO|1.16"
   ```

3. Deploy the application:
   ```
   az webapp deploy --resource-group <resource-group-name> --name <app-name> --src-path ./hello-world
   ```

4. Access your application at:
   ```
   https://<app-name>.azurewebsites.net
   ```

You should see the output:
```
Hello, World!
```

## License

This project is licensed under the MIT License.