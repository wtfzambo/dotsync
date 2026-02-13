# Go Crash Course for Python/TypeScript Developers

A practical guide to reading and writing Go, with examples from the dotsync codebase.

## Table of Contents
1. [Basic Syntax Differences](#1-basic-syntax-differences)
2. [Variables, Types, and Type Inference](#2-variables-types-and-type-inference)
3. [Structs (Go's Classes)](#3-structs-gos-classes)
4. [Functions and Methods](#4-functions-and-methods)
5. [Interfaces](#5-interfaces)
6. [Error Handling](#6-error-handling)
7. [Packages and Imports](#7-packages-and-imports)
8. [Goroutines and Channels](#8-goroutines-and-channels-brief-intro)
9. [Common Patterns in This Codebase](#9-common-patterns-in-this-codebase)

---

## 1. Basic Syntax Differences

### Declaration Order
Go is **type-after-name**, unlike TypeScript.

```typescript
// TypeScript
const name: string = "dotsync";
function greet(name: string): string { ... }
```

```go
// Go
var name string = "dotsync"
func greet(name string) string { ... }
```

### Exported vs Private
Go uses **capitalization** instead of keywords.

```python
# Python
class MyClass:
    def public_method(self):  # public by convention
    def _private_method(self):  # private by convention
```

```go
// Go
type MyStruct struct {
    PublicField  string  // Exported (accessible from other packages)
    privateField string  // Unexported (only in this package)
}

func PublicFunc() {}   // Exported
func privateFunc() {}  // Unexported
```

**Example from codebase** (`internal/manifest/manifest.go:10-28`):
```go
type Manifest struct {
    Version int                `json:"version"`  // Exported
    Entries map[string]Entry   `json:"entries"`  // Exported
}

type Entry struct {
    Root  string   `json:"root"`   // Exported
    Files []string `json:"files"`  // Exported
}
```

### No Classes, No Inheritance
Go has **structs with methods**, not classes. No inheritance, only **composition**.

```typescript
// TypeScript
class Animal {
    name: string;
    constructor(name: string) { this.name = name; }
    speak() { console.log("..."); }
}

class Dog extends Animal {
    bark() { console.log("woof"); }
}
```

```go
// Go - composition instead of inheritance
type Animal struct {
    Name string
}

type Dog struct {
    Animal  // Embedded struct (composition)
}

func (a Animal) Speak() { fmt.Println("...") }
func (d Dog) Bark() { fmt.Println("woof") }

// Usage
dog := Dog{Animal: Animal{Name: "Buddy"}}
dog.Speak()  // Inherited from Animal
dog.Bark()
```

### No Semicolons (Usually)
Like Python, Go auto-inserts semicolons at line endings. **Don't put opening braces on new lines!**

```go
// CORRECT
func main() {
    fmt.Println("hello")
}

// WRONG - won't compile
func main()
{
    fmt.Println("hello")
}
```

---

## 2. Variables, Types, and Type Inference

### Declaration Styles

```go
// Explicit type
var name string = "dotsync"

// Type inference
var name = "dotsync"

// Short declaration (most common, only inside functions)
name := "dotsync"

// Multiple variables
var x, y int = 1, 2
a, b := "hello", "world"
```

### Zero Values
Unlike TypeScript's `undefined` or Python's `None`, Go gives **zero values** to uninitialized variables.

```go
var i int       // 0
var f float64   // 0.0
var b bool      // false
var s string    // "" (empty string)
var p *int      // nil (pointer)
var slice []int // nil (slice)
var m map[string]int // nil (map)
```

**Example from codebase** (`internal/manifest/manifest.go:32-35`):
```go
func New() *Manifest {
    return &Manifest{
        Version: CurrentVersion,
        Entries: make(map[string]Entry), // Must initialize maps!
    }
}
```

**Important**: Maps and slices need initialization with `make()`.

### Basic Types

```go
// Integers
int8, int16, int32, int64
uint8, uint16, uint32, uint64
int, uint  // Platform-dependent size (32 or 64 bit)

// Floats
float32, float64

// Others
bool
string
byte     // alias for uint8
rune     // alias for int32, represents a Unicode code point

// Composite
[]T       // slice of T
[5]T      // array of 5 T's (fixed size)
map[K]V   // map with keys K and values V
*T        // pointer to T
```

### Type Conversion (Explicit Only)
Unlike Python, Go **requires explicit type conversion**.

```python
# Python - implicit
x = 10
y = 3.5
result = x + y  # Works: 13.5
```

```go
// Go - explicit required
x := 10
y := 3.5
result := float64(x) + y  // Must convert x to float64
```

---

## 3. Structs (Go's Classes)

Structs are Go's way of creating custom types. Think of them as Python dataclasses or TypeScript interfaces with data.

### Basic Struct

```typescript
// TypeScript
interface Config {
    root: string;
    autoSync: boolean;
}

const cfg: Config = {
    root: "/path/to/files",
    autoSync: true
};
```

```go
// Go
type Config struct {
    Root     string
    AutoSync bool
}

cfg := Config{
    Root:     "/path/to/files",
    AutoSync: true,
}
// or shorter: cfg := Config{"/path/to/files", true}
```

**Example from codebase** (`internal/config/config.go:5-11`):
```go
type LocalConfig struct {
    Root string `json:"root"`
}

type Config struct {
    Local *LocalConfig `json:"local,omitempty"`
}
```

### Struct Tags
Tags provide metadata (commonly used for JSON serialization).

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email,omitempty"`  // Omit if empty
    Age   int    `json:"-"`                // Never serialize
}
```

### Embedding (Composition)
Go uses **embedding** instead of inheritance.

```go
type Base struct {
    ID   int
    Name string
}

type Extended struct {
    Base           // Embedded struct - fields "promoted"
    Extra string
}

e := Extended{
    Base:  Base{ID: 1, Name: "test"},
    Extra: "more data",
}

fmt.Println(e.ID)    // Access Base fields directly: 1
fmt.Println(e.Name)  // "test"
```

---

## 4. Functions and Methods

### Functions

```python
# Python
def add(x: int, y: int) -> int:
    return x + y

def multiple_returns() -> tuple[int, str]:
    return 42, "hello"
```

```go
// Go
func add(x int, y int) int {
    return x + y
}

// Multiple return values (common in Go!)
func multipleReturns() (int, string) {
    return 42, "hello"
}

// Named return values
func divide(a, b float64) (result float64, err error) {
    if b == 0 {
        err = fmt.Errorf("division by zero")
        return  // Returns zero value for result, err
    }
    result = a / b
    return  // Returns result and nil for err
}
```

### Methods
Methods are functions with a **receiver**.

```typescript
// TypeScript
class Manifest {
    entries: Map<string, Entry>;
    
    addFile(name: string, root: string, path: string): boolean {
        // ...
    }
}
```

```go
// Go
type Manifest struct {
    Entries map[string]Entry
}

// Method with pointer receiver (can modify the struct)
func (m *Manifest) AddFile(name, root, relPath string) bool {
    // m is mutable
    entry := m.Entries[name]
    entry.Files = append(entry.Files, relPath)
    m.Entries[name] = entry
    return true
}

// Method with value receiver (read-only)
func (m Manifest) GetEntryCount() int {
    // m is a copy, changes don't affect original
    return len(m.Entries)
}
```

**Example from codebase** (`internal/manifest/manifest.go:43-74`):
```go
func (m *Manifest) AddFile(name, root, relPath string) bool {
    entry, exists := m.Entries[name]
    if !exists {
        m.Entries[name] = Entry{
            Root:  root,
            Files: []string{relPath},
        }
        return true
    }

    for _, f := range entry.Files {
        if f == relPath {
            return false
        }
    }

    entry.Files = append(entry.Files, relPath)
    m.Entries[name] = entry
    return true
}
```

**Rule of thumb**: Use pointer receivers (`*T`) when:
- Method needs to modify the receiver
- Struct is large (avoid copying)
- Consistency (if some methods use `*T`, all should)

### Variadic Functions

```go
func sum(numbers ...int) int {
    total := 0
    for _, n := range numbers {
        total += n
    }
    return total
}

sum(1, 2, 3, 4)  // 10
```

---

## 5. Interfaces

Go interfaces are **implicit**. If a type has the right methods, it implements the interface automatically.

```typescript
// TypeScript - explicit
interface Speaker {
    speak(): string;
}

class Dog implements Speaker {
    speak(): string { return "woof"; }
}
```

```go
// Go - implicit
type Speaker interface {
    Speak() string
}

type Dog struct {
    Name string
}

// Dog implements Speaker automatically by having Speak() method
func (d Dog) Speak() string {
    return "woof"
}

func makeItSpeak(s Speaker) {
    fmt.Println(s.Speak())
}

dog := Dog{Name: "Buddy"}
makeItSpeak(dog)  // Works!
```

### The Empty Interface
`interface{}` (or `any` in Go 1.18+) accepts any type.

```go
func printAnything(v interface{}) {
    fmt.Println(v)
}

printAnything(42)
printAnything("hello")
printAnything([]int{1, 2, 3})
```

### Type Assertions
Convert interface back to concrete type.

```go
var i interface{} = "hello"

s := i.(string)        // Type assertion
fmt.Println(s)         // "hello"

// Safe type assertion
s, ok := i.(string)
if ok {
    fmt.Println("It's a string:", s)
}
```

**Example from codebase** (`cmd/dotsync/cmd/add.go:68-83`):
```go
if err := pathutil.ValidateForAdd(absPath); err != nil {
    if valErr, ok := err.(pathutil.ValidationError); ok {
        if valErr.IsWarn {
            fmt.Printf("Warning: %s\n", valErr.Message)
            // Continue...
        } else {
            return valErr
        }
    }
    return fmt.Errorf("validation failed: %w", err)
}
```

### Common Standard Interfaces

```go
// io.Reader - anything that can be read from
type Reader interface {
    Read(p []byte) (n int, err error)
}

// io.Writer - anything that can be written to
type Writer interface {
    Write(p []byte) (n int, err error)
}

// error - anything that can be an error
type error interface {
    Error() string
}

// fmt.Stringer - anything that can be a string
type Stringer interface {
    String() string
}
```

**Example from codebase** (`internal/symlink/symlink.go:74-83`):
```go
type Status int

// Implements fmt.Stringer interface
func (s Status) String() string {
    switch s {
    case StatusNotExist:
        return "not_exist"
    case StatusLinked:
        return "linked"
    // ...
    }
}
```

---

## 6. Error Handling

**No exceptions!** Go uses explicit error returns.

### Basic Pattern

```python
# Python
try:
    result = risky_operation()
except Exception as e:
    print(f"Error: {e}")
```

```go
// Go - errors are values
result, err := riskyOperation()
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return err
}
// Use result...
```

### Creating Errors

```go
import "errors"

// Simple error
err := errors.New("something went wrong")

// Formatted error
err := fmt.Errorf("failed to process %s: %s", filename, reason)
```

### Error Wrapping
Add context while preserving the original error.

```go
// Wrap with %w
if err := os.Open(file); err != nil {
    return fmt.Errorf("opening config file: %w", err)
}

// Later, unwrap to check the original error
if errors.Is(err, os.ErrNotExist) {
    // Handle file not found
}
```

**Example from codebase** (`internal/symlink/symlink.go:16-18`):
```go
if err := os.MkdirAll(parentDir, 0755); err != nil {
    return fmt.Errorf("creating parent directory: %w", err)
}
```

### Custom Errors
Create error types with extra data.

**Example from codebase** (`internal/pathutil/validate.go:13-22`):
```go
type ValidationError struct {
    Path    string
    Message string
    IsWarn  bool
}

func (e ValidationError) Error() string {
    return e.Message
}

// Usage
return ValidationError{
    Path:    path,
    Message: "path is inside a cloud storage directory",
    IsWarn:  true,
}
```

### Idiomatic Error Handling Patterns

```go
// Early return on error
func doSomething() error {
    if err := step1(); err != nil {
        return err
    }
    
    if err := step2(); err != nil {
        return err
    }
    
    return nil
}

// Defer for cleanup (always runs)
func readFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()  // Runs when function exits
    
    // Use f...
    return nil
}
```

**Example from codebase** (`internal/backup/backup.go:97-114`):
```go
func copyFile(src, dst string) error {
    sourceFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer sourceFile.Close()  // Cleanup

    destFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destFile.Close()  // Cleanup

    _, err = io.Copy(destFile, sourceFile)
    return err
}
```

### Error Checking Helpers

```go
// Check specific error types
if os.IsNotExist(err) {
    // File doesn't exist
}

if errors.Is(err, io.EOF) {
    // End of file
}
```

---

## 7. Packages and Imports

### Package Declaration
Every Go file starts with a package declaration.

```go
// Package comment (shows in docs)
// Package manifest handles reading and writing the dotsync manifest file.
package manifest

import (
    "fmt"
    "os"
)
```

### Import Styles

```go
// Single import
import "fmt"

// Multiple imports
import (
    "fmt"
    "os"
    "strings"
)

// Aliased import
import (
    "fmt"
    oldpkg "github.com/old/package"
)

// Blank import (runs init() only)
import _ "github.com/lib/pq"
```

### Internal Packages
The `internal/` directory creates package boundaries.

```
dotsync/
├── cmd/dotsync/              # Main application
├── internal/
│   ├── manifest/             # Only accessible within dotsync
│   ├── config/
│   └── symlink/
```

**Example from codebase**:
```go
// cmd/dotsync/cmd/add.go
import (
    "github.com/wtfzambo/dotsync/internal/config"     // OK
    "github.com/wtfzambo/dotsync/internal/manifest"   // OK
    "github.com/wtfzambo/dotsync/internal/pathutil"   // OK
)
```

External packages **cannot** import from `internal/`.

### Package Organization Patterns

**By feature** (this codebase):
```
internal/
├── backup/       # Backup operations
├── config/       # Configuration management
├── manifest/     # Manifest handling
├── pathutil/     # Path utilities
├── storage/      # Cloud storage detection
└── symlink/      # Symlink operations
```

### Init Functions
Run automatically when package is imported.

```go
package database

func init() {
    // Runs once when package is imported
    registerDrivers()
}
```

---

## 8. Goroutines and Channels (Brief Intro)

Go's concurrency primitives. **Note**: This codebase doesn't use them (CLI tool with sequential operations).

### Goroutines
Lightweight threads.

```python
# Python
import threading

def task():
    print("running")

thread = threading.Thread(target=task)
thread.start()
```

```go
// Go
func task() {
    fmt.Println("running")
}

go task()  // Runs in a new goroutine
```

### Channels
Communicate between goroutines.

```go
// Create a channel
ch := make(chan int)

// Send to channel (blocks until received)
ch <- 42

// Receive from channel (blocks until sent)
value := <-ch

// Buffered channel (doesn't block until full)
ch := make(chan int, 10)
```

### Example

```go
func worker(id int, jobs <-chan int, results chan<- int) {
    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
        results <- job * 2
    }
}

func main() {
    jobs := make(chan int, 5)
    results := make(chan int, 5)

    // Start 3 workers
    for i := 1; i <= 3; i++ {
        go worker(i, jobs, results)
    }

    // Send jobs
    for j := 1; j <= 5; j++ {
        jobs <- j
    }
    close(jobs)

    // Collect results
    for r := 1; r <= 5; r++ {
        <-results
    }
}
```

### When NOT to use goroutines
- Sequential operations (file I/O)
- When order matters
- CLI tools with user prompts
- Simple scripts

---

## 9. Common Patterns in This Codebase

### 1. Constructor Functions

```go
// Pattern: New() returns a pointer to initialized struct
func New() *Manifest {
    return &Manifest{
        Version: CurrentVersion,
        Entries: make(map[string]Entry),  // Initialize map!
    }
}

// Usage
m := manifest.New()
```

**From**: `internal/manifest/manifest.go:30-35`

### 2. Error Wrapping Chain

```go
// Pattern: Add context at each layer
func AddFile(path string) error {
    if err := validatePath(path); err != nil {
        return fmt.Errorf("validating path: %w", err)
    }
    
    if err := copyFile(path); err != nil {
        return fmt.Errorf("copying file: %w", err)
    }
    
    return nil
}

// Creates error chain:
// "adding file: copying file: creating directory: permission denied"
```

### 3. Table-Driven Tests

```go
func TestInferName(t *testing.T) {
    tests := []struct {
        name     string
        path     string
        want     string
    }{
        {"config file", "~/.config/opencode/config.json", "opencode"},
        {"dotfile", "~/.zshrc", "zsh"},
        {"app support", "~/Library/Application Support/Code/settings.json", "vscode"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := InferName(tt.path)
            if got != tt.want {
                t.Errorf("InferName(%q) = %q, want %q", tt.path, got, tt.want)
            }
        })
    }
}
```

**From**: `internal/manifest/manifest_test.go:109-136`

### 4. Enum-like Types with String Conversion

```go
type Status int

const (
    StatusNotExist  Status = iota  // 0
    StatusLinked                   // 1
    StatusBroken                   // 2
    StatusIncorrect                // 3
)

// Implement fmt.Stringer for nice printing
func (s Status) String() string {
    switch s {
    case StatusNotExist:
        return "not_exist"
    case StatusLinked:
        return "linked"
    case StatusBroken:
        return "broken"
    case StatusIncorrect:
        return "incorrect"
    default:
        return "unknown"
    }
}
```

**From**: `internal/symlink/symlink.go:63-83`

### 5. Defer for Resource Cleanup

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()  // Guaranteed to run when function exits
    
    // Multiple defers stack (LIFO order)
    lock.Lock()
    defer lock.Unlock()
    
    // Do work...
    return nil
}
```

### 6. Comma-Ok Idiom

```go
// Check if map key exists
entry, exists := m.Entries[name]
if !exists {
    // Handle missing key
}

// Type assertion
if valErr, ok := err.(ValidationError); ok {
    // It's a ValidationError
}

// Channel receive
value, ok := <-ch
if !ok {
    // Channel closed
}
```

**From**: `cmd/dotsync/cmd/add.go:68-83`

### 7. Early Returns for Guard Clauses

```go
func ValidatePath(path string) error {
    if path == "" {
        return errors.New("path cannot be empty")
    }
    
    if !filepath.IsAbs(path) {
        return errors.New("path must be absolute")
    }
    
    if _, err := os.Stat(path); err != nil {
        return fmt.Errorf("path does not exist: %w", err)
    }
    
    return nil  // All checks passed
}
```

### 8. Separation of I/O and Logic

```go
// manifest.go - business logic
type Manifest struct { ... }
func (m *Manifest) AddFile(...) { ... }

// io.go - file operations
func Load(path string) (*Manifest, error) { ... }
func (m *Manifest) Save(path string) error { ... }
```

**Structure**:
- `internal/manifest/manifest.go` - domain logic
- `internal/manifest/io.go` - persistence

### 9. Rollback Pattern

```go
// Pattern: Backup -> Operation -> Verify -> Cleanup/Restore
func moveWithRollback(src, dst string) error {
    // Create backup
    bk, err := backup.Create(src)
    if err != nil {
        return err
    }
    
    // Try operation
    if err := os.Rename(src, dst); err != nil {
        bk.Restore()  // Rollback on failure
        return err
    }
    
    bk.Cleanup()  // Success, remove backup
    return nil
}
```

**From**: `cmd/dotsync/cmd/add.go:185-214`

### 10. Platform-Specific Code

```go
import "runtime"

func getConfigDir() string {
    switch runtime.GOOS {
    case "darwin":
        return "/Library/Application Support"
    case "linux":
        return "~/.config"
    case "windows":
        return "%APPDATA%"
    default:
        return "~"
    }
}
```

**From**: `internal/storage/providers.go:57-69`

---

## Quick Reference Card

### Common Types
```go
string, int, float64, bool
[]T          // slice
map[K]V      // map
*T           // pointer
interface{}  // any type
```

### Variable Declaration
```go
var x int = 42      // Explicit
var x = 42          // Inferred
x := 42             // Short (only in functions)
```

### Conditionals
```go
if x > 0 {
    // ...
} else if x < 0 {
    // ...
} else {
    // ...
}

// With initialization
if err := doSomething(); err != nil {
    return err
}
```

### Loops
```go
// For loop
for i := 0; i < 10; i++ {
    // ...
}

// While-style
for condition {
    // ...
}

// Infinite loop
for {
    // ...
}

// Range over slice/array
for i, v := range slice {
    // i is index, v is value
}

// Range over map
for k, v := range myMap {
    // k is key, v is value
}

// Ignore index/key
for _, v := range slice {
    // ...
}
```

### Functions
```go
func name(param Type) ReturnType {
    return value
}

// Multiple returns
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
```

### Structs and Methods
```go
type Person struct {
    Name string
    Age  int
}

func (p *Person) Birthday() {
    p.Age++
}
```

### Interfaces
```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

// Implement by having the method
type MyReader struct{}
func (r MyReader) Read(p []byte) (int, error) {
    // ...
}
```

### Error Handling
```go
result, err := riskyOperation()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
// Use result...
```

---

## Practice Reading This Codebase

Start with these files to practice:

1. **Simple structs**: `internal/config/config.go:5-11`
2. **Methods**: `internal/manifest/manifest.go:43-74`
3. **Error handling**: `internal/symlink/symlink.go:16-18`
4. **Tests**: `internal/manifest/manifest_test.go:109-136`
5. **CLI commands**: `cmd/dotsync/cmd/add.go`

Look for these patterns:
- `:=` for variable declaration
- `if err != nil` for error checking
- `defer` for cleanup
- `*Type` for pointer receivers
- Capitalized names for exported items

---

## Additional Resources

- [Official Tour of Go](https://tour.golang.org/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [Go Proverbs](https://go-proverbs.github.io/)

---

**Next Steps**: Start reading files in `internal/` to see these patterns in action. The code is well-structured and uses idiomatic Go throughout.
