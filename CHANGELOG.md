## v0.2.0 (2026-02-17)

### Feat

- **windows**: add symlink permission warning for Windows 10/11
- **storage**: add Windows cloud storage path detection and fix tests
- add Windows support with PowerShell installer

### Fix

- **install.ps1**: prevent terminal closure when running via iex
- **lint**: suppress false positive gosec warnings
- **installer**: prevent PowerShell installer from hanging after successful install
- correct Windows release artifacts and enable ARM64 builds

### Refactor

- **storage**: remove deprecated ExpandHome and ContractHome functions
- **storage**: consolidate home expansion to pathutil package

## v0.1.1 (2026-02-16)

### Fix

- **storage**: remove iCloud from Linux provider paths

## v0.1.0 (2026-02-15)

### Feat

- **ci**: add release automation and CI/CD setup
