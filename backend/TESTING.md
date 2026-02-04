# Backend Testing Documentation

## Test Coverage Summary

All tests passing ✅

| Package | Coverage | Test Files |
|---------|----------|------------|
| handlers | 70.5% | identify_test.go |
| services | 90.7% | care_data_test.go, ml_client_test.go |
| utils | 75.0% | config_test.go, file_test.go, plant_test.go |

**Overall:** Strong test coverage across all critical components

## Test Files Overview

### 1. handlers/identify_test.go
Tests the main HTTP handler for plant identification.

**Test Cases:**
- ✅ `TestIdentifyHandlerHandle` - End-to-end handler tests
  - Successful identification with high confidence
  - Low confidence returns genus only
  - Method not allowed (non-POST requests)
  - No image file provided
- ✅ `TestProcessMLResponse` - Business logic tests
  - High confidence shows species
  - Low confidence shows genus only
  - Threshold boundary testing (exactly at 0.4)

**Key Features:**
- Uses mock interfaces for dependencies
- Tests multipart form handling
- Validates JSON responses
- Tests confidence threshold logic (0.4 cutoff)

### 2. services/care_data_test.go
Tests care instructions loading and retrieval.

**Test Cases:**
- ✅ `TestNewCareDataService` - Service initialization
  - Valid care data file loading
  - Non-existent file error handling
  - Empty path error handling
- ✅ `TestGetCareInstructions` - Care data retrieval
  - Species-level care retrieval
  - Fallback to genus-level care
  - Error handling for missing data
  - Empty species handling

**Key Features:**
- Uses test fixtures (testdata/care_data_test.json)
- Tests fallback logic (species → genus)
- Validates error scenarios

### 3. services/ml_client_test.go
Tests ML service HTTP client.

**Test Cases:**
- ✅ `TestNewMLClient` - Client initialization
- ✅ `TestInfer` - Inference requests
  - Successful inference
  - ML service error responses
  - Empty predictions handling
- ✅ `TestHealthCheck` - Health check endpoint
  - Healthy service
  - Unhealthy service
  - Service unavailable
- ✅ `TestInferServerDown` - Connection failure handling

**Key Features:**
- Mock HTTP server (httptest)
- Tests request/response JSON marshaling
- Validates error propagation
- Tests timeout and connection failures

### 4. utils/config_test.go
Tests configuration loading from environment variables.

**Test Cases:**
- ✅ `TestLoadConfig` - Configuration loading
  - Default configuration values
  - Custom configuration from env vars
  - Partial custom configuration

**Key Features:**
- Tests environment variable parsing
- Validates default values
- Tests type conversions (string to int64, float64)

### 5. utils/file_test.go
Tests file upload, validation, and storage.

**Test Cases:**
- ✅ `TestNewFileUploader` - Uploader initialization
  - Valid configuration
  - Nested directory creation
- ✅ `TestValidateFile` - File validation
  - Valid JPG/PNG/JPEG files
  - File size validation (max 5MB)
  - Invalid extension rejection
  - Empty file rejection
  - Case insensitive extensions
- ✅ `TestSaveFile` - File saving
  - Save valid files with UUID names
  - Verify file content
- ✅ `TestDeleteFile` - File deletion
  - Delete existing files
  - Handle non-existent files

**Key Features:**
- Mock multipart.File implementation
- UUID filename generation testing
- File system operations
- Validation logic testing

### 6. utils/plant_test.go
Tests plant label parsing and formatting.

**Test Cases:**
- ✅ `TestParseLabel` - Label parsing
  - Valid genus_species labels
  - Multiple underscores handling
  - Genus only labels
  - Empty labels
- ✅ `TestFormatGenus` - Genus formatting
  - Lowercase to capitalized
  - Already capitalized handling
  - Single letter handling
- ✅ `TestFormatSpecies` - Species formatting
  - Valid species formatting
  - Multiple word species
  - Genus only (no species)

**Key Features:**
- String manipulation testing
- Edge case handling
- Format validation

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests with Verbose Output
```bash
go test ./... -v
```

### Run Tests with Coverage
```bash
go test ./... -cover
```

