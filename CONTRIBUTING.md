# Contributing to simply_syslog

Thank you for your interest in contributing to simply_syslog! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.25.1 or later
- Docker (optional, for container testing)
- Git

### Development Setup

1. **Fork and clone the repository**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/simply_syslog.git
   cd simply_syslog
   ```

2. **Install Task (optional but recommended)**:
   ```bash
   # macOS
   brew install go-task
   
   # Linux (using sh)
   sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
   
   # Or use go install
   go install github.com/go-task/task/v3/cmd/task@latest
   ```

3. **Download dependencies**:
   ```bash
   go mod download
   ```

4. **Build the application**:
   ```bash
   task build
   # Or
   go build -o build/simply-syslog ./cmd/simplysyslog/main.go
   ```

5. **Run the application**:
   ```bash
   task run
   # Or
   go run ./cmd/simplysyslog/main.go
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the coding standards below

3. **Run tests** (when they exist):
   ```bash
   go test ./...
   ```

4. **Format your code**:
   ```bash
   go fmt ./...
   ```

5. **Run static analysis**:
   ```bash
   go vet ./...
   ```

6. **Commit your changes**:
   ```bash
   git add .
   git commit -m "Brief description of changes"
   ```

7. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

8. **Open a Pull Request** on GitHub

### Coding Standards

- Follow standard Go conventions and idioms
- Use `gofmt` for code formatting
- Write clear, descriptive commit messages
- Add comments for exported functions and types
- Keep functions small and focused
- Handle errors explicitly
- Use meaningful variable and function names

### Testing

- Write unit tests for new functionality
- Ensure all tests pass before submitting a PR
- Aim for high test coverage (>70% target)
- Test both success and failure cases
- Use table-driven tests where appropriate

Example test structure:
```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "result", false},
        {"invalid input", "", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Something(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Something() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Something() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Project Structure

```
simply_syslog/
├── cmd/
│   └── simplysyslog/      # Main application entry point
│       └── main.go
├── internal/              # Private application code
│   ├── buffer/           # Message buffering logic
│   ├── config/           # Configuration management
│   ├── server/           # UDP/TCP server implementations
│   ├── syslog/           # Syslog parsing logic
│   └── utils/            # Utility functions
├── pkg/                  # Public libraries (can be imported)
│   └── applogger/        # Application logging package
├── config/               # Configuration files
│   ├── config.json       # Default configuration
│   └── regex.json        # Syslog regex patterns
├── .github/
│   └── workflows/        # GitHub Actions CI/CD
├── Dockerfile            # Docker build configuration
└── Taskfile.yml          # Task runner configuration
```

## Areas for Contribution

Check [NEXT_STEPS.md](NEXT_STEPS.md) for a prioritized list of tasks that need attention. Good areas for contribution include:

1. **Unit Tests** - The project needs comprehensive test coverage
2. **Documentation** - Improve existing docs or add new guides
3. **Features** - Implement planned features like database logging
4. **Bug Fixes** - Fix reported issues
5. **Performance** - Optimize critical code paths

## Pull Request Guidelines

### Before Submitting

- [ ] Code follows Go conventions
- [ ] All tests pass
- [ ] Code is formatted with `gofmt`
- [ ] No `go vet` warnings
- [ ] Documentation updated (if applicable)
- [ ] Commit messages are clear and descriptive

### PR Description

Include in your PR description:
- Summary of changes
- Motivation and context
- Testing done
- Screenshots (if UI changes)
- Related issues (use "Fixes #123" to auto-close issues)

### Review Process

1. A maintainer will review your PR
2. Address any feedback or requested changes
3. Once approved, a maintainer will merge your PR
4. Your contribution will be included in the next release

## Reporting Issues

### Bug Reports

When reporting bugs, include:
- Description of the issue
- Steps to reproduce
- Expected behavior
- Actual behavior
- Environment (OS, Go version, Docker version)
- Log output (if applicable)
- Configuration used

### Feature Requests

When requesting features, include:
- Description of the feature
- Use case and motivation
- Proposed implementation (if you have ideas)
- Any alternatives considered

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Give constructive feedback
- Focus on what's best for the project
- Show empathy towards others

## Questions?

If you have questions about contributing, feel free to:
- Open an issue with the "question" label
- Reach out to maintainers
- Check existing issues and documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to simply_syslog! 🎉
