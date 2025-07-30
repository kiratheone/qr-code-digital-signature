#!/bin/bash

# Optimized Docker build script for production
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
BUILD_ARGS=""
PARALLEL_BUILDS=true
CACHE_FROM_REGISTRY=false
REGISTRY_URL=""
PUSH_TO_REGISTRY=false
TARGET_PLATFORM=""

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -s, --service SERVICE   Build specific service only"
    echo "  -p, --parallel          Enable parallel builds (default: true)"
    echo "  -c, --cache-from REG    Pull cache from registry"
    echo "  -r, --registry URL      Registry URL for caching"
    echo "  --platform PLATFORM     Target platform (default: auto-detect)"
    echo "  --multi-platform        Build for multiple platforms"
    echo "  --push                  Push images to registry after build"
    echo "  --no-cache              Build without using cache"
    echo "  --squash                Squash layers (experimental)"
    echo ""
    echo "Platform examples:"
    echo "  linux/amd64            x86_64 architecture"
    echo "  linux/arm64            ARM64 architecture"
    echo "  linux/arm/v7           ARMv7 architecture"
    echo ""
    echo "Examples:"
    echo "  $0                      Build all services (auto-detect platform)"
    echo "  $0 -s backend           Build only backend service"
    echo "  --platform linux/amd64  Build for specific platform"
    echo "  $0 --multi-platform --push Build for multiple platforms and push"
    echo "  $0 --no-cache           Build without cache"
    echo "  $0 -r registry.io --push Build and push to registry"
}

get_build_args() {
    local version=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
    local build_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local git_commit=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    
    BUILD_ARGS="--build-arg VERSION=${version} --build-arg BUILD_TIME=${build_time} --build-arg GIT_COMMIT=${git_commit}"
    
    log_debug "Build arguments: VERSION=${version}, BUILD_TIME=${build_time}, GIT_COMMIT=${git_commit:0:8}"
}

# Add function to detect current platform
detect_platform() {
    local arch=$(uname -m)
    local os="linux"
    
    case "$arch" in
        x86_64)
            echo "${os}/amd64"
            ;;
        aarch64|arm64)
            echo "${os}/arm64"
            ;;
        armv7l)
            echo "${os}/arm/v7"
            ;;
        *)
            log_warn "Unknown architecture: $arch, defaulting to linux/amd64"
            echo "${os}/amd64"
            ;;
    esac
}

setup_buildx() {
    log_info "Setting up Docker Buildx..."
    
    # Create buildx builder if it doesn't exist
    if ! docker buildx inspect multiarch >/dev/null 2>&1; then
        docker buildx create --name multiarch --driver docker-container --use
        docker buildx inspect --bootstrap
    else
        docker buildx use multiarch
    fi
    
    log_info "Docker Buildx setup completed"
}

build_service() {
    local service=$1
    local dockerfile_path=""
    local context_path=""
    local image_name=""
    
    case "$service" in
        "frontend")
            dockerfile_path="./frontend/Dockerfile.prod"
            context_path="./frontend"
            image_name="qds-frontend"
            ;;
        "backend")
            dockerfile_path="./backend/Dockerfile.prod"
            context_path="./backend"
            image_name="qds-backend"
            ;;
        *)
            log_error "Unknown service: $service"
            return 1
            ;;
    esac
    
    log_info "Building $service..."
    
    local build_cmd="docker buildx build"
    
    # Add build arguments
    build_cmd="$build_cmd $BUILD_ARGS"
    
    # Add cache options
    if [ "$CACHE_FROM_REGISTRY" = true ] && [ -n "$REGISTRY_URL" ]; then
        build_cmd="$build_cmd --cache-from type=registry,ref=${REGISTRY_URL}/${image_name}:cache"
        build_cmd="$build_cmd --cache-to type=registry,ref=${REGISTRY_URL}/${image_name}:cache,mode=max"
    else
        build_cmd="$build_cmd --cache-from type=local,src=/tmp/.buildx-cache-${image_name}"
        build_cmd="$build_cmd --cache-to type=local,dest=/tmp/.buildx-cache-${image_name},mode=max"
    fi
    
    # Add platform support
    local platform=""
    if [ -n "$TARGET_PLATFORM" ]; then
        platform="$TARGET_PLATFORM"
    else
        platform=$(detect_platform)
    fi
    
    build_cmd="$build_cmd --platform $platform"
    log_debug "Building for platform: $platform"
    
    # Add tags
    local tag="${image_name}:latest"
    if [ -n "$REGISTRY_URL" ]; then
        tag="${REGISTRY_URL}/${tag}"
    fi
    build_cmd="$build_cmd -t $tag"
    
    # For multi-platform builds, must push to registry
    if [[ "$platform" == *","* ]]; then
        if [ "$PUSH_TO_REGISTRY" = false ]; then
            log_warn "Multi-platform build detected, forcing push to registry"
            PUSH_TO_REGISTRY=true
        fi
    fi
    
    # Add push option
    if [ "$PUSH_TO_REGISTRY" = true ]; then
        build_cmd="$build_cmd --push"
    else
        build_cmd="$build_cmd --load"
    fi
    
    # Add dockerfile and context
    build_cmd="$build_cmd -f $dockerfile_path $context_path"
    
    log_debug "Build command: $build_cmd"
    
    # Execute build
    if eval "$build_cmd"; then
        log_info "$service build completed successfully"
        return 0
    else
        log_error "$service build failed"
        return 1
    fi
}

