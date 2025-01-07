# Result Module

This module is responsible for collecting and processing test results in the OpenStress project.

## Overview

The `result` package includes functionalities for:
- Collecting test data
- Processing and storing results
- Logging and reporting

## Key Components

### Collector
- **Collector**: The main struct for collecting results. It manages the collection process and stores results.
- **CollectorConfig**: Configuration struct for initializing the collector, including batch size, output format, and logger.

### ResultData
- **ResultData**: Struct that represents a single test result, including fields like ID, response time, and status code.

## Usage

To use the `result` module, follow these steps:
1. Create a logger that implements the `Logger` interface.
2. Initialize a `CollectorConfig` with desired settings.
3. Create a new `Collector` using `NewCollector(config)`.
4. Call `InitializeCollector()` to prepare for data collection.
5. Use `CollectDataWithParams(...)` to collect results.
6. Finally, call `CloseCollector()` to clean up resources.

## Example

```go
package main

import (
	"openStress/result"
	"time"
)

func main() {
	logger := &result.ConsoleLogger{}
	config := result.CollectorConfig{
		BatchSize:    10,
		OutputFormat: "json",
		JTLFilePath:  "path/to/jtl/file.jtl",
		Logger:       logger,
		NumGoroutines: 2,
		CollectInterval: 5,
	}
	collector, err := result.NewCollector(config)
	collector.InitializeCollector()
	collector.CollectDataWithParams(...)
	collector.CloseCollector()
}
```

## License

This project is licensed under the MIT License.
