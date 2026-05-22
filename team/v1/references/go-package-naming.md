# Go Package Naming Conventions: "pkg" Clarification

> **Consensus**: Official Go guidance does NOT recommend using `pkg` as a package name, but DOES accept `/pkg` as a directory organization convention at the project root. These apply to different layers and are not contradictory.

---

## 1. NOT Recommended: Using `pkg` as a Package Name

### Official Sources

- Effective Go: https://golang.org/doc/effective_go#package-names
- Code Review Comments: https://github.com/golang/go/wiki/CodeReviewComments#package-names

### Core Principle

Package names should be concise, lowercase, single-word, and accurately describe the package's functionality. Avoid generic, meaningless names like `pkg`, `util`, `common`, `base`, `misc`.

### Reasons

- Generic names blur responsibility boundaries
- Easily become "code dumpsters" where anything gets dumped
- Lead to confused inter-package dependencies
- Reduce readability and maintainability

### Examples

| Bad | Good | Why |
|-----|------|-----|
| `package pkg` | `package auth` | Specific, describes functionality |
| `package util` | `package fileutil` | Clarifies what utilities |
| `package common` | `package httpclient` | Describes the concrete purpose |

```go
// BAD: too generic, no clear responsibility
package pkg

// GOOD: clearly expresses this package provides configuration parsing
package config
```

---

## 2. Recommended: Using `/pkg` as a Directory Organization Convention

### Sources

- Go Project Layout (community standard): https://github.com/golang-standards/project-layout
- Go standard library's own directory structure follows this pattern

### Core Principle

At the project root, a `/pkg` directory can hold public library code that is safe for external projects to import. This is a file organization convention, NOT a package naming rule.

### Directory Structure Example

```
/myproject
├── /cmd          # Executable entry points
├── /internal     # Private code, not importable externally
├── /pkg          # Public library code, importable externally
│   ├── /client   # Package name: client
│   └── /config   # Package name: config
└── /api          # API definitions
```

### `/pkg` vs `/internal`

| Directory | Purpose | Externally Importable |
|-----------|---------|----------------------|
| `/pkg` | Public library code, designed for reuse | Yes |
| `/internal` | Private code, internal use only | No (Go compiler enforces) |

### What `/pkg` Actually Means

- Does NOT affect package name: code in `/pkg/client` still uses `package client`, not `package pkg`
- Expresses intent: clearly tells other developers this code is "public and safe to import"
- Optional convention: not mandatory, small projects can skip it

---

## 3. Summary Table

| Scenario | Recommended | Explanation |
|----------|-------------|-------------|
| Creating a package named `pkg` | No | Generic and meaningless, violates naming conventions |
| Creating a `/pkg` directory at project root | Yes | Good directory organization convention |
| Creating a package named `util` | No | Too generic, prone to abuse |
| Creating a package named `common` | No | Same as above |
| Creating a package named `stringutil` | Yes | Name clearly expresses functionality |

---

## 4. Practical Guidelines

### When naming packages

Ask: "What single-responsibility capability does this package provide?" The name should directly answer this question.

### When organizing project directories

Follow the `/pkg` + `/internal` pattern to make public/private intent clear at a glance.

### Remember the principle

- Directory structure is an **organization problem** -> `/pkg` is allowed
- Package name is a **semantic expression problem** -> Do NOT name it `pkg`

---

## 5. References

- [Effective Go - Package names](https://golang.org/doc/effective_go#package-names)
- [Go Code Review Comments - Package names](https://github.com/golang/go/wiki/CodeReviewComments#package-names)
- [Go Project Layout (community standard)](https://github.com/golang-standards/project-layout)
- [Package names in Go - Go Blog](https://go.dev/blog/package-names)
