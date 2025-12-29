# Tridorian ZTNA - Docker Build and Push Script
# PowerShell version for Windows development

param(
    [string]$ProjectId = "trivpn-demo-prj",
    [string]$Region = "asia-southeast1",
    [string]$RepoName = "triztna",
    [string]$ImageTag = "latest",
    [switch]$BuildOnly,
    [switch]$PushOnly,
    [string]$Service = "all"
)

$ErrorActionPreference = "Stop"

# Configuration
$Registry = "$Region-docker.pkg.dev"
$BaseImage = "$Registry/$ProjectId/$RepoName"

# Image names
$Images = @{
    "management-api" = @{
        "Image" = "$BaseImage/management-api:$ImageTag"
        "Dockerfile" = "Dockerfile.management-api"
        "Context" = "."
    }
    "gateway-controlplane" = @{
        "Image" = "$BaseImage/gateway-controlplane:$ImageTag"
        "Dockerfile" = "Dockerfile.gateway-controlplane"
        "Context" = "."
    }
    "auth-api" = @{
        "Image" = "$BaseImage/auth-api:$ImageTag"
        "Dockerfile" = "Dockerfile.auth-api"
        "Context" = "."
    }
    "tenant-admin" = @{
        "Image" = "$BaseImage/tenant-admin:$ImageTag"
        "Dockerfile" = "Dockerfile.tenant-admin"
        "Context" = "."
    }
    "backoffice" = @{
        "Image" = "$BaseImage/backoffice:$ImageTag"
        "Dockerfile" = "Dockerfile.backoffice"
        "Context" = "."
    }
}

function Write-Header {
    param([string]$Message)
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host $Message -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Error-Msg {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Build-Image {
    param(
        [string]$Name,
        [hashtable]$Config
    )
    
    Write-Info "Building $Name..."
    
    $dockerfile = $Config.Dockerfile
    $context = $Config.Context
    $image = $Config.Image
    
    try {
        docker build --no-cache -t $image -f $dockerfile $context
        if ($LASTEXITCODE -eq 0) {
            Write-Success "$Name built successfully"
            return $true
        } else {
            Write-Error-Msg "$Name build failed"
            return $false
        }
    } catch {
        Write-Error-Msg "Failed to build $Name : $_"
        return $false
    }
}

function Push-Image {
    param(
        [string]$Name,
        [string]$Image
    )
    
    Write-Info "Pushing $Name..."
    
    try {
        docker push $Image
        if ($LASTEXITCODE -eq 0) {
            Write-Success "$Name pushed successfully"
            return $true
        } else {
            Write-Error-Msg "$Name push failed"
            return $false
        }
    } catch {
        Write-Error-Msg "Failed to push $Name : $_"
        return $false
    }
}

function Setup-ArtifactRegistry {
    Write-Header "Setting up Artifact Registry"
    
    Write-Info "Creating repository..."
    gcloud artifacts repositories create $RepoName `
        --repository-format=docker `
        --location=$Region `
        --description="Tridorian ZTNA Docker images" `
        --project=$ProjectId
    
    if ($LASTEXITCODE -eq 0 -or $LASTEXITCODE -eq 1) {
        Write-Success "Repository ready"
    }
    
    Write-Info "Configuring Docker authentication..."
    gcloud auth configure-docker "$Region-docker.pkg.dev"
    
    Write-Success "Artifact Registry setup complete"
}

function Show-Configuration {
    Write-Header "Configuration"
    Write-Host "Project ID:    $ProjectId"
    Write-Host "Region:        $Region"
    Write-Host "Repository:    $RepoName"
    Write-Host "Image Tag:     $ImageTag"
    Write-Host "Service:       $Service"
    Write-Host ""
}

# Main execution
Write-Header "Tridorian ZTNA - Docker Build and Push"
Show-Configuration

# Check Docker
Write-Info "Checking Docker..."
try {
    docker --version | Out-Null
    Write-Success "Docker is available"
} catch {
    Write-Error-Msg "Docker is not installed or not running"
    Write-Host "Please install Docker Desktop and ensure it's running"
    exit 1
}

# Check gcloud
Write-Info "Checking gcloud..."
try {
    gcloud --version | Out-Null
    Write-Success "gcloud is available"
} catch {
    Write-Error-Msg "gcloud is not installed"
    Write-Host "Please install Google Cloud SDK"
    exit 1
}

# Determine which services to process
$ServicesToProcess = @()
if ($Service -eq "all") {
    $ServicesToProcess = $Images.Keys
} elseif ($Images.ContainsKey($Service)) {
    $ServicesToProcess = @($Service)
} else {
    $available = $Images.Keys -join ", "
    Write-Error-Msg "Unknown service: $Service"
    Write-Host "Available services: $available"
    exit 1
}

# Build phase
if (-not $PushOnly) {
    Write-Header "Building Images"
    
    $buildResults = @{}
    foreach ($svc in $ServicesToProcess) {
        $config = $Images[$svc]
        $result = Build-Image -Name $svc -Config $config
        $buildResults[$svc] = $result
    }
    
    # Summary
    Write-Host ""
    Write-Host "Build Summary:" -ForegroundColor Cyan
    foreach ($svc in $buildResults.Keys) {
        if ($buildResults[$svc]) {
            Write-Success "$svc"
        } else {
            Write-Error-Msg "$svc"
        }
    }
    
    # Check if any builds failed
    $hasFailed = $false
    foreach ($val in $buildResults.Values) {
        if (-not $val) { $hasFailed = $true }
    }
    
    if ($hasFailed) {
        Write-Error-Msg "Some builds failed. Aborting push."
        exit 1
    }
}

# Push phase
if (-not $BuildOnly) {
    Write-Header "Pushing Images"
    
    $pushResults = @{}
    foreach ($svc in $ServicesToProcess) {
        $image = $Images[$svc].Image
        $result = Push-Image -Name $svc -Image $image
        $pushResults[$svc] = $result
    }
    
    # Summary
    Write-Host ""
    Write-Host "Push Summary:" -ForegroundColor Cyan
    foreach ($svc in $pushResults.Keys) {
        if ($pushResults[$svc]) {
            Write-Success "$svc"
        } else {
            Write-Error-Msg "$svc"
        }
    }
    
    # Check if any pushes failed
    $hasFailedPush = $false
    foreach ($val in $pushResults.Values) {
        if (-not $val) { $hasFailedPush = $true }
    }
    if ($hasFailedPush) {
        Write-Error-Msg "Some pushes failed."
        exit 1
    }
}

Write-Header "Complete!"
Write-Success "All operations completed successfully"

# Show next steps
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "1. Verify images in Artifact Registry:"
Write-Host "   gcloud artifacts docker images list $Registry/$ProjectId/$RepoName"
Write-Host ""
Write-Host "2. Deploy to GKE:"
Write-Host "   kubectl apply -k k8s/overlays/prod/"
Write-Host ""
Write-Host "3. Check deployment status:"
Write-Host "   kubectl get pods -n tridorian-ztna"
Write-Host ""
