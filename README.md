# GoTOON

A Go library for **TOON (Token-Oriented Object Notation)** - a data serialization format designed to reduce token usage by 30-60% compared to JSON while maintaining human readability.

Perfect for LLM applications where token efficiency directly impacts API costs.

## Features

- **Token Efficient**: 30-60% fewer tokens than JSON
- **Human Readable**: Clean, intuitive syntax
- **Three Array Formats**: Inline, Tabular, and List formats
- **Fast**: Optimized for performance
- **Familiar**: Uses struct tags like JSON
- **Round-trip Safe**: Perfect Marshal/Unmarshal compatibility

## Installation

```bash
go get github.com/l00pss/gotoon
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/l00pss/gotoon"
)

type User struct {
    ID   int    `toon:"id"`
    Name string `toon:"name"`
    Age  int    `toon:"age"`
}

func main() {
    user := User{ID: 1, Name: "Alice", Age: 30}
    
    data, _ := toon.Marshal(user)
    fmt.Println(string(data))
    // Output:
    // id: 1
    // name: Alice
    // age: 30
    
    var decoded User
    toon.Unmarshal(data, &decoded)
    fmt.Printf("%+v\n", decoded)
}
```

## TOON Format Examples

### 1. Nested Objects

```
context:
  task: Our favorite hikes together
  location: Boulder
  season: spring_2025
```

vs JSON:
```json
{
  "context": {
    "task": "Our favorite hikes together",
    "location": "Boulder", 
    "season": "spring_2025"
  }
}
```

### 2. Inline Arrays (Primitives)

```
friends[3]: ana,luis,sam
numbers[5]: 1,2,3,4,5
```

vs JSON:
```json
{
  "friends": ["ana", "luis", "sam"],
  "numbers": [1, 2, 3, 4, 5]
}
```

### 3. Tabular Arrays (CSV-style for structs)

```
hikes[3]{id,name,distanceKm,elevationGain,companion,wasSunny}:
  1,Blue Lake Trail,7.5,320,ana,true
  2,Ridge Overlook,9.2,540,luis,false  
  3,Wildflower Loop,5.1,180,sam,true
```

vs JSON:
```json
{
  "hikes": [
    {
      "id": 1,
      "name": "Blue Lake Trail",
      "distanceKm": 7.5,
      "elevationGain": 320,
      "companion": "ana",
      "wasSunny": true
    },
    {
      "id": 2,
      "name": "Ridge Overlook", 
      "distanceKm": 9.2,
      "elevationGain": 540,
      "companion": "luis",
      "wasSunny": false
    },
    {
      "id": 3,
      "name": "Wildflower Loop",
      "distanceKm": 5.1, 
      "elevationGain": 180,
      "companion": "sam",
      "wasSunny": true
    }
  ]
}
```

### 4. List Arrays (YAML-style for varied structures)

```
items[2]:
  - id: 1
    name: Item One
  - id: 2
    name: Item Two
```

## Advanced Usage

### Custom Options

```go
opts := toon.MarshalOptions{
    Indent:     4,
    Delimiter:  toon.DelimiterTab,
    UseTabular: true,
}

data, err := toon.MarshalWithOptions(obj, opts)
```

### Delimiter Options

| Delimiter | Character | Token Efficiency | Readability | Use Case |
|-----------|-----------|------------------|-------------|----------|
| **Comma** | `,` | Good | Excellent | Default, most readable |
| **Tab** | `\t` | **Best** | Good | Maximum token savings |
| **Pipe** | `\|` | Good | Good | Data contains commas |

### Struct Tags

```go
type User struct {
    ID       int    `toon:"id"`
    FullName string `toon:"name"`
    Password string `toon:"-"`
}
```

## Performance

```bash
go test -bench=.
```

```
BenchmarkMarshal-10               349092              3367 ns/op
BenchmarkUnmarshal-10              84189             13794 ns/op
```

- **Marshal**: 3.4 microseconds per operation
- **Unmarshal**: 13.8 microseconds per operation
- **Throughput**: ~349K marshal ops/sec, ~84K unmarshal ops/sec

## Token Efficiency

Real-world example (hiking data):

| Format | Size | Tokens* | Savings |
|--------|------|---------|----------|
| JSON | 680 chars | ~170 | - |
| **TOON** | **287 chars** | **~71** | **57.8%** |

*Estimated at 4 chars per token

## API Reference

### Core Functions

```go
// Marshal with default options
func Marshal(v any) ([]byte, error)

// Marshal with custom options  
func MarshalWithOptions(v any, opts MarshalOptions) ([]byte, error)

// Unmarshal TOON data
func Unmarshal(data []byte, v any) error

// Validate TOON syntax
func Valid(data []byte) bool
```

### Types

```go
type MarshalOptions struct {
    Indent     int       // Indentation spaces (default: 2)
    Delimiter  Delimiter // Array delimiter (default: comma) 
    UseTabular bool      // Use tabular format for structs (default: true)
}

type Delimiter string
const (
    DelimiterComma Delimiter = ","   // Most readable
    DelimiterTab   Delimiter = "\t"  // Most efficient  
    DelimiterPipe  Delimiter = "|"   // Safe for commas
)
```

## Use Cases

- LLM applications (reduce token costs)
- Configuration files
- Data exchange APIs
- Log structured data

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details.
