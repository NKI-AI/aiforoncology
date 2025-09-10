# SlideInsight Python SDK

A modern, Pythonic SDK for the SlideInsight API with comprehensive support for studies, cases, slides, annotations, and user management.

## Features

- 🚀 **Modern async/await API** - Built with asyncio for high performance
- 📊 **Full pagination support** - Automatic pagination, async iteration, and manual control
- 🔒 **Automatic authentication** - JWT token management with automatic refresh
- 🛡️ **Type safety** - Full typing support with modern Python 3.11+ syntax
- 📦 **Resource managers** - Organized API access for different resources
- 🎯 **Error handling** - Comprehensive exception hierarchy with proper HTTP status mapping
- 🔄 **Retry logic** - Automatic retry with exponential backoff
- 📋 **CLI tools** - Modern command-line interface with rich formatting

## Installation

```bash
pip install -r requirements.txt
```

## Quick Start

### Basic Usage

```python
import asyncio
from slideinsight_sdk import SlideInsightClient

async def main():
    # Using async context manager (recommended)
    async with SlideInsightClient("http://localhost:3000") as client:
        # Authenticate
        await client.login("email", "password")

        # List studies with pagination
        studies = await client.studies.list(page=1, limit=10)
        print(f"Found {studies.pagination.total} studies")

        # Create a new study
        study = await client.studies.create(
            name="My Study",
            description="Study created via SDK"
        )

        # Create a case and add to study
        case = await client.cases.create(name="My Case")
        await client.studies.add_case(study.study_uid, case.case_uid)

        # Add a slide to the case
        slide = await client.cases.add_slide(
            case_uid=case.case_uid,
            slide_uri="/path/to/slide.svs",
            slide_name="My Slide"
        )

        # Add annotations
        raster_annotation = await client.slides.add_raster_annotation(
            slide_uid=slide.slide_uid,
            mask_uri="/path/to/mask.tiff",
            mask_name="Tumor Mask"
        )

        # Add vector annotations
        vector_annotation = await client.slides.add_vector_annotation(
            slide_uid=slide.slide_uid,
            file_uri="/path/to/annotations.geojson",
            name="ROI Annotations",
            format="geojson"
        )

asyncio.run(main())
```

### Pagination Examples

The SDK provides three ways to handle pagination:

```python
# 1. Manual pagination
page = await client.studies.list(page=1, limit=20)
print(f"Page 1: {len(page.items)} items")

# 2. Get all items automatically
all_studies = await client.studies.list_all(page_size=50)
print(f"Retrieved {len(all_studies)} studies")

# 3. Async iteration through pages
async for page in client.studies.list_paginated(page_size=25):
    for study in page.items:
        print(f"Study: {study.name}")
```

### Search and Filtering

```python
# Search studies
results = await client.studies.list(
    search="cancer",
    sort="name",
    sort_dir="asc",
    is_published=True
)

# Search cases with filters
cases = await client.cases.list(
    search="patient",
    name="specific_case",
    has_annotations=True
)
```

### Bulk Operations

The SDK provides convenience methods for common bulk operations:

```python
# Create study with multiple cases
study_uid, case_uids = await client.create_study_with_cases(
    study_name="Bulk Study",
    study_description="Created with bulk operations",
    cases_data=[
        {"name": "Case 1", "metadata": '{"type": "control"}'},
        {"name": "Case 2", "metadata": '{"type": "treatment"}'},
    ]
)

# Create case with multiple slides
case_uid, slide_uids = await client.create_case_with_slides(
    case_name="Multi-slide Case",
    slides_data=[
        {"slide_uri": "/path/slide1.svs", "slide_name": "Slide 1"},
        {"slide_uri": "/path/slide2.svs", "slide_name": "Slide 2"},
    ],
    study_uid=study_uid
)

# Bulk add annotations
annotation_uids = await client.bulk_add_raster_annotations([
    {
        "slide_uid": slide_uids[0],
        "mask_uri": "/path/mask1.tiff",
        "mask_name": "Mask 1",
        "labels": {"region": "tumor"}
    },
    # ... more annotations
])

# Bulk add vector annotations
vector_annotations = []
for i, slide_uid in enumerate(slide_uids):
    vector_annotation = await client.slides.add_vector_annotation(
        slide_uid=slide_uid,
        file_uri=f"/path/annotations_{i}.geojson",
        name=f"Vector Annotation {i}",
        format="geojson",
        labels={"type": "roi"}
    )
    vector_annotations.append(vector_annotation)
```

## Resource Managers

The SDK organizes API access through resource managers:

