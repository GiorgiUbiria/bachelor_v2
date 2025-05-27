# Test script for bachelor project services
Write-Host "Testing ML-Powered E-Commerce Application" -ForegroundColor Green

# Check if Docker is running
try {
    docker --version | Out-Null
    Write-Host "✓ Docker is available" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker is not available" -ForegroundColor Red
    exit 1
}

# Test database connection
Write-Host "`nTesting PostgreSQL database..." -ForegroundColor Yellow
docker run --rm -d --name test-postgres -e POSTGRES_DB=bachelor_db -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres123 -p 5432:5432 postgres:15-alpine

Start-Sleep 10

# Test if database is ready
$dbReady = $false
for ($i = 0; $i -lt 30; $i++) {
    try {
        docker exec test-postgres psql -U postgres -d bachelor_db -c "SELECT 1;" | Out-Null
        $dbReady = $true
        break
    } catch {
        Start-Sleep 2
    }
}

if ($dbReady) {
    Write-Host "✓ PostgreSQL database is running" -ForegroundColor Green
} else {
    Write-Host "✗ PostgreSQL database failed to start" -ForegroundColor Red
    docker stop test-postgres
    exit 1
}

Write-Host "`nApplication is ready for testing!" -ForegroundColor Green
Write-Host "You can now run 'docker-compose up --build' to start all services" -ForegroundColor Cyan

# Cleanup
docker stop test-postgres 