# Repository rules

## Project understanding

- When you need to understand the project, start by reading [INDEX.md](INDEX.md). It links to all relevant documentation. Browse only the files relevant to the task at hand.

- [Taskfile.yml](Taskfile.yml) is used to run most common commands. Check it first before running a command from scratch.

## Software Engineering principles

- **Skateboard before Lamborghini**: If an idea or concept needs to be validated, create first an MVP instead of a full fledged feature, validate the assumptions (e.g. if it can work, if it's useful, etc...) and only then expand it.
- **ABT**: Always Be Testing. Write extremely high-quality tests, use them extensively and follow Test Driven Design.
- **Locality of Behavior**: It's easier to understand parts of code that work together if they are colocated. Try to keep this in mind when writing.
- **YAGNI**: You Aren't Gonna Need It. Use coding best practices but code for the now. There's always time to add stuff later.
- **Modularity**: Build for flexibility and optionality, so it's easy to add, remove or change features.
- **Two-way door decisions**: Make decisions that can be easily reverted (like a two-way door, you can go in and go out).

## Commit Standards

This project uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) enforced by commitizen. Commit messages must follow this format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `fix` - Bug fixes
- `feat` - New user facing features
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `build` - Changes that affect the build system or external dependencies
- `ci` - Changes to our CI configuration files and scripts
- `perf` - A code change that improves performance
- `chore` - Dependencies, etc. Anything that doesn't fit the above types

If a commit introduces a breaking API change (correlating with MAJOR in Semantic Versioning), use `BREAKING CHANGE:` in the footer or append a `!` after the type/scope. A BREAKING CHANGE can be part of commits of any type.

Use `git commit` (pre-commit hooks will validate).