### Run Tests for Specific Package
```bash
go test ./handlers -v
go test ./services -v
go test ./utils -v
```

### Generate Coverage Report
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Tests with Race Detection
```bash
go test ./... -race
```

## Test Fixtures

### testdata/care_data_test.json
Mock care data for testing care data service.

```json
{
  "test_genus": {
    "sunlight": "Test genus sunlight",
    "watering": "Test genus watering",
    "soil": "Test genus soil",
    "notes": "Test genus notes"
  },
  "test_genus_species": {
    "sunlight": "Test species sunlight",
    "watering": "Test species watering",
    "soil": "Test species soil",
    "notes": "Test species notes"
  }
}
```

### Temporary Directories
Tests create and clean up these directories automatically:
- `testdata/uploads` - File upload tests
- `testdata/uploads_test` - File save tests
- `testdata/uploads_delete` - File delete tests
- `testdata/uploads_handler_test` - Handler tests

## Mock Implementations

### mockMLClient
Simulates ML service responses for handler testing.

```go
type mockMLClient struct {
    response *models.MLInferenceResponse
    err      error
}
```

### mockCareDataService
Simulates care data retrieval for handler testing.

```go
type mockCareDataService struct {
    care models.CareInstructions
    err  error
}
```

### mockFile
Implements multipart.File interface for file upload testing.

```go
type mockFile struct {
    *bytes.Reader
}
```

## Test Architecture

### Interfaces for Testability
To enable mocking, handlers use interfaces instead of concrete types:

- `MLClientInterface` - ML service client
- `CareDataServiceInterface` - Care data service
- `FileUploaderInterface` - File uploader

This allows tests to inject mocks without depending on real services.

## Test Principles

1. **Isolation**: Each test is independent and doesn't rely on others
2. **Cleanup**: Tests clean up resources (files, directories) after execution
3. **Mocking**: External dependencies are mocked for fast, reliable tests
4. **Coverage**: Tests cover success paths, error paths, and edge cases
5. **Table-Driven**: Most tests use table-driven patterns for multiple scenarios

## What's Tested

✅ **HTTP Handling**
- Multipart form parsing
- Request validation
- Response formatting
- Error responses

✅ **Business Logic**
- Confidence threshold (0.4)
- Species vs genus display logic
- Care data fallback (species → genus)
- Label parsing (genus_species format)

✅ **File Operations**
- Upload validation (type, size)
- File storage with UUID names
- File deletion
- Directory creation

✅ **External Integration**
- ML service HTTP client
- Health check endpoint
- Error handling and retries

✅ **Configuration**
- Environment variable parsing
- Default values
- Type conversions

## What's NOT Tested (Future Improvements)

- Integration tests with real ML service
- Performance/load testing
- Concurrent request handling
- Real file upload from HTTP clients
- Database operations (none in V1)

## Continuous Integration

Tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run tests
  run: go test ./... -v -cover

- name: Check coverage
  run: |
    go test ./... -coverprofile=coverage.out
    go tool cover -func=coverage.out
```

## Troubleshooting Tests

### Tests Fail to Create Directories
**Issue:** Permission errors when creating testdata directories

**Solution:** Ensure the test process has write permissions
```bash
chmod +w testdata/
```

### Mock HTTP Server Timeouts
**Issue:** httptest server tests timeout

**Solution:** Check that server.Close() is called in defer

### Care Data Not Found
**Issue:** Tests can't find testdata/care_data_test.json

**Solution:** Run tests from the backend/ directory or use relative paths

## Best Practices

1. **Always clean up** - Use defer to remove test files/directories
2. **Use table-driven tests** - Makes adding test cases easy
3. **Test error paths** - Don't just test happy paths
4. **Mock external dependencies** - Never call real ML service in tests
5. **Keep tests fast** - Unit tests should run in milliseconds

## Test Maintenance

When adding new features:
1. Write tests first (TDD approach)
2. Ensure coverage stays above 70%
3. Update this documentation
4. Add test fixtures if needed
5. Consider edge cases and error scenarios

---

Last Updated: 2026-02-04