### Studies (`client.studies`)

- `list()`, `list_all()`, `list_paginated()` - List studies with pagination
- `get(study_uid)` - Get a specific study
- `create()` - Create a new study
- `update()` - Update study details
- `get_metadata()` - Get study statistics
- `get_cases()` - Get cases in a study
- `add_case()`, `remove_case()` - Manage study cases
- `soft_delete()`, `restore()` - Soft deletion operations

### Cases (`client.cases`)

- `list()`, `list_all()`, `list_paginated()` - List cases with pagination
- `get(case_uid)` - Get a specific case
- `create()` - Create a new case
- `update()` - Update case details
- `get_slides()` - Get slides in a case
- `add_slide()` - Add a slide to a case
- `get_neighbors()` - Get neighboring cases in a study
- `soft_delete()`, `restore()` - Soft deletion operations

### Slides (`client.slides`)

- `list()`, `list_all()`, `list_paginated()` - List slides with pagination
- `get(slide_uid)` - Get a specific slide
- `create()` - Create a new slide
- `get_metadata()` - Get slide metadata and properties
- `get_tile()` - Get slide tiles for visualization
- `get_annotations_overview()` - Get annotation summary
- `get_raster_annotations()`, `add_raster_annotation()` - Manage raster annotations
- `get_vector_annotations()`, `add_vector_annotation()` - Manage vector annotations
- `get_raster_annotation_tile()` - Get annotation tiles

### Users (`client.users`)

- `list()`, `list_all()`, `list_paginated()` - List users with pagination
- `get(user_uid)` - Get a specific user
- `create()` - Create a new user
- `update()` - Update user details
- `get_current()` - Get current authenticated user
- `change_password()` - Change user password

### Tenants (`client.tenants`)

- `list()`, `list_all()`, `list_paginated()` - List tenants with pagination
- `get(tenant_uid)` - Get a specific tenant
- `create()` - Create a new tenant
- `update()` - Update tenant details
- `get_domains()`, `add_domain()`, `remove_domain()` - Manage tenant domains

## CLI Tools

The SDK includes a modern CLI tool with rich formatting:

### Installation

```bash
# Make CLI executable
chmod +x slideinsight_cli/main.py

# Or run with python
python -m slideinsight_cli.main
```

### Usage

```bash
# Test login
python -m slideinsight_cli.main login

# List studies
python -m slideinsight_cli.main list-studies --limit 5

# Bulk import from CSV
python -m slideinsight_cli.main bulk-import \
    --url http://localhost:3000 \
    --email myuser@slideinsight.net \
    --study-id STUDY123 \
    --csv data.csv

# Dry run (preview without changes)
python -m slideinsight_cli.main bulk-import --csv data.csv --dry-run
```

### CSV Format for Bulk Import

The bulk import tool supports both raster and vector annotations based on file extension:

```csv
casename,slide_uri,annotation_uri
TCGA-Sample-01,/data/slides/slide001.svs,/data/masks/mask001.tiff
TCGA-Sample-02,/data/slides/slide002.svs,/data/annotations/annotation002.geojson
TCGA-Sample-03,/data/slides/slide003.svs,/data/annotations/annotation003.json
TCGA-Sample-04,/data/slides/slide004.svs,
```

Columns:

- `casename`: Human-readable name for the case/slide
- `slide_uri`: Path or URI to the slide file
- `annotation_uri`: (Optional) Path or URI to the annotation file

**Annotation Types (auto-detected by file extension):**

- **Raster Annotations (Masks)**: `.tiff`, `.png` files
  - Imported via `/api/v1/slides/{slide_uid}/annotations/raster` endpoint
  - Used for pixel-level segmentation masks
- **Vector Annotations**: `.json`, `.geojson` files
  - Imported via `/api/v1/slides/{slide_uid}/annotations/vector` endpoint
  - Used for geometric annotations (polygons, points, etc.)

The CLI tool automatically detects the annotation type based on the file extension and routes it to the appropriate API endpoint.

## Error Handling

The SDK provides a comprehensive exception hierarchy:

```python
from slideinsight_sdk import (
    SlideInsightError,        # Base exception
    AuthenticationError,    # 401 errors
    AuthorizationError,     # 403 errors
    NotFoundError,          # 404 errors
    ValidationError,        # 400 errors
    ConflictError,          # 409 errors
    RateLimitError,         # 429 errors
    ServerError,            # 5xx errors
    NetworkError,           # Network issues
)

try:
    await client.studies.get("nonexistent")
except NotFoundError:
    print("Study not found")
except AuthenticationError:
    print("Please login first")
except SlideInsightError as e:
    print(f"API error: {e.message}")
```

