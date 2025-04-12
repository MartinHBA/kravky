# Define the API endpoint
$apiUrl = "http://localhost:8080/fetch-data"

# Send a POST request to the API
$response = Invoke-RestMethod -Uri $apiUrl -Method Post

# Output the response
Write-Host "Response from API:"
$response