build_all_services() {
    log_info "Building all services..."
    
    local services=("backend" "frontend")
    local failed_builds=()
    
    if [ "$PARALLEL_BUILDS" = true ]; then
        log_info "Building services in parallel..."
        
        # Start builds in background
        local pids=()
        for service in "${services[@]}"; do
            build_service "$service" &
            pids+=($!)
        done
        
        # Wait for all builds to complete
        local failed=false
        for i in "${!pids[@]}"; do
            if ! wait "${pids[$i]}"; then
                failed_builds+=("${services[$i]}")
                failed=true
            fi
        done
        
        if [ "$failed" = true ]; then
            log_error "Failed to build services: ${failed_builds[*]}"
            return 1
        fi
    else
        log_info "Building services sequentially..."
        
        for service in "${services[@]}"; do
            if ! build_service "$service"; then
                failed_builds+=("$service")
            fi
        done
        
        if [ ${#failed_builds[@]} -gt 0 ]; then
            log_error "Failed to build services: ${failed_builds[*]}"
            return 1
        fi
    fi
    
    log_info "All services built successfully"
}

optimize_images() {
    log_info "Optimizing Docker images..."
    
    # Remove dangling images
    docker image prune -f
    
    # Show image sizes
    log_info "Image sizes:"
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | grep -E "(qds-frontend|qds-backend)"
    
    log_info "Image optimization completed"
}

run_security_scan() {
    log_info "Running security scan on images..."
    
    local services=("frontend" "backend")
    local image_names=("qds-frontend" "qds-backend")
    
    for i in "${!services[@]}"; do
        local service="${services[$i]}"
        local image_name="${image_names[$i]}"
        
        if command -v trivy &> /dev/null; then
            log_info "Scanning $service with Trivy..."
            trivy image --exit-code 1 --severity HIGH,CRITICAL "${image_name}:latest" || log_warn "$service has security vulnerabilities"
        elif command -v docker &> /dev/null && docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy:latest --version &> /dev/null; then
            log_info "Scanning $service with Trivy (Docker)..."
            docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy:latest image --exit-code 1 --severity HIGH,CRITICAL "${image_name}:latest" || log_warn "$service has security vulnerabilities"
        else
            log_warn "Trivy not available, skipping security scan"
            break
        fi
    done
    
    log_info "Security scan completed"
}

# Parse command line arguments
SPECIFIC_SERVICE=""
NO_CACHE=false
SQUASH=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -s|--service)
            SPECIFIC_SERVICE="$2"
            shift 2
            ;;
        -p|--parallel)
            PARALLEL_BUILDS=true
            shift
            ;;
        -c|--cache-from)
            CACHE_FROM_REGISTRY=true
            REGISTRY_URL="$2"
            shift 2
            ;;
        -r|--registry)
            REGISTRY_URL="$2"
            shift 2
            ;;
        --platform)
            TARGET_PLATFORM="$2"
            shift 2
            ;;
        --multi-platform)
            TARGET_PLATFORM="linux/amd64,linux/arm64"
            shift
            ;;
        --push)
            PUSH_TO_REGISTRY=true
            shift
            ;;
        --no-cache)
            NO_CACHE=true
            shift
            ;;
        --squash)
            SQUASH=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    log_info "Starting optimized Docker build process..."
    
    # Get build arguments
    get_build_args
    
    # Setup buildx for multi-platform builds
    setup_buildx
    
    # Add no-cache option if specified
    if [ "$NO_CACHE" = true ]; then
        BUILD_ARGS="$BUILD_ARGS --no-cache"
    fi
    
    # Add squash option if specified (experimental)
    if [ "$SQUASH" = true ]; then
        BUILD_ARGS="$BUILD_ARGS --squash"
    fi
    
    # Build specific service or all services
    if [ -n "$SPECIFIC_SERVICE" ]; then
        build_service "$SPECIFIC_SERVICE"
    else
        build_all_services
    fi
    
    # Optimize images
    optimize_images
    
    # Run security scan
    run_security_scan
    
    log_info "Build process completed successfully!"
}

# Check requirements
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed"
    exit 1
fi

if ! command -v git &> /dev/null; then
    log_warn "Git is not installed, version info will be limited"
fi

main