## Authentication

The SDK handles JWT authentication automatically:

```python
# Login stores tokens automatically
await client.login("email", "password")

# Check authentication status
if client.is_authenticated():
    print("Authenticated!")

# Tokens are refreshed automatically when needed
# Manual refresh if needed:
await client.refresh_tokens()

# Logout clears tokens
await client.logout()
```

## Configuration

```python
# Initialize with custom settings
client = SlideInsightClient(
    base_url="https://slideinsight.example.com",
    timeout=60,           # Request timeout in seconds
    max_retries=5,        # Max retry attempts
    retry_backoff=2.0,    # Backoff factor for retries
    auth_cookie_name="_auth",  # Authentication cookie name (default: "_auth")
    refresh_cookie_name="_auth_refresh",  # Refresh cookie name (default: "{auth_cookie_name}_refresh")
    debug=True,           # Enable debug logging (default: False)
)
```

## Debug Logging

The SDK provides comprehensive debug logging to help troubleshoot API requests and responses:

### SDK Debug Mode

```python
# Enable debug logging in the SDK
async with SlideInsightClient("http://localhost:3000", debug=True) as client:
    await client.login("email", "password")
    # All HTTP requests will be logged with full details
    studies = await client.studies.list()
```

### CLI Debug Mode

```bash
# Enable debug logging in CLI commands
python -m slideinsight_cli.main --debug login

# Debug logging with bulk import
python -m slideinsight_cli.main --debug bulk-import \
    --url http://localhost:3000 \
    --study-id STUDY123 \
    --csv data.csv

# Debug logging with list commands
python -m slideinsight_cli.main --debug list-studies --page-size 5
```

### Debug Output Features

When debug mode is enabled, you'll see detailed logging for every HTTP request:

- **Request Method and URL**: `GET /api/v1/studies`
- **Query Parameters**: `{"page": 1, "limit": 10, "q": "search"}`
- **Request Headers**: Including authentication headers (tokens redacted)
- **Cookies**: Authentication cookies (sensitive data redacted)
- **Request Body**: JSON payload (passwords and secrets redacted)

**Security Note**: Sensitive data like passwords, tokens, and API keys are automatically redacted in debug logs to prevent accidental exposure.

### Example Debug Output

```
============================================================
🔄 HTTP REQUEST: GET http://localhost:3000/api/v1/studies
============================================================
📝 Query Parameters: {'page': 1, 'limit': 10, 'q': 'cancer'}
📋 Headers: {'Authorization': 'Bear...4567', 'Accept': 'application/json'}
🍪 Cookies: {'_auth': 'eyJ0...***REDACTED***', '_auth_refresh': 'eyJ0...***REDACTED***'}
============================================================
```

### Cookie Configuration

The SDK uses configurable cookie names that should match your SlideInsight server configuration. The defaults match the standard SlideInsight server configuration:

- **Authentication cookie**: `_auth` (default)
- **Refresh token cookie**: `_auth_refresh` (default)

If your server uses different cookie names (e.g., from `config.example.yaml`), you can configure them:

```python
# For servers using "slideinsight_token" cookie name
client = SlideInsightClient(
    base_url="https://slideinsight.example.com",
    auth_cookie_name="slideinsight_token",
    # refresh_cookie_name will automatically be "slideinsight_token_refresh"
)

# Or specify both explicitly
client = SlideInsightClient(
    base_url="https://slideinsight.example.com",
    auth_cookie_name="custom_auth",
    refresh_cookie_name="custom_refresh",
)
```

**Note**: The cookie names must match your SlideInsight server's `auth.cookie.name` configuration. Check your server's config file:

```yaml
# In your server's config.yaml
auth:
  cookie:
    name: "_auth" # This should match auth_cookie_name parameter
```

## Development

### Running Tests

```bash
# Install test dependencies
pip install pytest pytest-asyncio

# Run tests
pytest tests/
```

### Type Checking

```bash
# Install mypy
pip install mypy

# Run type checking
mypy slideinsight_sdk/
```

### Code Formatting

```bash
# Install black
pip install black

# Format code
black slideinsight_sdk/ slideinsight_cli/
```

## License

SlideInsight SDK is Licensed under the Apache 2.0 License.

## Contributing

1. Follow Google Python Style Guide (120 char line length allowed)
2. Use modern Python 3.11+ features (no `List`, `Dict` - use `list`, `dict`)
3. Add type annotations to all functions
4. Use numpy docstring format
5. Write tests for new functionality
6. Use pytest for testing
